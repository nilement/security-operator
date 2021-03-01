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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ChaosApiSpec defines the desired state of ChaosApi
type ChaosApiSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of ChaosApi. Edit ChaosApi_types.go to remove/update
	Size int32  `json:"size"`
	Foo  string `json:"foo,omitempty"`
}

// ChaosApiStatus defines the observed state of ChaosApi
type ChaosApiStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file`
	Nodes []string `json:"size"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ChaosApi is the Schema for the chaosapis API
// +kubebuilder:subresource:status
type ChaosApi struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ChaosApiSpec   `json:"spec,omitempty"`
	Status ChaosApiStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ChaosApiList contains a list of ChaosApi
type ChaosApiList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ChaosApi `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ChaosApi{}, &ChaosApiList{})
}
