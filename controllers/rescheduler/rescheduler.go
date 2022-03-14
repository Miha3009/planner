/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rescheduler

import (
	"context"

	appsv1 "github.com/miha3009/planner/api/v1"
	corev1 "k8s.io/api/core/v1"
	"github.com/prometheus/common/log"
	resource "k8s.io/apimachinery/pkg/api/resource"
	constraints "github.com/miha3009/planner/controllers/rescheduler/constraints"
	preferences "github.com/miha3009/planner/controllers/rescheduler/preferences"
	types "github.com/miha3009/planner/controllers/types"
	helper "github.com/miha3009/planner/controllers/helper"
)

type Algorithm interface {
	Run(ctx context.Context, nodes []types.NodeInfo) (updatedNodes []types.NodeInfo)
}

type NodePolicy interface {
	Run(ctx context.Context, algo Algorithm, nodes []types.NodeInfo) (updatedNodes []types.NodeInfo, nodesToCreate []types.NodeInfo, nodesToDelete []types.NodeInfo)
}

func GenPlan(ctx context.Context, events chan types.Event, cache *types.PlannerCache, planner appsv1.PlannerSpec) {
	cst := planner.Constraints
	prf := planner.Preferences
	
	rawNodes := cache.Nodes
	rawPods := cache.Pods
	
	nodes := convertNodes(rawNodes, rawPods)

	cl := constraints.ConvertArgs(&cst)
	pl := preferences.ConvertArgs(&prf)
	
	algo := RandomAlgorithm{1000, cl, pl}
	nodePolicy := KeepNodePolicy{}
	updatedNodes, nodesToCreate, nodesToDelete := nodePolicy.Run(ctx, &algo, nodes)
	if helper.ContextEnded(ctx) {
		return
	}
	
	movementsInfo := calcDiff(nodes, updatedNodes)
	movements := convertMovement(movementsInfo, rawNodes, rawPods)
	plan := types.Plan{
		Movements: movements,
		NodesToCreate: genNodesFromInfo(nodesToCreate),
		NodesToDelete: matchNodes(rawNodes, nodesToDelete),
	}
	
	log.Info("Plan generated")
	cache.Plan = &plan
	events <- types.PlanningEnded
}

func convertNodes(rawNodes []corev1.Node, rawPods [][]corev1.Pod) []types.NodeInfo {
	nodes := make([]types.NodeInfo, len(rawNodes))
	
	for i, node := range rawNodes {
		avalibleCpu := resourceToInt(node.Status.Allocatable.Cpu(), "cpu")
		avalibleMemory := resourceToInt(node.Status.Allocatable.Memory(), "mem")
		maxCpu := resourceToInt(node.Status.Capacity.Cpu(), "cpu")
		maxMemory := resourceToInt(node.Status.Capacity.Memory(), "mem")

		pods := convertPods(rawPods[i])
		
		/*for _, pod := range pods {
			avalibleCpu += pod.Cpu
			avalibleMemory += pod.Memory
		}*/

		nodes[i] = types.NodeInfo {
			node.Name,
			maxCpu,
			maxMemory,
			avalibleCpu,
			avalibleMemory,
			pods,
		}
	}
	
	return nodes
}

func convertPods(rawPods []corev1.Pod) []types.PodInfo {
	pods := make([]types.PodInfo, 0)
	
	for i := range rawPods {
		if !suitableForRescheduling(&rawPods[i]) {
			continue
		}
	
		cpuSum := int64(0)
		memorySum := int64(0)
		for _, container := range rawPods[i].Spec.Containers {
			cpuSum += resourceToInt(container.Resources.Requests.Cpu(), "cpu")
			memorySum += resourceToInt(container.Resources.Requests.Memory(), "mem")
		}
		
		pod := types.PodInfo {rawPods[i].Name, cpuSum, memorySum,}
		pods = append(pods, pod)
	}
	
	return pods
}

func suitableForRescheduling(pod *corev1.Pod) bool {
	return true
}

func resourceToInt(res *resource.Quantity, kind string) int64 {
	if kind == "cpu" {
		return res.MilliValue()
	} else if kind == "mem" {
		return res.MilliValue() / int64(1000)
	} else {
		log.Info("Unknown resource type")
		return int64(0)
	}
}

func calcDiff(oldNodes []types.NodeInfo, newNodes []types.NodeInfo) []types.MovementInfo {
	moves := make([]types.MovementInfo, 0)
	nodesMap := make(map[string]types.NodeInfo)
	podsMap := make(map[string]string)
	
	for i := range oldNodes {
		nodesMap[oldNodes[i].Name] = oldNodes[i]
		for j := range oldNodes[i].Pods {
			podsMap[oldNodes[i].Pods[j].Name] = oldNodes[i].Name
		}
	}
	
	for i := range newNodes {
		for j := range newNodes[i].Pods {
			if podsMap[newNodes[i].Pods[j].Name] != newNodes[i].Name {
				move := types.MovementInfo {
					Pod: newNodes[i].Pods[j],
					OldNode: nodesMap[podsMap[newNodes[i].Pods[j].Name]],
					NewNode: newNodes[i],
				}
				moves = append(moves, move)
			}
		}
	}

	return moves
}

func convertMovement(movesInfo []types.MovementInfo, nodes []corev1.Node, pods [][]corev1.Pod) []types.Movement {
	moves := make([]types.Movement, len(movesInfo))
	if len(movesInfo) == 0 {
		return moves
	}
	
	nodesMap := make(map[string]*corev1.Node)
	podsMap := make(map[string]*corev1.Pod)
	
	for i := range nodes {
		nodesMap[nodes[i].Name] = &nodes[i]
	}

	for i := range pods {
		for j := range pods[i] {
			podsMap[pods[i][j].Name] = &pods[i][j]
		}
	}
	
	for i := range moves {
		move := types.Movement {
			Pod: podsMap[movesInfo[i].Pod.Name],
			OldNode: nodesMap[movesInfo[i].OldNode.Name],
			NewNode: nodesMap[movesInfo[i].NewNode.Name],
		}
		moves[i] = move
	}
	
	return moves
}

func matchNodes(nodes []corev1.Node, matchingNodes []types.NodeInfo) []corev1.Node {
	match := make([]corev1.Node, len(matchingNodes))

	for i := range nodes {
		for j := range matchingNodes {
			if matchingNodes[j].Name == nodes[i].Name {
				match[j] = nodes[i]
				break
			}
		}
	}

	return match
}

func genNodesFromInfo(nodes []types.NodeInfo) []corev1.Node {
	coreNodes := make([]corev1.Node, len(nodes))

	for i := range nodes {
		coreNodes[i] = corev1.Node {}
		coreNodes[i].Name = nodes[i].Name
	}

	return coreNodes
}

