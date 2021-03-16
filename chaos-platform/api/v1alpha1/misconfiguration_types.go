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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MisconfigurationSpec defines the desired state of Misconfiguration
type MisconfigurationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// A list of pointers to currently running jobs.
	// +optional
	Active []corev1.ObjectReference `json:"active,omitempty"`

	// Information when was the last time a misconfiguration was successfully applied.
	// +optional
	LastMisconfiguration *metav1.Time `json:"lastScheduleTime,omitempty"`

	// Specifies the job that will be created when executing a CronJob.
	JobTemplate batchv1beta1.JobTemplateSpec `json:"jobTemplate"`

	// Specifies how many times it should be applied
	Completions *int64 `json:"completions"`
}

// MisconfigurationStatus defines the observed state of Misconfiguration
type MisconfigurationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Misconfiguration is the Schema for the misconfigurations API
type Misconfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MisconfigurationSpec   `json:"spec,omitempty"`
	Status MisconfigurationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MisconfigurationList contains a list of Misconfiguration
type MisconfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Misconfiguration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Misconfiguration{}, &MisconfigurationList{})
}
