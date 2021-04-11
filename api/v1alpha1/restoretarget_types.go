/*
Copyright 2021.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RestoreTargetSpec struct {
	DbName          string `json:"db_name"`
	StorageLocation string `json:"storage_location"`
	StorageType     string `json:"storage_type"`
}

// RestoreTargetStatus defines the observed state of RestoreTarget
type RestoreTargetStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RestoreTarget is the Schema for the restoretargets API
type RestoreTarget struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RestoreTargetSpec   `json:"spec,omitempty"`
	Status RestoreTargetStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RestoreTargetList contains a list of RestoreTarget
type RestoreTargetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RestoreTarget `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RestoreTarget{}, &RestoreTargetList{})
}
