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

package informer

import (
	"context"
	"time"
	types "github.com/miha3009/planner/controllers/types"
	appsv1 "github.com/miha3009/planner/api/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/prometheus/common/log"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metrics "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
	helper "github.com/miha3009/planner/controllers/helper"
)

func GetInfo(ctx context.Context, events chan types.Event, cache *types.PlannerCache, clt client.Client, planner appsv1.PlannerSpec) {
	nodes, err := getNodes(clt, ctx)
	if err != nil {
		log.Error(err, ". Failed to get nodes")
		events <- types.PhaseEndedWithError
		return
	}
	
	pods, err := getPods(&planner, nodes, clt, ctx)
	if err != nil {
		log.Error(err, ". Failed to get pods")
		events <- types.PhaseEndedWithError
		return
	}
	
	cache.Nodes = nodes
	cache.Pods = pods

	log.Info("Info collected")
	events <- types.InformingEnded
	return
}

func RunMetircsListener(ctx context.Context, cache *types.PlannerCache, mclt *metricsv.Clientset, planner appsv1.PlannerSpec) {
	period := time.Second * time.Duration(planner.MeticsFetchPeriod)
	for {
		if helper.ContextEnded(ctx) {
			break
		}
		
		nodeMetrics, err := getNodeMetrics(mclt, ctx)
		if err != nil {
			log.Warn(err, ". Failed to get node metrics")
			helper.SleepWithContext(ctx, period)
			continue
		}

		podMetrics, err := getPodMetrics(&planner, mclt, ctx)
		if err != nil {
			log.Warn(err, ". Failed to get pod metrics")
			helper.SleepWithContext(ctx, period)
			continue
		}
				
		p := types.MetricsPackage {
			NodeMetrics: nodeMetrics,
			PodMetrics: podMetrics,
			Timestamp: time.Now(),
		}
		
		cache.Metrics.Push(p)
		cache.Metrics.Shrink()
		
		helper.SleepWithContext(ctx, period)
	}
}

func getNodes(clt client.Client, ctx context.Context) ([]corev1.Node, error) {	
	nodeList := &corev1.NodeList{}
	if err := clt.List(ctx, nodeList); err != nil {
		return nil, err
	}
	
	return nodeList.Items, nil
}

func getPods(planner *appsv1.PlannerSpec, nodes []corev1.Node, clt client.Client, ctx context.Context) ([][]corev1.Pod, error) {	
	pods := make([][]corev1.Pod, len(nodes))
	
	for i := range nodes {
		pods[i] = make([]corev1.Pod, 0)
	}
	
	for _, namespace := range planner.Namespaces {
		for i, node := range nodes {
			podList := &corev1.PodList{}

			opts := []client.ListOption{
				client.MatchingFields{"spec.nodeName": node.Name},
				client.InNamespace(namespace),
			}
			
			if err := clt.List(ctx, podList, opts...); err != nil {
				return nil, err
			}
			for _, pod := range podList.Items {
				pods[i] = append(pods[i], pod)		
			}
		}
    	}
	
	return pods, nil
}

func getNodeMetrics(mclt *metricsv.Clientset, ctx context.Context) (map[string]metrics.NodeMetrics, error) {
	res := make(map[string]metrics.NodeMetrics)
	m, err := mclt.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	
	for i := range m.Items {
		res[m.Items[i].Name] = m.Items[i]
	}
	
	return res, nil
}

func getPodMetrics(planner *appsv1.PlannerSpec, mclt *metricsv.Clientset, ctx context.Context) (map[string]metrics.PodMetrics, error) {
	res := make(map[string]metrics.PodMetrics)

	for _, namespace := range planner.Namespaces {
		if m, err := mclt.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{}); err == nil {
			for j := range m.Items {
				res[m.Items[j].Name] = m.Items[j]
			}
		} else {
			return res, err
		}
	}

	return res, nil
}

