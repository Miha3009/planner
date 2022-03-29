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

package executor

import (
    "context"
    "sort"
    "time"

    appsv1 "github.com/miha3009/planner/api/v1"
    types "github.com/miha3009/planner/controllers/types"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    clientset "k8s.io/client-go/kubernetes"

    "sigs.k8s.io/controller-runtime/pkg/client"

    "github.com/prometheus/common/log"
)

type NodeDriver interface {
    AddNode() bool
    DeleteNode(node *corev1.Node) bool
}

type Executor interface {
    ExecutePlan(ctx context.Context, events chan types.Event, cache *types.PlannerCache, clt client.Client, cltset *clientset.Clientset, planner appsv1.PlannerSpec)
}

type DefaultExecutor struct{}

func (exe *DefaultExecutor) ExecutePlan(ctx context.Context, events chan types.Event, cache *types.PlannerCache, clt client.Client, cltset *clientset.Clientset, planner appsv1.PlannerSpec) {
    plan := cache.Plan
    movements := unite(cache.Nodes, plan.Movements, cache.UpdatedPods)
    movements = prioritizeMovements(movements)

    for _, move := range movements {
        movePod(ctx, cltset, move)
    }

    events <- types.ExecutingEnded
}

func movePod(ctx context.Context, cltset *clientset.Clientset, move types.Movement) bool {
    newPod := &corev1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Namespace: move.Pod.Namespace,
            Name:      NewNameForPod(move.Pod),
        },
        Spec: move.Pod.Spec,
    }

    err := createPod(ctx, cltset, newPod, move.NewNode)
    if err != nil {
        log.Info(err)
        return false
    }

    for !isPodRunning(ctx, cltset, newPod) {
        time.Sleep(time.Second)
    }

    deletePod(ctx, cltset, move.Pod)

    newPod.Labels = move.Pod.Labels
    _, err = cltset.CoreV1().Pods("default").Update(ctx, newPod, metav1.UpdateOptions{})
    if err != nil {
        log.Info(err)
    }

    return true
}

func isPodRunning(ctx context.Context, cltset *clientset.Clientset, pod *corev1.Pod) bool {
    pod, err := cltset.CoreV1().Pods(pod.Namespace).Get(ctx, pod.Name, metav1.GetOptions{})
    if err != nil {
        log.Info(err)
    }
    return pod.Status.Phase == corev1.PodRunning
}

func deletePod(ctx context.Context, cltset *clientset.Clientset, pod *corev1.Pod) {
    err := cltset.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
    if err != nil {
        log.Info(err)
    }
}

func createPod(ctx context.Context, cltset *clientset.Clientset, pod *corev1.Pod, node *corev1.Node) error {
    pod.Spec.NodeName = node.Name
    _, err := cltset.CoreV1().Pods(pod.Namespace).Create(ctx, pod, metav1.CreateOptions{})
    return err
}

func unite(nodes []corev1.Node, moves []types.Movement, updatedPods []corev1.Pod) []types.Movement {
    nodeByName := make(map[string]int)
    for i := range nodes {
        nodeByName[nodes[i].Name] = i
    }

    moveByPodName := make(map[string]int)
    for i := range moves {
        moveByPodName[moves[i].Pod.Name] = i
    }

    for i := range updatedPods {
        podName := updatedPods[i].Name
        if _, ok := moveByPodName[podName]; !ok {
            nodeName := updatedPods[i].Spec.NodeName
            moves = append(moves, types.Movement{
                Pod:     &updatedPods[i],
                OldNode: &nodes[nodeByName[nodeName]],
                NewNode: &nodes[nodeByName[nodeName]],
            })
        }
    }

    return moves
}

func prioritizeMovements(moves []types.Movement) []types.Movement {
    sort.Slice(moves, func(i, j int) bool {
        return *moves[i].Pod.Spec.Priority > *moves[j].Pod.Spec.Priority
    })
    return moves
}
