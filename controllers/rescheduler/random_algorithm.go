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

package rescheduler

import (
        "math/rand"
	constraints "github.com/miha3009/planner/controllers/constraints"
	preferences "github.com/miha3009/planner/controllers/preferences"
	types "github.com/miha3009/planner/controllers/types"
	helper "github.com/miha3009/planner/controllers/helper"
)

type RandomAlgorithm struct {
	Attempts int
	Constraints constraints.ConstraintList
	Preferences preferences.PreferenceList
}

func (a *RandomAlgorithm) Run(oldNodes []types.NodeInfo) []types.NodeInfo {
	if len(oldNodes) == 0 {
		return oldNodes
	}

	N := len(oldNodes)
	nodes := helper.DeepCopyNodes(oldNodes)
	
	for j := 0; j <= a.Attempts; j++{
		oldNode := &nodes[rand.Intn(N)]
		newNode := &nodes[rand.Intn(N)]

		if oldNode == newNode {
			continue
		}
	
		M := len(oldNode.Pods)
		if M == 0 {
			continue
		}
		podNum := rand.Intn(M)
	
		movement := types.MovementInfo {
			oldNode.Pods[podNum],
			*oldNode,
			*newNode,
		}
		
		if a.Constraints.ApplyForMove(&movement){
			oldScore, newScore := a.Preferences.ApplyForMove(&movement)
			if newScore > oldScore {
				a.doMovement(oldNode, newNode, podNum)
			}
		}
		
	}
	
	return nodes
}

func (a *RandomAlgorithm) doMovement(oldNode *types.NodeInfo, newNode *types.NodeInfo, podNum int) {
	newNode.Pods = append(newNode.Pods, oldNode.Pods[podNum])
	oldNode.Pods = append(oldNode.Pods[:podNum], oldNode.Pods[podNum+1:]...)
}

