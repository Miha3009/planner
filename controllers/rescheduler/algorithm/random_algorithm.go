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

package algorithm

import (
    "context"
    "math/rand"

    helper "github.com/miha3009/planner/controllers/helper"
    constraints "github.com/miha3009/planner/controllers/rescheduler/constraints"
    preferences "github.com/miha3009/planner/controllers/rescheduler/preferences"
    types "github.com/miha3009/planner/controllers/types"
)

type RandomAlgorithm struct {
    Attempts    int
    Constraints constraints.ConstraintList
    Preferences preferences.PreferenceList
}

func (a *RandomAlgorithm) Run(ctx context.Context, oldNodes []types.NodeInfo, freePods []types.PodInfo) ([]types.NodeInfo, bool) {
    if helper.ContextEnded(ctx) || len(oldNodes) == 0 {
        return oldNodes, true
    }

    N := len(oldNodes)
    nodes := helper.DeepCopyNodes(oldNodes)
    a.Constraints.Init(nodes)
    a.Preferences.Init(nodes)

    for j := 0; j <= a.Attempts; j++ {
        if helper.ContextEnded(ctx) {
            break
        }
        if len(nodes) > 1 {
            a.TryToAddRandomPod(&nodes[rand.Intn(N)], &freePods)
        }
        a.TryToReschedule(nodes, &freePods)
    }

    return nodes, a.Constraints.CheckForAll(nodes) && len(freePods) == 0
}

func (a *RandomAlgorithm) AddPod(node *types.NodeInfo, pod *types.PodInfo) {
    node.AddPod(*pod)
    a.Constraints.AddPod(node, pod)
    a.Preferences.AddPod(node, pod)
}

func (a *RandomAlgorithm) RemovePod(node *types.NodeInfo, pod *types.PodInfo) {
    node.RemovePod(*pod)
    a.Constraints.RemovePod(node, pod)
    a.Preferences.RemovePod(node, pod)
}

func (a *RandomAlgorithm) TryToReschedule(nodes []types.NodeInfo, freePods *[]types.PodInfo) {
    N := len(nodes)
    oldNode := &nodes[rand.Intn(N)]
    newNode := &nodes[rand.Intn(N)]

    if oldNode == newNode {
        return
    }

    M := len(oldNode.Pods)
    if M == 0 {
        return
    }
    pod := oldNode.Pods[rand.Intn(M)]

    movement := types.MovementInfo{
        Pod:     pod,
        OldNode: helper.DeepCopyNode(oldNode),
        NewNode: helper.DeepCopyNode(newNode),
    }

    if a.Constraints.CheckForMove(movement) || !a.Constraints.Check(oldNode) || !a.Constraints.Check(newNode) {
        oldScore, newScore := a.Preferences.ApplyForMove(movement)
        if newScore > oldScore {
            a.AddPod(newNode, &pod)
            a.RemovePod(oldNode, &pod)
            a.TryToAddRandomPod(oldNode, freePods)
        }
    }
}

func (a *RandomAlgorithm) TryToAddRandomPod(node *types.NodeInfo, pods *[]types.PodInfo) {
    if len(*pods) == 0 {
        return
    }

    podNum := rand.Intn(len(*pods))
    if a.TryToAddPod(node, (*pods)[podNum]) {
        *pods = append((*pods)[:podNum], (*pods)[podNum+1:]...)
    }
}

func (a *RandomAlgorithm) TryToAddPod(node *types.NodeInfo, pod types.PodInfo) bool {
    a.AddPod(node, &pod)
    if !a.Constraints.Check(node) {
        a.RemovePod(node, &pod)
        return false
    }
    return true
}
