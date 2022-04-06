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
    "math/rand"

    "github.com/miha3009/planner/controllers/helper"
    algorithm "github.com/miha3009/planner/controllers/rescheduler/algorithm"
    types "github.com/miha3009/planner/controllers/types"
)

type ShrinkNodePolicy struct{
    Optimizer *algorithm.Optimizer
    MaxNodes int
}

func (a *ShrinkNodePolicy) Run(ctx context.Context, algo algorithm.Algorithm, nodes []types.NodeInfo) ([]types.NodeInfo, []types.NodeInfo, []types.NodeInfo) {
    if len(nodes) == 0 {
        return nodes, nil, nil
    }

    newNodes, ok := algo.Run(ctx, nodes, []types.PodInfo{})
    if ok {
        nodes = newNodes
        if a.Optimizer != nil {
            newNodes = a.Optimizer.Optimize(ctx, nodes)
            if len(nodes) == len(newNodes) {
                return nodes, nil, nil
            } else {
                nodesToDelete := getDiffNodes(nodes, newNodes)
                return newNodes, nil, nodesToDelete
            }
        } else {
            nodesToDelete := make([]types.NodeInfo, 0)
            for {
                if helper.ContextEnded(ctx) || len(nodes) == 1 {
                    return nodes, nil, nodesToDelete
                }

                nodeI := choseNodeForDelete(nodes)
                newNodes = helper.DeepCopyNodes(nodes)
                newNodes = append(newNodes[:nodeI], newNodes[nodeI+1:]...)
                newNodes, ok = algo.Run(ctx, newNodes, helper.DeepCopyPods(nodes[nodeI].Pods))
                if ok {
                    nodesToDelete = append(nodesToDelete, nodes[nodeI])
                    nodes = newNodes
                } else {
                    return nodes, nil, nodesToDelete
                }
            }
        }
    } else {
        return grow(ctx, algo, newNodes, a.MaxNodes)
    }
}

func choseNodeForDelete(nodes []types.NodeInfo) int {
    for i := range nodes {
        if len(nodes[i].Pods) == 0 {
            return i
        }
    }

    return rand.Intn(len(nodes))
}

func getDiffNodes(oldNodes []types.NodeInfo, newNodes []types.NodeInfo) []types.NodeInfo {
    nodesByName := make(map[string]struct{})
    for i := range newNodes {
        nodesByName[newNodes[i].Name] = struct{}{}
    }

    diffNodes := make([]types.NodeInfo, 0)    
    for i := range oldNodes {
        if _, ok := nodesByName[oldNodes[i].Name]; !ok {
            diffNodes = append(diffNodes, oldNodes[i])
        }
    }
    return diffNodes
}

