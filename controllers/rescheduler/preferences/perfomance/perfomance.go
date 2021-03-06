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

package perfomance

import (
    helper "github.com/miha3009/planner/controllers/helper"
    types "github.com/miha3009/planner/controllers/types"
)

type Perfomance struct{}

func (r Perfomance) Init(node *types.NodeInfo) {
}

func (r Perfomance) AddPod(node *types.NodeInfo, pod *types.PodInfo) {
}

func (r Perfomance) RemovePod(node *types.NodeInfo, pod *types.PodInfo) {
}

func (r Perfomance) Apply(nodes []types.NodeInfo) float64 {
    N := len(nodes)
    cpuPercentage := make([]float64, N)
    memoryPercentage := make([]float64, N)

    for i, node := range nodes {
        cpuSum := node.MaxCpu - node.AvalibleCpu + node.PodsCpu
        memorySum := node.MaxMemory - node.AvalibleMemory + node.PodsMemory
        cpuPercentage[i] = float64(cpuSum) / float64(node.MaxCpu)
        memoryPercentage[i] = float64(memorySum) / float64(node.MaxMemory)
    }

    cpuVariance := calcNormalizedVariance(cpuPercentage)
    memoryVariance := calcNormalizedVariance(memoryPercentage)

    return (cpuVariance + memoryVariance) / 2
}

func calcNormalizedVariance(nums []float64) float64 {
    if len(nums) == 0 {
        return float64(0)
    }

    maxVariance := float64(0.25)
    return helper.Variance(nums) * types.MaxPreferenceScore / maxVariance
}
