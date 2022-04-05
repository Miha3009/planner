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
    "context"
    "testing"
    "os"
    "fmt"
    "strconv"

    "github.com/miha3009/planner/controllers/rescheduler/algorithm"
    ts "github.com/miha3009/planner/testing"
    types "github.com/miha3009/planner/controllers/types"
    helper "github.com/miha3009/planner/controllers/helper"
)

func TestQuality(t *testing.T) {
    start := 2
    nodes, rawPods := genCluster(1)
    pods := ts.Schedule(nodes, rawPods, "minFreeSpace")
    nodesInfo := transformNodes(nodes, pods)
    fmt.Println(len(nodes))
    fmt.Println(countEmptyNodes(nodesInfo))
    fmt.Printf("%.3f\n", calcResourceUtilization(nodesInfo))
        
    for k := start; k <= 10; k++ {
        optimizer := algorithm.NewOptimizer(10000, k)
        res := optimizer.Optimize(context.TODO(), nodesInfo)
        fmt.Printf("%d %.3f\n", len(res), calcResourceUtilization(res))
    }
    fmt.Println()
}

func TestGenTimes(t *testing.T) {
    nodes, rawPods := genCluster(1)
    pods := ts.Schedule(nodes, rawPods, "minFreeSpace")
    nodesInfo := transformNodes(nodes, pods)
    
    start := 5
    times := make([][]int, 0)
    for k := start; k <= 50; k++ {
        
        optimizer := algorithm.NewOptimizer(20000, k)
        res := optimizer.Optimize(context.TODO(), nodesInfo)
        times = append(times, getTimes(res))
    }
    timesToFile(start, times)
}


func timesToFile(start int, times [][]int) {
    file, _ := os.OpenFile("times.txt", os.O_RDWR|os.O_CREATE, 0755)
    fmt.Fprintf(file, "%d\n", start)
    for i := range times {
        for j := range times[i] {
            if times[i][j] < 20000 {
                fmt.Fprintf(file, "%d ", times[i][j])
            }
        }
        fmt.Fprintf(file, "\n")
    }
}

func transformNodes(nodes []ts.ResourceInfo, pods [][]ts.ResourceInfo) []types.NodeInfo {
    nodesInfo := make([]types.NodeInfo, len(nodes))
    for i := range nodes {
        nodesInfo[i] = types.NodeInfo{Name: strconv.Itoa(i), MaxCpu: int64(nodes[i].Cpu), MaxMemory: int64(nodes[i].Memory),}
    }
    for i := range pods {
        podsInfo := transformPods(pods[i])
        for j := range podsInfo {
            nodesInfo[i].AddPod(podsInfo[j])
        }
    }

    return nodesInfo
}

func transformPods(pods []ts.ResourceInfo) []types.PodInfo {
    podsInfo := make([]types.PodInfo, len(pods))
    for i := range pods {
        podsInfo[i] = types.PodInfo{Cpu: int64(pods[i].Cpu), Memory: int64(pods[i].Memory),}
    }
    return podsInfo
}

func getTimes(nodes []types.NodeInfo) []int {
    timeStrings := nodes[0].UntoleratedPods
    times := make([]int, len(timeStrings))
    for i := range times {
        times[i], _ = strconv.Atoi(timeStrings[i])
    }
    return times
}

func genCluster(num int) ([]ts.ResourceInfo, []ts.ResourceInfo) {
    switch num {
        case 0:
            return ts.GenCluster(50, 250)
        case 1:
            return ts.GenCluster(250, 1200)
        case 2:
            return ts.GenCluster(2000, 10000)
        default:
            return nil, nil
    }
}

func countEmptyNodes(nodes []types.NodeInfo) int {
    c := 0
    for i := range nodes {
        if len(nodes[i].Pods) == 0 {
            c++
        }
    }
    return c
}

func calcResourceUtilization(nodes []types.NodeInfo) float64 {
    res := make([]float64, 0)
    
    for i := range nodes {
        if len(nodes[i].Pods) != 0 {
            cpu := float64(helper.CalcPodsCpu(&nodes[i])) / float64(nodes[i].MaxCpu)
            memory := float64(helper.CalcPodsMemory(&nodes[i])) / float64(nodes[i].MaxMemory)
            res = append(res, (cpu + memory)/2)
        }
    }
    
    return helper.Mean(res)
}

