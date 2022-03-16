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
	helper "github.com/miha3009/planner/controllers/helper"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/prometheus/common/log"
	metrics "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

func RunMetircsListener(ctx context.Context, cache *types.PlannerCache, mclt *metricsv.Clientset, planner appsv1.PlannerSpec) {
	if planner.MeticsFetchPeriod == 0 {
		return
	}

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

