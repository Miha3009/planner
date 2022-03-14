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

package resourceupdater

import (
	"context"
	types "github.com/miha3009/planner/controllers/types"
	appsv1 "github.com/miha3009/planner/api/v1"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	
	"github.com/prometheus/common/log"
)

type ContainerMetrics struct {
	Cpu []int64
	Memory []int64
}

type PodMetrics map[string]ContainerMetrics

func UpdatePodResources(ctx context.Context, events chan types.Event, cache *types.PlannerCache, planner appsv1.PlannerSpec) {
	cache.Metrics.Lock()
	defer cache.Metrics.Unlock()
		
	for i := range cache.Pods {
		for j := range cache.Pods[i] {
			if newPod, needUpdate := updatePod(ctx, &cache.Pods[i][j], cache.Metrics, planner.ResourceUpdateStrategy); needUpdate {
				cache.UpdatedPods = append(cache.UpdatedPods, *newPod)
			}
		}
	}
		
	log.Info("Pods updated")
	events <- types.ResourceUpdatingEnded
}

func updatePod(ctx context.Context, pod *corev1.Pod, q types.MetricsQueue, strategy string) (*corev1.Pod, bool) {
	if strategy == "none" {
		return nil, false
	}
	
	m := getPodMetrics(pod.Name, q)

	newPod := *pod
	for i, container := range pod.Spec.Containers {
		requestCpu := calcRecommendation(m[container.Name].Cpu, strategy)
		requestMemory := calcRecommendation(m[container.Name].Memory, strategy)

		newPod.Spec.Containers[i].Resources.Requests = corev1.ResourceList{
			"cpu": *resource.NewMilliQuantity(requestCpu, resource.Format("DecimalSI")),
			"memory": *resource.NewMilliQuantity(requestMemory, resource.Format("DecimalSI")),
		}
	}
	
	*pod = newPod
	
	return &newPod, true
}

func getPodMetrics(podName string, q types.MetricsQueue) PodMetrics {
	podMetrics := PodMetrics{}

	N := q.Size()
	for i := 0; i < N; i++ {
		p := q.Get(i).PodMetrics[podName]
		for j := range p.Containers {
			containerName := p.Containers[j].Name
			cpu := p.Containers[j].Usage.Cpu().MilliValue()
			memory := p.Containers[j].Usage.Memory().MilliValue()
			if _, ok := podMetrics[containerName]; ok {
				podMetrics[containerName] = ContainerMetrics{
					Cpu: append(podMetrics[containerName].Cpu, cpu),
					Memory: append(podMetrics[containerName].Memory, memory),
				}
			} else {
				podMetrics[containerName] = ContainerMetrics{
					Cpu: []int64{cpu},
					Memory: []int64{memory},
				}
			}
		}
	}
	
	return podMetrics
}

func calcRecommendation(x []int64, strategy string) int64 {
	switch strategy {
		case "max":
			return calcMaxStrategy(x)
		default:
			return int64(0)
	}
}

func calcMaxStrategy(x []int64) int64 {
	maxValue := int64(0)
	for i := range x {
		if x[i] > maxValue {
			maxValue = x[i]
		}
	}
	
	return maxValue	
}

