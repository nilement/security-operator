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
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ValidationSpec defines the desired state of Validation
type ValidationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Specifies the amount of experiments that need to be done to trigger
	ExperimentsToTrigger int `json:"experimentsToTrigger"`
	// Specifies the job that will be created when executing a Job.
	JobTemplate batchv1beta1.JobTemplateSpec `json:"jobTemplate"`
}

// ValidationStatus defines the observed state of Validation
type ValidationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	LastRun             metav1.Time `json:"lastScheduleTime,omitempty"`
	LastExperimentCount int         `json:"lastExperimentCount"`
	// +optional
	LastPod metav1.Time `json:"lastPod"`
	// +optional
	LastJob metav1.Time `json:"lastJob"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Validation is the Schema for the validations API
type Validation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ValidationSpec   `json:"spec,omitempty"`
	Status ValidationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ValidationList contains a list of Validation
type ValidationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Validation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Validation{}, &ValidationList{})
}
