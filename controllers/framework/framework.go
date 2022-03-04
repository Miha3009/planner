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

package framework

import (
	corev1 "k8s.io/api/core/v1"
	"github.com/prometheus/common/log"
	resource "k8s.io/apimachinery/pkg/api/resource"
)

type Permutation struct {
	PodName string
	OldNodeName string
	NewNodeName string
}

type Plan []Permutation

type PodInfo struct {
	Name string
	Cpu int64
	Memory int64
}

type NodeInfo struct {
	Name string
	MaxCpu int64
	MaxMemory int64
	AvalibleCpu int64
	AvalibleMemory int64
	Pods []PodInfo
}

func GenPlan(rawNodes []corev1.Node, rawPods [][]corev1.Pod) Plan {
	nodes := convertNodes(rawNodes, rawPods)
	
	for _, node := range nodes {
		log.Info(node.Name)
		for _, pod := range node.Pods {
			log.Info(pod.Name)
		}
	}
	
	return Plan{}
}

func convertNodes(rawNodes []corev1.Node, rawPods [][]corev1.Pod) []NodeInfo {
	nodes := make([]NodeInfo, len(rawNodes))
	
	for i, node := range rawNodes {
		capacityCpu := resourceToInt(node.Status.Capacity.Cpu())
		capacityMemory := resourceToInt(node.Status.Capacity.Memory())
		allocatableCpu := resourceToInt(node.Status.Allocatable.Cpu())
		allocatableMemory := resourceToInt(node.Status.Allocatable.Memory())

		nodes[i] = NodeInfo {
			node.Name,
			capacityCpu,
			capacityMemory,
			allocatableCpu,
			allocatableMemory,
			convertPods(rawPods[i]),
		}
	}
	
	return nodes
}

func convertPods(rawPods []corev1.Pod) []PodInfo {
	pods := make([]PodInfo, len(rawPods))
	
	for i := range rawPods {
		cpuSum := int64(0)
		memorySum := int64(0)
		for _, container := range rawPods[i].Spec.Containers {
			cpuSum += resourceToInt(container.Resources.Requests.Cpu())
			memorySum += resourceToInt(container.Resources.Requests.Memory())
		}
		
		pods[i] = PodInfo {rawPods[i].Name, cpuSum, memorySum,}
	}
	
	return pods
}


func resourceToInt(res *resource.Quantity) int64 {
	if num, ok := res.AsInt64(); ok == false {
		log.Info("Failed to convert resource");
		return 0;
	} else {
		return num;
	}
}

func CheckLimits(node NodeInfo) bool {
	cpu := int64(0)
	for _, pod := range node.Pods {
		cpu = cpu + pod.Cpu
	}
	return cpu <= node.MaxCpu
}
