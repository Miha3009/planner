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

    appsv1 "github.com/miha3009/planner/api/v1"
    "github.com/miha3009/planner/controllers/rescheduler/algorithm"
    ts "github.com/miha3009/planner/testing"
    types "github.com/miha3009/planner/controllers/types"
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


