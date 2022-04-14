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
    "fmt"
    "log"
    "os"
    "strconv"
    "time"

    appsv1 "github.com/miha3009/planner/api/v1"
    controllers "github.com/miha3009/planner/controllers"
    types "github.com/miha3009/planner/controllers/types"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/resource"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func Run(controller *controllers.PlannerReconciler) {
    RunWithPeriod(controller, time.Millisecond * 50)
}

func RunWithPeriod(controller *controllers.PlannerReconciler, d time.Duration) {
    for {
        controller.Reconcile(context.TODO(), reconcile.Request{})
        planner, _ := controller.Informer.GetPlanner(context.TODO(), nil, reconcile.Request{})
        time.Sleep(d)
        if !planner.Status.Active {
            break
        }
    }
}

func NewController(planner *appsv1.Planner, nodes []corev1.Node, pods [][]corev1.Pod, metrics types.MetricsQueue) *controllers.PlannerReconciler {
    events := make(chan types.Event, 10)
    events <- types.Start

    return &controllers.PlannerReconciler{
        Client:         nil,
        Clientset:      nil,
        Log:            ctrl.Log.WithName("controllers").WithName("Planner"),
        Scheme:         nil,
        MetricsClient:  nil,
        Events:         events,
        Cache:          types.NewCache(),
        MainProcess:    nil,
        MetricsProcess: nil,
        LastStart:      time.Time{},
        Informer:       NewInformer(planner, nodes, pods, metrics),
        Executor:       &TestingExecutor{},
    }
}

type ResourceInfo struct {
    Cpu    int
    Memory int
}

func GenControllerForReschedulingFromFile(fileName, nodePolicy string, constraints appsv1.ConstraintArgsList,
    preferences appsv1.PreferenceArgsList) *controllers.PlannerReconciler {

    planner := &appsv1.Planner{
        Spec: appsv1.PlannerSpec{
            Namespaces:             []string{"default"},
            ResourceUpdateStrategy: "none",
            PlanningInterval:       1000000,
            NodePolicy:             nodePolicy,
            Constraints:            constraints,
            Preferences:            preferences,
        },
    }

    file, err := os.Open(fileName)
    if err != nil {
        log.Fatal(err)
    }

    var nodesCount int
    fmt.Fscanf(file, "%d\n", &nodesCount)
    nodesInfo := make([]ResourceInfo, nodesCount)
    podsInfo := make([][]ResourceInfo, nodesCount)
    for i := 0; i < nodesCount; i++ {
        var podsCount int
        fmt.Fscanf(file, "%d %d %d\n", &nodesInfo[i].Cpu, &nodesInfo[i].Memory, &podsCount)
        podsInfo[i] = make([]ResourceInfo, podsCount)
        for j := 0; j < podsCount; j++ {
            fmt.Fscanf(file, "%d %d\n", &podsInfo[i][j].Cpu, &podsInfo[i][j].Memory)
        }
    }

    nodes := transformToCoreNodes(nodesInfo)
    pods := transformToCorePods(podsInfo)

    return NewController(planner, nodes, pods, types.NewMetricsQueue())
}

func GenControllerForReschedulingFromGenerator(nodePolicy string, constraints appsv1.ConstraintArgsList, preferences appsv1.PreferenceArgsList,
    algoArgs appsv1.AlgorithmArgs, nodesInfo []ResourceInfo, podsInfo [][]ResourceInfo) *controllers.PlannerReconciler {

    planner := &appsv1.Planner{
        Spec: appsv1.PlannerSpec{
            Namespaces:             []string{"default"},
            ResourceUpdateStrategy: "none",
            PlanningInterval:       1000000,
            NodePolicy:             nodePolicy,
            Constraints:            constraints,
            Preferences:            preferences,
            Algorithm:              &algoArgs,
        },
    }

    nodes := transformToCoreNodes(nodesInfo)
    pods := transformToCorePods(podsInfo)

    return NewController(planner, nodes, pods, types.NewMetricsQueue())
}

func transformToCoreNodes(nodesInfo []ResourceInfo) []corev1.Node {
    nodes := make([]corev1.Node, len(nodesInfo))
    for i := range nodesInfo {
        resList := resInfoToResList(nodesInfo[i])
        nodes[i] = corev1.Node{
            ObjectMeta: metav1.ObjectMeta{Name: strconv.Itoa(i)},
            Status:     corev1.NodeStatus{Capacity: resList, Allocatable: resList},
        }
    }
    return nodes
}

func transformToCorePods(podsInfo [][]ResourceInfo) [][]corev1.Pod {
    pods := make([][]corev1.Pod, len(podsInfo))
    c := 0
    for i := range podsInfo {
        pods[i] = make([]corev1.Pod, len(podsInfo[i]))
        for j := range podsInfo[i] {
            pods[i][j] = corev1.Pod{
                ObjectMeta: metav1.ObjectMeta{Name: strconv.Itoa(c)},
            }
            pods[i][j].Spec.Containers = []corev1.Container{{
                Resources: corev1.ResourceRequirements{
                    Requests: resInfoToResList(podsInfo[i][j]),
                }},
            }
            c++
        }
    }
    return pods
}

func resInfoToResList(info ResourceInfo) corev1.ResourceList {
    return corev1.ResourceList{
        corev1.ResourceCPU:    *(resource.NewMilliQuantity(int64(info.Cpu), resource.DecimalSI)),
        corev1.ResourceMemory: *(resource.NewQuantity(int64(info.Memory), resource.DecimalSI)),
    }
}
