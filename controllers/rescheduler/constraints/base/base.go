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

package base

import (
	types "github.com/miha3009/planner/controllers/types"
)

type Base struct {}

func (r Base) Init(node *types.NodeInfo) {
	node.PodsCpu = int64(0)
	node.PodsMemory = int64(0)
	for i := range node.Pods {
		r.AddPod(node, &node.Pods[i])
	}
}

func (r Base) AddPod(node *types.NodeInfo, pod *types.PodInfo) {
	node.PodsCpu += pod.Cpu
	node.PodsMemory += pod.Memory
}

func (r Base) RemovePod(node *types.NodeInfo, pod *types.PodInfo) {
	node.PodsCpu -= pod.Cpu
	node.PodsMemory -= pod.Memory	
}

func (r Base) Check(node *types.NodeInfo) bool {
	return node.PodsCpu <= node.AvalibleCpu &&
	       node.PodsMemory <= node.AvalibleMemory
}
