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

package constraints

import (
	appsv1 "github.com/miha3009/planner/api/v1"
	types "github.com/miha3009/planner/controllers/types"
	base "github.com/miha3009/planner/controllers/rescheduler/constraints/base"
	ports "github.com/miha3009/planner/controllers/rescheduler/constraints/ports"
	podscount "github.com/miha3009/planner/controllers/rescheduler/constraints/podscount"
	podaffinity "github.com/miha3009/planner/controllers/rescheduler/constraints/podaffinity"
	tainttoleration "github.com/miha3009/planner/controllers/rescheduler/constraints/tainttoleration"
	resourcerange "github.com/miha3009/planner/controllers/rescheduler/constraints/resourcerange"
	"github.com/prometheus/common/log"
)

type ConstraintList struct {
	Items []Constraint
}

func ConvertArgs(cst *appsv1.ConstraintArgsList) ConstraintList {
	cl := make([]Constraint, 4)
	cl[0] = base.Base{}
	cl[1] = ports.Ports{}
	cl[2] = tainttoleration.TaintToleration{}
	cl[3] = podaffinity.PodAffinity{}
	
	if cst.ResourceRange != nil {
		if resourcerange.Validate(cst.ResourceRange) {
			cl = append(cl, resourcerange.ResourceRange{Args: *cst.ResourceRange})
		} else {
			log.Info("Constraint 'resource range' has wrong args")
		}
	}
	
	if cst.PodsCount != nil {
		cl = append(cl, podscount.PodsCount{Args: *cst.PodsCount})
	}
	
	return ConstraintList{Items: cl}
}

func (cl *ConstraintList) Init(nodes []types.NodeInfo) {
	for i := range nodes {
		for j := range cl.Items {
			cl.Items[j].Init(&nodes[i])
		}
	}
}

func (cl *ConstraintList) AddPod(node *types.NodeInfo, pod *types.PodInfo) {
	for i := range cl.Items {
		cl.Items[i].AddPod(node, pod)
	}
}

func (cl *ConstraintList) RemovePod(node *types.NodeInfo, pod *types.PodInfo) {
	for i := range cl.Items {
		cl.Items[i].RemovePod(node, pod)
	}
}

func (cl *ConstraintList) CheckForAll(nodes []types.NodeInfo) bool {
	for _, node := range nodes {
		if !cl.Check(&node) {
			return false
		}
	}
	return true
}

func (cl *ConstraintList) Check(node *types.NodeInfo) bool {
	for _, c := range cl.Items {
		if !c.Check(node) {
			return false
		}
	}
	return true
}

func (cl *ConstraintList) CheckForMove(move types.MovementInfo) bool {
	move.NewNode.AddPod(move.Pod)
	cl.AddPod(&move.NewNode, &move.Pod)
	move.OldNode.RemovePod(move.Pod)
	cl.RemovePod(&move.OldNode, &move.Pod)

	return cl.Check(&move.NewNode) && cl.Check(&move.OldNode)
}
