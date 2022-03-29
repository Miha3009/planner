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

package nodepolicies

import (
    "context"

    algorithm "github.com/miha3009/planner/controllers/rescheduler/algorithm"
    types "github.com/miha3009/planner/controllers/types"
)

type NodePolicy interface {
    Run(ctx context.Context, algo algorithm.Algorithm, nodes []types.NodeInfo) (updatedNodes []types.NodeInfo, nodesToCreate []types.NodeInfo, nodesToDelete []types.NodeInfo)
}
