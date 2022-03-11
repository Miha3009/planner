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

package resourcerange

import (
	"testing"

	rr "github.com/miha3009/planner/controllers/framework/constraints/resourcerange"	
	framework "github.com/miha3009/planner/controllers/framework"	
)

func TestGood (t *testing.T) {
	node := framework.NodeInfo {
		"name", 500, 200, 100, 100, 
		[]framework.PodInfo {{"name", 50, 10}},
	}
	
	constraint := rr.ResourceRange{
		MinCpu: 10,
		MinMemory: 10,
		MaxCpu: 90,
		MaxMemory: 90,
	}
	
	if !constraint.Apply(&node) {
		t.FailNow()
	}
}

func TestBad1 (t *testing.T) {
	node := framework.NodeInfo {
		"name", 500, 200, 0, 0, 
		[]framework.PodInfo {},
	}
	
	constraint := rr.ResourceRange{
		MinCpu: 10,
		MinMemory: 10,
		MaxCpu: 90,
		MaxMemory: 90,  
	}
	
	if constraint.Apply(&node) {
		t.FailNow()
	}
}

func TestBad2 (t *testing.T) {
	node := framework.NodeInfo {
		"name", 500, 200, 100, 100, 
		[]framework.PodInfo {{"name", 50, 20},{"name", 50, 30},{"name", 50, 40}},
	}
	
	constraint := rr.ResourceRange{
		MinCpu: 10,
		MinMemory: 10,
		MaxCpu: 90,
		MaxMemory: 90,
	}
	
	if constraint.Apply(&node) {
		t.FailNow()
	}
}
