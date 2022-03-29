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

package balanced

import (
    types "github.com/miha3009/planner/controllers/types"
)

type Balanced struct{}

func (r Balanced) Init(node *types.NodeInfo) {
}

func (r Balanced) AddPod(node *types.NodeInfo, pod *types.PodInfo) {
}

func (r Balanced) RemovePod(node *types.NodeInfo, pod *types.PodInfo) {
}

func (r Balanced) Apply(nodes []types.NodeInfo) float64 {
    varSum := float64(0)

    for _, node := range nodes {
        cpuSum := node.MaxCpu - node.AvalibleCpu + node.PodsCpu
        memorySum := node.MaxMemory - node.AvalibleMemory + node.PodsMemory
        cpuPercentage := float64(cpuSum) / float64(node.MaxCpu)
        memoryPercentage := float64(memorySum) / float64(node.MaxMemory)
        varSum += r.calcNormalizedVariance(cpuPercentage, memoryPercentage)
    }

    return 100 - varSum/float64(len(nodes))
}

func (r Balanced) calcNormalizedVariance(a, b float64) float64 {
    m := (a + b) / 2
    return ((a-m)*(a-m) + (b-m)*(b-m)) / float64(0.5)
}
