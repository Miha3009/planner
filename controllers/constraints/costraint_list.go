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
	base "github.com/miha3009/planner/controllers/constraints/base"
	resourcerange "github.com/miha3009/planner/controllers/constraints/resourcerange"
	"github.com/prometheus/common/log"
)

type ConstraintList struct {
	Items []Constraint
}

func ConvertArgs(cst *appsv1.ConstraintArgsList) ConstraintList {
	cl := make([]Constraint, 1)
	cl[0] = base.Base{}
	
	if cst.ResourceRange != nil {
		if resourcerange.Validate(cst.ResourceRange) {
			cl = append(cl, resourcerange.ResourceRange{Args: *cst.ResourceRange})
		} else {
			log.Info("Constraint 'resource range' has wrong args")
		}
	}
	
	return ConstraintList{Items: cl}
}

func (cl *ConstraintList) ApplyForAll(nodes []types.NodeInfo) bool {
	for _, node := range nodes {
		if !cl.Apply(&node) {
			return false
		}
	}
	return true
}

func (cl *ConstraintList) Apply(node *types.NodeInfo) bool {
	for _, c := range cl.Items {
		if !c.Apply(node) {
			return false
		}
	}
	return true
}

func (cl *ConstraintList) ApplyForMove(move *types.MovementInfo) bool {
	podNum := -1
	for i, pod := range move.OldNode.Pods {
		if pod.Name == move.Pod.Name {
			podNum = i
			break
		}
	}

	if podNum == -1 {
		return false
	}

	move.NewNode.Pods = append(move.NewNode.Pods, move.Pod)
	move.OldNode.Pods = append(move.OldNode.Pods[:podNum], move.OldNode.Pods[podNum+1:]...)

	return cl.Apply(&move.NewNode) && cl.Apply(&move.OldNode)
}
