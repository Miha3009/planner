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

package testing

import (
    "context"
    "sort"

    appsv1 "github.com/miha3009/planner/api/v1"
    types "github.com/miha3009/planner/controllers/types"
    corev1 "k8s.io/api/core/v1"
    clientset "k8s.io/client-go/kubernetes"

    "sigs.k8s.io/controller-runtime/pkg/client"
)

type TestingExecutor struct{}

func (exe *TestingExecutor) ExecutePlan(ctx context.Context, events chan types.Event, cache *types.PlannerCache, clt client.Client, cltset *clientset.Clientset, planner appsv1.PlannerSpec) {
    events <- types.ExecutingEnded
    events <- types.Stop
}

func GetPodDistribution(cache *types.PlannerCache) [][]string {
    nodes := cache.Nodes
    pods := cache.Pods

    nodeByName := make(map[string]int)
    for i := range nodes {
        nodeByName[nodes[i].Name] = i
    }

    for _, move := range cache.Plan.Movements {
        movePod(move, nodeByName, pods)
    }

    nodesToDelete := make(map[int]struct{})
    for _, node := range cache.Plan.NodesToDelete {
        nodesToDelete[nodeByName[node.Name]] = struct{}{}
    }

    c := 0
    for i := range pods {
        if _, ok := nodesToDelete[i]; !ok {
            pods[c] = pods[i]
            c++
        }
    }
    pods = pods[:c]

    podNames := make([][]string, len(pods))
    for i := range pods {
        podNames[i] = make([]string, len(pods[i]))
        for j := range pods[i] {
            podNames[i][j] = pods[i][j].Name
        }
        sort.Slice(podNames[i], func(j, k int) bool {
            return podNames[i][j] < podNames[i][k]
        })
    }
    return podNames
}

func movePod(move types.Movement, nodeByName map[string]int, pods [][]corev1.Pod) {
    newNode := nodeByName[move.NewNode.Name]
    pods[newNode] = append(pods[newNode], *move.Pod)

    oldNode := nodeByName[move.OldNode.Name]

    for i := range pods[oldNode] {
        if pods[oldNode][i].Name == move.Pod.Name {
            pods[oldNode] = append(pods[oldNode][:i], pods[oldNode][i+1:]...)
            return
        }
    }
}
