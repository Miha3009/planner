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

package algorithm

import (
    "context"
    "math"
    "math/rand"

    helper "github.com/miha3009/planner/controllers/helper"
    constraints "github.com/miha3009/planner/controllers/rescheduler/constraints"
    preferences "github.com/miha3009/planner/controllers/rescheduler/preferences"
    types "github.com/miha3009/planner/controllers/types"
    "github.com/prometheus/common/log"
)

type ImprovedAlgorithm struct {
    Attempts    int
    Constraints constraints.ConstraintList
    Preferences preferences.PreferenceList
}

func (a *ImprovedAlgorithm) Run(ctx context.Context, oldNodes []types.NodeInfo, freePods []types.PodInfo) ([]types.NodeInfo, bool) {
    if helper.ContextEnded(ctx) || len(oldNodes) == 0 {
        return oldNodes, true
    }

    nodes := helper.DeepCopyNodes(oldNodes)
    a.Constraints.Init(nodes)
    a.Preferences.Init(nodes)

    for j := 0; j <= a.Attempts; j++ {
        if helper.ContextEnded(ctx) {
            break
        }
        newNodes := helper.DeepCopyNodes(oldNodes)
        removedPods := a.GetRandomPods(newNodes)
        pods := append(helper.DeepCopyPods(freePods), removedPods...)
        a.MaxFit(newNodes, &pods)

        log.Info(a.Preferences.Apply(newNodes), " ", a.Preferences.Apply(nodes))
        if a.Preferences.Apply(newNodes) > a.Preferences.Apply(nodes) {
            nodes = newNodes
            freePods = make([]types.PodInfo, 0)
        }
    }

    return nodes, a.Constraints.CheckForAll(nodes) && len(freePods) == 0
}

func (a *ImprovedAlgorithm) AddPod(node *types.NodeInfo, pod *types.PodInfo) {
    node.AddPod(*pod)
    a.Constraints.AddPod(node, pod)
    a.Preferences.AddPod(node, pod)
}

func (a *ImprovedAlgorithm) RemovePod(node *types.NodeInfo, pod *types.PodInfo) {
    node.RemovePod(*pod)
    a.Constraints.RemovePod(node, pod)
    a.Preferences.RemovePod(node, pod)
}

func (a *ImprovedAlgorithm) TryToAddPod(node *types.NodeInfo, pod types.PodInfo) bool {
    a.AddPod(node, &pod)
    if !a.Constraints.Check(node) {
        a.RemovePod(node, &pod)
        return false
    }
    return true
}

func (a *ImprovedAlgorithm) GetRandomPods(nodes []types.NodeInfo) []types.PodInfo {
    pods := make([]types.PodInfo, 0)
    for i := 0; i < 3000; i++ {
        node := &nodes[rand.Intn(len(nodes))]
        if len(node.Pods) == 0 {
            continue
        }
        podI := rand.Intn(len(node.Pods))
        pods = append(pods, node.Pods[podI])
        a.RemovePod(node, &node.Pods[podI])
    }
    return pods
}

func (a *ImprovedAlgorithm) MaxFit(nodes []types.NodeInfo, pods *[]types.PodInfo) {
    for j := range *pods {
        minFreeSpace := float64(-2)
        nodeI := -1
        for i := range nodes {
            if a.TryToAddPod(&nodes[i], (*pods)[j]) {
                freeSpace := a.GetFreeSpace(&nodes[i])
                //log.Info(freeSpace)
                if freeSpace > minFreeSpace {
                    minFreeSpace = freeSpace
                    nodeI = i
                }
                a.RemovePod(&nodes[i], &(*pods)[j])
            }
        }
        if nodeI != -1 {
            //log.Info(minFreeSpace)
            a.AddPod(&nodes[nodeI], &(*pods)[j])
        }
    }

    /*pods = &[]types.PodInfo{}

    minFreeSpace := float64(2)
    nodeI := -1
    podI := -1
    for i := range nodes {
        for j := range *pods {
            if a.TryToAddPod(&nodes[i], (*pods)[j]) {
                freeSpace := a.GetFreeSpace(&nodes[i])
                if freeSpace < minFreeSpace {
                    minFreeSpace = freeSpace
                    nodeI = i
                    podI = j
                }
                a.RemovePod(&nodes[i], &(*pods)[j])
            }
        }
    }

    if podI == -1 {
        return
    }

    nodes[nodeI].AddPod((*pods)[podI])
    *pods = append((*pods)[:podI], (*pods)[podI+1:]...)
    if len(*pods) > 0 {
        a.MaxFit(nodes, pods)
    }*/
}

func (a *ImprovedAlgorithm) GetFreeSpace(node *types.NodeInfo) float64 {
    x := float64(node.MaxCpu-node.PodsCpu) / float64(node.MaxCpu)
    y := float64(node.MaxMemory-node.PodsMemory) / float64(node.MaxMemory)
    return math.Max(x, y)
}
