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

package podscount

import (
    appsv1 "github.com/miha3009/planner/api/v1"
    types "github.com/miha3009/planner/controllers/types"
)

type PodsCount struct {
    Args appsv1.PodsCountArgs
}

func (r PodsCount) Init(node *types.NodeInfo) {
}

func (r PodsCount) AddPod(node *types.NodeInfo, pod *types.PodInfo) {
}

func (r PodsCount) RemovePod(node *types.NodeInfo, pod *types.PodInfo) {
}

func (r PodsCount) Check(node *types.NodeInfo) bool {
    return len(node.Pods) <= r.Args.MaxCount
}
