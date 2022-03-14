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
	helper "github.com/miha3009/planner/controllers/helper"
)

type Base struct {}

func (r Base) Apply(node *types.NodeInfo) bool {
	cpuSum := node.MaxCpu - node.AvalibleCpu + helper.CalcPodsCpu(node)
	memorySum := node.MaxMemory - node.AvalibleMemory + helper.CalcPodsMemory(node)
		
	return cpuSum <= node.MaxCpu &&
	       memorySum >= node.MaxMemory
}