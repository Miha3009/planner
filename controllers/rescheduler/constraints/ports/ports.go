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

package ports

import (
	types "github.com/miha3009/planner/controllers/types"
)

type Ports struct {}

func (r Ports) Init(node *types.NodeInfo) {
	node.Ports = make(map[string]map[int32]struct{})
	node.PortsConflict = make(map[string]map[int32]int)
	for i := range node.Pods {
		r.AddPod(node, &node.Pods[i])
	}
}

func (r Ports) AddPod(node *types.NodeInfo, pod *types.PodInfo) {
	p := pod.Pod
	for i := range p.Spec.Containers {
		container := &p.Spec.Containers[i]
		for j := range container.Ports {
			ip := container.Ports[j].HostIP
			port := container.Ports[j].HostPort
			r.addPort(node, ip, port)
		}
	}
}

func (r Ports) RemovePod(node *types.NodeInfo, pod *types.PodInfo) {
	p := pod.Pod
	for i := range p.Spec.Containers {
		container := &p.Spec.Containers[i]
		for j := range container.Ports {
			ip := container.Ports[j].HostIP
			port := container.Ports[j].HostPort
			r.removePort(node, ip, port)
		}
	}
}

func (r Ports) Check(node *types.NodeInfo) bool {
	return len(node.PortsConflict) == 0
}

func (r Ports) addPort(node *types.NodeInfo, ip string, port int32) {
	if _, ok := node.Ports[ip]; !ok {
		node.Ports[ip] = make(map[int32]struct{})
	}
	if _, ok := node.Ports[ip][port]; !ok {
		if _, ok2 := node.PortsConflict[ip]; !ok2 {
			node.PortsConflict[ip] = make(map[int32]int)					
		}
		if _, ok2 := node.PortsConflict[ip][port]; !ok2 {
			node.PortsConflict[ip][port] = 2
		} else {
			node.PortsConflict[ip][port]++
		}
	}
	node.Ports[ip][port] = struct{}{}
}

func (r Ports) removePort(node *types.NodeInfo, ip string, port int32) {
	delete(node.Ports[ip], port)
	
	if len(node.Ports[ip]) == 0 {
		delete(node.Ports, ip)
	}

	if _, ok := node.PortsConflict[ip]; ok {
		if val, ok2 := node.PortsConflict[ip][port]; ok2 {
			if val == 2 {
				delete(node.PortsConflict[ip], port)
			} else {
				node.PortsConflict[ip][port]--
			}
		}
		
		if len(node.PortsConflict[ip]) == 0 {
			delete(node.PortsConflict, ip)
		}
	}
}

