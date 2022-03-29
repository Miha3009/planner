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

package tainttoleration

import (
    types "github.com/miha3009/planner/controllers/types"
    corev1 "k8s.io/api/core/v1"
)

type TaintToleration struct{}

func (r TaintToleration) Init(node *types.NodeInfo) {
    node.UntoleratedPods = make([]string, 0)
    for i := range node.Pods {
        r.AddPod(node, &node.Pods[i])
    }
}

func (r TaintToleration) AddPod(node *types.NodeInfo, pod *types.PodInfo) {
    tolerations := pod.Pod.Spec.Tolerations
    var taints []corev1.Taint
    if node.Node == nil {
        taints = []corev1.Taint{}
    } else {
        taints = node.Node.Spec.Taints
    }
    taints = r.filterTaints(taints)

    if !r.tolerationsTolerateTaints(taints, tolerations) {
        node.UntoleratedPods = append(node.UntoleratedPods, pod.Name)
    }
}

func (r TaintToleration) RemovePod(node *types.NodeInfo, pod *types.PodInfo) {
    idx := -1
    for i, s := range node.UntoleratedPods {
        if pod.Name == s {
            idx = i
            break
        }
    }

    if idx != -1 {
        node.UntoleratedPods = append(node.UntoleratedPods[:idx], node.UntoleratedPods[idx+1:]...)
    }
}

func (r TaintToleration) Check(node *types.NodeInfo) bool {
    return len(node.UntoleratedPods) == 0
}

func (r TaintToleration) filterTaints(taints []corev1.Taint) []corev1.Taint {
    filteredTaints := make([]corev1.Taint, 0)
    for _, taint := range taints {
        if taint.Effect == corev1.TaintEffectNoSchedule || taint.Effect == corev1.TaintEffectNoExecute {
            filteredTaints = append(filteredTaints, taint)
        }
    }
    return filteredTaints
}

func (r TaintToleration) tolerationsTolerateTaints(taints []corev1.Taint, tolerations []corev1.Toleration) bool {
    for i := range taints {
        tolerate := false
        for j := range tolerations {
            if tolerations[j].ToleratesTaint(&taints[i]) {
                tolerate = true
                break
            }
        }
        if !tolerate {
            return false
        }
    }
    return true
}
