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

package topologyspread

import (
	appsv1 "github.com/miha3009/planner/api/v1"
	types "github.com/miha3009/planner/controllers/types"
	helper "github.com/miha3009/planner/controllers/helper"
)

type TopologyGroup struct {
	Size int
	PodCount int
}

type TopologySpread struct {
	keys []appsv1.TopologyKey
	topologyCount map[string]map[string]TopologyGroup
}

func New(keys []appsv1.TopologyKey) TopologySpread {
	return TopologySpread {
		keys: keys,
		topologyCount: make(map[string]map[string]TopologyGroup),
	}
}

func (r TopologySpread) Init(node *types.NodeInfo) {
	for _, key := range r.keys {
		if val, ok := node.Node.Labels[key.Name]; ok {
			r.addToTopology(key.Name, val, false)
		}
	}
	for i := range node.Pods {
		r.AddPod(node, &node.Pods[i])
	}
}

func (r TopologySpread) AddPod(node *types.NodeInfo, pod *types.PodInfo) {
	for _, key := range r.keys {
		if val, ok := node.Node.Labels[key.Name]; ok {
			r.addToTopology(key.Name, val, true)
		}
	}
}

func (r TopologySpread) RemovePod(node *types.NodeInfo, pod *types.PodInfo) {
	for _, key := range r.keys {
		if val, ok := node.Node.Labels[key.Name]; ok {
			group := r.topologyCount[key.Name][val]
			group.PodCount--
			r.topologyCount[key.Name][val] = group
		}
	}
}

func (r TopologySpread) Apply(nodes []types.NodeInfo) float64 {
	weightSum := float64(0)
	for _, key := range r.keys {
		weightSum += float64(key.Weight)
	}

	scores := make([]float64, len(r.keys))
	
	for i := range r.keys {
		p := make([]float64, 0)
		for _, v := range r.topologyCount[r.keys[i].Name] {
			p = append(p, float64(v.PodCount) / float64(v.Size))
		}

		if len(p) == 0 {
			scores[i] = types.MaxPreferenceScore
			continue
		}

		max := helper.Max(p)
		for i := range p {
			p[i] /= max
		}

		maxVariance := float64(0.5)
		weight := float64(r.keys[i].Weight) / weightSum
		scores[i] = helper.Variance(p) / maxVariance * weight
	}
	
	finalScore := helper.Sum(scores)
	return finalScore
}

func (r TopologySpread) addToTopology(key, val string, pod bool) {
	if _, ok := r.topologyCount[key]; !ok {
		r.topologyCount[key] = make(map[string]TopologyGroup)
	}
	if _, ok := r.topologyCount[key][val]; !ok {
		r.topologyCount[key][val] = TopologyGroup{0, 0}
	}
	group := r.topologyCount[key][val]
	if pod {
		group.PodCount++
	} else {
		group.Size++
	}
	r.topologyCount[key][val] = group
}

