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

package uniform

import (
	types "github.com/miha3009/planner/controllers/types"
	helper "github.com/miha3009/planner/controllers/helper"
)

type Uniform struct {}

func (r Uniform) Apply(nodes []types.NodeInfo) float64 {
	N := len(nodes)
	cpuPercentage := make([]float64, N)
	memoryPercentage := make([]float64, N)

	for i, node := range nodes {
		cpuSum := node.MaxCpu - node.AvalibleCpu + helper.CalcPodsCpu(&node)
		memorySum := node.MaxMemory - node.AvalibleMemory + helper.CalcPodsMemory(&node)
		cpuPercentage[i] = float64(cpuSum) / float64(node.MaxCpu) 
		memoryPercentage[i] = float64(memorySum) / float64(node.MaxMemory) 
	}

	cpuVariance := calcNormalizedVariance(cpuPercentage)
	memoryVariance := calcNormalizedVariance(memoryPercentage)
	
	return 100 - (cpuVariance + memoryVariance) / 2
}

func calcNormalizedVariance(nums []float64) float64 {
	if len(nums) == 0 {
		return float64(0)
	}

	N := len(nums)
	sum := float64(0)
	for _, num := range nums {
		sum += num
	}
	mean := sum / float64(N)
	
	variance := float64(0)
	for _, num := range nums {
		variance += (num - mean) * (num - mean)
	}
	
	maxVariance := float64(0)
	for i := 0; i < N; i++ {
		if sum > 1 {
			maxVariance += (1 - mean) * (1 - mean)				
			sum -= 1
		} else {
			maxVariance += (sum - mean) * (sum - mean)
			sum = 0
		}
	}	
	
	return variance * 100 / maxVariance
}
