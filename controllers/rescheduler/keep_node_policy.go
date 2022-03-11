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

package rescheduler

import (
	"context"
	types "github.com/miha3009/planner/controllers/types"
)

type KeepNodePolicy struct {}

func (a *KeepNodePolicy) Run(ctx context.Context, algo Algorithm, nodes []types.NodeInfo) ([]types.NodeInfo, []types.NodeInfo, []types.NodeInfo) {
	return algo.Run(ctx, nodes), []types.NodeInfo{}, []types.NodeInfo{}
}

