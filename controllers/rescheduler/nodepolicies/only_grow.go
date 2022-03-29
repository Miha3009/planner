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
    "strconv"

    "github.com/miha3009/planner/controllers/helper"
    algorithm "github.com/miha3009/planner/controllers/rescheduler/algorithm"
    types "github.com/miha3009/planner/controllers/types"
)

type OnlyGrowNodePolicy struct{}

func (a *OnlyGrowNodePolicy) Run(ctx context.Context, algo algorithm.Algorithm, nodes []types.NodeInfo) ([]types.NodeInfo, []types.NodeInfo, []types.NodeInfo) {
    if len(nodes) == 0 {
        return nodes, nil, nil
    }

    newNodes, ok := algo.Run(ctx, nodes, []types.PodInfo{})
    if !ok {
        return grow(ctx, algo, newNodes)
    }

    return newNodes, nil, nil
}

func grow(ctx context.Context, algo algorithm.Algorithm, nodes []types.NodeInfo) ([]types.NodeInfo, []types.NodeInfo, []types.NodeInfo) {
    nodesToCreate := make([]types.NodeInfo, 0)
    for {
        if helper.ContextEnded(ctx) {
            return nodes, nodesToCreate, nil
        }

        newNode := genNewNode(nodes, len(nodesToCreate))
        nodesToCreate = append(nodesToCreate, newNode)
        nodes = append(nodes, newNode)
        newNodes, ok := algo.Run(ctx, nodes, []types.PodInfo{})
        if ok {
            return newNodes, nodesToCreate, nil
        } else {
            nodes = newNodes
        }
    }
}

func genNewNode(nodes []types.NodeInfo, num int) types.NodeInfo {
    return types.NodeInfo{
        Name:           strconv.Itoa(num),
        MaxCpu:         nodes[0].MaxCpu,
        MaxMemory:      nodes[0].MaxMemory,
        AvalibleCpu:    nodes[0].MaxCpu,
        AvalibleMemory: nodes[0].MaxMemory,
        Pods:           []types.PodInfo{},
    }
}
