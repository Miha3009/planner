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

    appsv1 "github.com/miha3009/planner/api/v1"
    types "github.com/miha3009/planner/controllers/types"

    corev1 "k8s.io/api/core/v1"
    metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

type TestingInformer struct {
    planner *appsv1.Planner
    nodes   []corev1.Node
    pods    [][]corev1.Pod
    metrics types.MetricsQueue
}

func NewInformer(planner *appsv1.Planner, nodes []corev1.Node, pods [][]corev1.Pod, metrics types.MetricsQueue) *TestingInformer {
    return &TestingInformer{
        planner: planner,
        nodes:   nodes,
        pods:    pods,
        metrics: metrics,
    }
}

func (inf *TestingInformer) GetInfo(ctx context.Context, events chan types.Event, cache *types.PlannerCache, clt client.Client, planner appsv1.PlannerSpec) {
    cache.Nodes = inf.nodes
    cache.Pods = inf.pods
    events <- types.InformingEnded
}

func (inf *TestingInformer) RunMetircsListener(ctx context.Context, cache *types.PlannerCache, mclt *metricsv.Clientset, planner appsv1.PlannerSpec) {
    cache.Metrics = inf.metrics
}

func (inf *TestingInformer) GetPlanner(ctx context.Context, clt client.Client, req ctrl.Request) (*appsv1.Planner, error) {
    return inf.planner, nil
}

func (inf *TestingInformer) UpdatePlanner(ctx context.Context, clt client.Client, planner *appsv1.Planner) {
    inf.planner = planner
}
