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

package fake_cluster

import (
    "testing"
    "time"
    "os"
    "fmt"

    appsv1 "github.com/miha3009/planner/api/v1"
    ts "github.com/miha3009/planner/testing"
)

const testFileName = "cluster.txt"

func TestPlanningTime(t *testing.T) {
    attemps, timeout, nodePerCycle, nodes, rawPods := readClusterFromFile(t, testFileName)
    pods := ts.Schedule(nodes, rawPods, "minFreeSpace")

    controller := ts.GenControllerForReschedulingFromGenerator("shrink",
        appsv1.ConstraintArgsList{},
        appsv1.PreferenceArgsList{Economy: &appsv1.EconomyArgs{Weight: 1}},
        appsv1.AlgorithmArgs{Attemps: attemps, StealPodChance: 10, UseOptimizer: true,
            OptimizerTimeLimitPerCycle: timeout, OptimizerMaxNodesPerCycle: nodePerCycle},
        nodes, pods)
    
    start := time.Now()
    ts.RunWithPeriod(controller, time.Millisecond)
    end := time.Now()
    dur := end.Sub(start).Milliseconds()

    t.Log(fmt.Sprintf("Plan was generated in %d millisecond", dur))
}

func readClusterFromFile(t *testing.T, fileName string) (int, int, int, []ts.ResourceInfo, []ts.ResourceInfo) {
    file, err := os.OpenFile(fileName, os.O_RDONLY, 0755)
    if err != nil {
        t.Fatal(err)
        return 0, 0, 0, nil, nil
    }
    
    var attemps, timeout, nodePerCycle int
    fmt.Fscanf(file, "%d %d %d\n", &attemps, &timeout, &nodePerCycle)
    
    var nodesCount int
    fmt.Fscanf(file, "%d\n", &nodesCount)
    var nodeCpu, nodeMemory int
    fmt.Fscanf(file, "%d %d\n", &nodeCpu, &nodeMemory)
    nodes := make([]ts.ResourceInfo, nodesCount)
    for i := 0; i < nodesCount; i++ {
       nodes[i].Cpu = nodeCpu
       nodes[i].Memory = nodeMemory
    }
    
    var podsCount int
    fmt.Fscanf(file, "%d\n", &podsCount)
    
    pods := make([]ts.ResourceInfo, podsCount)
    for i := 0; i < podsCount; i++ {
        fmt.Fscanf(file, "%d %d\n", &pods[i].Cpu, &pods[i].Memory)
    }
    
    return attemps, timeout, nodePerCycle, nodes, pods
}

