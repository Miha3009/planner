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

package preferences

import (
	appsv1 "github.com/miha3009/planner/api/v1"
	types "github.com/miha3009/planner/controllers/types"
	uniform "github.com/miha3009/planner/controllers/rescheduler/preferences/uniform"
	maximizeinequality "github.com/miha3009/planner/controllers/rescheduler/preferences/maximizeinequality"
)

type PreferenceList struct {
	Items []Preference
	Weights []float64
}

func ConvertArgs(prf *appsv1.PreferenceArgsList) PreferenceList {
	items := make([]Preference, 0)
	weights := make([]float64, 0)
	
	if prf.Uniform != nil {
		items = append(items, uniform.Uniform{})
		weights = append(weights, float64(prf.Uniform.Weight))
	}

	if prf.MaximizeInequality != nil {
		items = append(items, maximizeinequality.MaximizeInequality{})
		weights = append(weights, float64(prf.MaximizeInequality.Weight))
	}
	
	weightSum := float64(0)
	for _, weight := range weights {
		weightSum += weight
	}
	
	if weightSum < 0.1 {
		return PreferenceList{}
	}
	
	for i := range weights {
		weights[i] /= weightSum
	}
		
	return PreferenceList{Items: items, Weights: weights}
}

func (pl *PreferenceList) Apply(nodes []types.NodeInfo) float64 {
	score := float64(0)
	
	for i := range pl.Items {
		score += pl.Items[i].Apply(nodes) * pl.Weights[i]
	}
	
	return score
}

func (pl *PreferenceList) ApplyForMove(move *types.MovementInfo) (float64, float64) {
	podNum := -1
	for i, pod := range move.OldNode.Pods {
		if pod.Name == move.Pod.Name {
			podNum = i
			break
		}
	}

	if podNum == -1 {
		return 0.0, 0.0
	}

	oldScore := pl.Apply([]types.NodeInfo{move.OldNode, move.NewNode})

	move.NewNode.Pods = append(move.NewNode.Pods, move.Pod)
	move.OldNode.Pods = append(move.OldNode.Pods[:podNum], move.OldNode.Pods[podNum+1:]...)

	newScore := pl.Apply([]types.NodeInfo{move.OldNode, move.NewNode})
	
	return oldScore, newScore
}
