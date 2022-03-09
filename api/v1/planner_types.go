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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ResourceRangeArgs struct {
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	MinCpu int64 `json:"minCpu,omitempty"`

	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	MaxCpu int64 `json:"maxCpu,omitempty"`

	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	MinMemory int64 `json:"minMemory,omitempty"`

	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	MaxMemory int64 `json:"maxMemory,omitempty"`
}

type ConstraintArgsList struct {
	ResourceRange *ResourceRangeArgs `json:"resource_range,omitempty"`
}

type UniformArgs struct {
	// +kubebuilder:validation:Minimum=0
	Weight float64 `json:"weight"`
}

type MaximizeInequalityArgs struct {
	// +kubebuilder:validation:Minimum=0
	Weight float64 `json:"weight"`
}

type PreferenceArgsList struct {
	Uniform *UniformArgs `json:"uniform,omitempty"`
	MaximizeInequality *MaximizeInequalityArgs `json:"maximize_inequality,omitempty"`
}

// PlannerSpec defines the desired state of Planner
type PlannerSpec struct {

	Active bool `json:"active"`

	Namespaces []string `json:"namespaces,omitempty"`
	
	
	// +kubebuilder:validation:Minimum=1
	Delay int `json:"delay,omitempty"`
	
	Constraints ConstraintArgsList `json:"constraints,omitempty"`
	
	Preferences PreferenceArgsList `json:"preferences,omitempty"`
}

// PlannerStatus defines the observed state of Planner
type PlannerStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Planner is the Schema for the planners API
type Planner struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PlannerSpec   `json:"spec,omitempty"`
	Status PlannerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PlannerList contains a list of HelloApp
type PlannerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Planner `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Planner{}, &PlannerList{})
}
