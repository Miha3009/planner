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
    "fmt"
    "os"
    "testing"

    appsv1 "github.com/miha3009/planner/api/v1"
    "github.com/miha3009/planner/controllers/rescheduler/algorithm"
    ts "github.com/miha3009/planner/testing"
)

func TestCase1(t *testing.T) {
    runPolicyCase(t, "case1.txt", []int{0, 1, 1})
}

func TestCase2(t *testing.T) {
    runPolicyCase(t, "case2.txt", []int{0, -1, 0})
}

func TestCase3(t *testing.T) {
    runPolicyCase(t, "case3.txt", []int{0, 0, 0})
}

func TestCase4Uniform(t *testing.T) {
    controller := ts.GenControllerForReschedulingFromFile("case4.txt", "keep",
        appsv1.ConstraintArgsList{},
        appsv1.PreferenceArgsList{Uniform: &appsv1.UniformArgs{Weight: 1}})
    ts.Run(controller)
    d := ts.GetPodDistribution(controller.Cache)
    if !matchStrings(d, [][]string{{"0", "3"}, {"1", "2"}}) {
        t.Fail()
    }
}

func TestCase4MaximizeInequallity(t *testing.T) {
    controller := ts.GenControllerForReschedulingFromFile("case4.txt", "shrink",
        appsv1.ConstraintArgsList{},
        appsv1.PreferenceArgsList{MaximizeInequality: &appsv1.MaximizeInequalityArgs{Weight: 1}})
    ts.Run(controller)
    d := ts.GetPodDistribution(controller.Cache)
    if !matchStrings(d, [][]string{{"0", "1", "2", "3"}}) {
        t.Fail()
    }
}

func TestCase5(t *testing.T) {
    controller := ts.GenControllerForReschedulingFromFile("case5.txt", "keep",
        appsv1.ConstraintArgsList{},
        appsv1.PreferenceArgsList{Balanced: &appsv1.BalancedArgs{Weight: 100}, Uniform: &appsv1.UniformArgs{Weight: 1}})
    ts.Run(controller)
    d := ts.GetPodDistribution(controller.Cache)
    t.Log(d)
    if !matchStrings(d, [][]string{{"0", "1"}, {"2", "3"}}) && !matchStrings(d, [][]string{{"2", "3"}, {"0", "1"}}) {
        t.Fail()
    }
}

func TestCase6(t *testing.T) {
    algorithm.TEST_RUN()
    return
    nodes, rawPods := ts.GenCluster(100, 300)
    pods := ts.Schedule(nodes, rawPods, "minFreeSpace")
    toFile(nodes, pods)
    c := 0
    for i := range pods {
        if len(pods[i]) == 0 {
            c++
        }
    }
    t.Log(c)
    return
    controller := ts.GenControllerForReschedulingFromGenerator("keep",
        appsv1.ConstraintArgsList{},
        //        appsv1.PreferenceArgsList{Uniform: &appsv1.UniformArgs{Weight: 1}},
        appsv1.PreferenceArgsList{Uniform: &appsv1.UniformArgs{Weight: 1}}, //MaximizeInequality: &appsv1.MaximizeInequalityArgs{Weight: 1}},
        appsv1.AlgorithmArgs{Attemps: 100}, nodes, pods)
    ts.Run(controller)
    d := ts.GetPodDistribution(controller.Cache)
    c = 0
    for i := range d {
        if len(d[i]) == 0 {
            c++
        }
    }
    t.Log(len(controller.Cache.Plan.Movements))
    t.Log(c)
}

func toFile(nodes []ts.ResourceInfo, pods [][]ts.ResourceInfo) {
    file, _ := os.OpenFile("D:/work/myscheduler/gen.txt", os.O_RDWR|os.O_CREATE, 0755)
    fmt.Fprintf(file, "%d\n", len(nodes))
    for i := range nodes {
        fmt.Fprintf(file, "%d %d %d\n", nodes[i].Cpu, nodes[i].Memory, len(pods[i]))
        for j := range pods[i] {
            fmt.Fprintf(file, "%d %d\n", pods[i][j].Cpu, pods[i][j].Memory)
        }
    }
}

func runPolicyCase(t *testing.T, fileName string, expectedNodeChange []int) {
    policies := []string{"keep", "shrink", "only_grow"}
    for i := range policies {
        controller := ts.GenControllerForReschedulingFromFile(fileName, policies[i],
            appsv1.ConstraintArgsList{},
            appsv1.PreferenceArgsList{Uniform: &appsv1.UniformArgs{Weight: 1}})
        ts.Run(controller)
        plan := controller.Cache.Plan
        nodeChange := len(plan.NodesToCreate) - len(plan.NodesToDelete)
        if expectedNodeChange[i] != nodeChange {
            t.Fail()
        }
    }
}

func matchStrings(a, b [][]string) bool {
    if len(a) != len(b) {
        return false
    }
    for i := range a {
        if len(a[i]) != len(b[i]) {
            return false
        }
        for j := range a[i] {
            if a[i][j] != b[i][j] {
                return false
            }
        }
    }
    return true
}
