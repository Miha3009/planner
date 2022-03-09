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

package helper

import (
	types "github.com/miha3009/planner/controllers/types"
)

func CalcPodsCpu (node *types.NodeInfo) int64 {
	cpuSum := int64(0)
	
	for _, pod := range node.Pods {
		cpuSum += pod.Cpu
	}
	
	return cpuSum
}

func CalcPodsMemory (node *types.NodeInfo) int64 {
	memorySum := int64(0)
	
	for _, pod := range node.Pods {
		memorySum += pod.Memory
	}
	
	return memorySum
}

func DeepCopyPods(pods []types.PodInfo) []types.PodInfo {
	copyPods := make([]types.PodInfo, len(pods))
	copy(copyPods, pods)
	return copyPods
}

func DeepCopyNodes(nodes []types.NodeInfo) []types.NodeInfo {
	copyNodes := make([]types.NodeInfo, len(nodes))
	copy(copyNodes, nodes)
	for i := range nodes {
		copyNodes[i].Pods = DeepCopyPods(nodes[i].Pods)
	}
	return copyNodes
}
