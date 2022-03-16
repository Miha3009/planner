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
	types "github.com/miha3009/planner/controllers/types"
	appsv1 "github.com/miha3009/planner/api/v1"

	corev1 "k8s.io/api/core/v1"
	"github.com/prometheus/common/log"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

