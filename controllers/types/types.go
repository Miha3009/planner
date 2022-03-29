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
    "time"

    corev1 "k8s.io/api/core/v1"
    metrics "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type Event int

const (
    Start                 Event = 0
    Stop                        = 1
    InformingEnded              = 2
    ResourceUpdatingEnded       = 3
    PlanningEnded               = 4
    ExecutingEnded              = 5
    PhaseEndedWithError         = 6
)

const MaxPreferenceScore = float64(100)

type PodInfo struct {
    Pod    *corev1.Pod
    Name   string
    Cpu    int64
    Memory int64
}

type NodeInfo struct {
    Node           *corev1.Node
    Name           string
    MaxCpu         int64
    MaxMemory      int64
    AvalibleCpu    int64
    AvalibleMemory int64

    Pods            []PodInfo
    PodsCpu         int64
    PodsMemory      int64
    Ports           map[string]map[int32]struct{}
    PortsConflict   map[string]map[int32]int
    UntoleratedPods []string
}

func (n *NodeInfo) AddPod(p PodInfo) {
    n.Pods = append(n.Pods, p)
    n.PodsCpu += p.Cpu
    n.PodsMemory += p.Memory
}

func (n *NodeInfo) RemovePod(p PodInfo) {
    podNum := -1
    for i, pod := range n.Pods {
        if pod.Name == p.Name {
            podNum = i
            break
        }
    }

    if podNum == -1 {
        return
    }

    n.Pods = append(n.Pods[:podNum], n.Pods[podNum+1:]...)
    n.PodsCpu -= p.Cpu
    n.PodsMemory -= p.Memory
}

type MovementInfo struct {
    Pod     PodInfo
    OldNode NodeInfo
    NewNode NodeInfo
}

type Movement struct {
    Pod     *corev1.Pod
    OldNode *corev1.Node
    NewNode *corev1.Node
}

type Plan struct {
    Movements     []Movement
    NodesToDelete []corev1.Node
    NodesToCreate []corev1.Node
}

type MetricsPackage struct {
    NodeMetrics map[string]metrics.NodeMetrics
    PodMetrics  map[string]metrics.PodMetrics
    Timestamp   time.Time
}
