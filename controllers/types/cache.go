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

package types

import (
	corev1 "k8s.io/api/core/v1"
)

type PlannerCache struct {
	Nodes []corev1.Node
	Pods [][]corev1.Pod
	Metrics MetricsQueue
	Plan *Plan
}

func NewCache() *PlannerCache {
	return &PlannerCache{
		Nodes: make([]corev1.Node, 0),
		Pods: make([][]corev1.Pod, 0),
		Metrics: NewMetricsQueue(),
		Plan: nil,
	}
}

func (cache *PlannerCache) Clear() {
	cache.Nodes = make([]corev1.Node, 0)
	cache.Pods = make([][]corev1.Pod, 0)
	cache.Plan = nil
}

