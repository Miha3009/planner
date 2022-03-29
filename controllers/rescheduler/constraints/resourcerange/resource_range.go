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

package resourcerange

import (
    appsv1 "github.com/miha3009/planner/api/v1"
    types "github.com/miha3009/planner/controllers/types"
)

type ResourceRange struct {
    Args appsv1.ResourceRangeArgs
}

func Validate(args *appsv1.ResourceRangeArgs) bool {
    return args.MinCpu <= args.MaxCpu && args.MinMemory <= args.MaxMemory
}

func (r ResourceRange) Init(node *types.NodeInfo) {
}

func (r ResourceRange) AddPod(node *types.NodeInfo, pod *types.PodInfo) {
}

func (r ResourceRange) RemovePod(node *types.NodeInfo, pod *types.PodInfo) {
}

func (r ResourceRange) Check(node *types.NodeInfo) bool {
    cpuSum := node.MaxCpu - node.AvalibleCpu + node.PodsCpu
    memorySum := node.MaxMemory - node.AvalibleMemory + node.PodsMemory

    return cpuSum >= node.MaxCpu*r.Args.MinCpu/100 &&
        memorySum >= node.MaxMemory*r.Args.MinMemory/100 &&
        cpuSum <= node.MaxCpu*r.Args.MaxCpu/100 &&
        memorySum <= node.MaxMemory*r.Args.MaxMemory/100
}
