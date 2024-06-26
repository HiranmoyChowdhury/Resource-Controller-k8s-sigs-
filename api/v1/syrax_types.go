/*
Copyright 2024.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SyraxSpec defines the desired state of Syrax
type SyraxSpec struct {
	DeletionPolicy DeletionPolicy    `json:"deletionPolicy,omitempty"`
	DeploymentSpec DeploymentSpec    `json:"deploymentSpec"`
	ServiceSpec    ServiceSpec       `json:"serviceSpec,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
}

const (
	DeletionPolicyDelete  DeletionPolicy = "Delete"
	DeletionPolicyWipeOut DeletionPolicy = "WipeOut"
)

type DeletionPolicy string

type DeploymentSpec struct {
	// +optional
	Name     string   `json:"name,omitempty"`
	Replicas *int32   `json:"replicas,omitempty"`
	Image    string   `json:"image"`
	Commands []string `json:"commands,omitempty"`
}
type ServiceSpec struct {
	Name        string             `json:"name,omitempty"`
	ServiceType corev1.ServiceType `json:"type,omitempty"`
	Port        *int32             `json:"port,omitempty"`
	TargetPort  *int32             `json:"targetPort,omitempty"`
	NodePort    *int32             `json:"NodePort,omitempty"`
}

// SyraxStatus defines the observed state of Syrax
type SyraxStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	AvailableReplicas *int32 `json:"availableReplicas"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Syrax is the Schema for the syraxs API
type Syrax struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SyraxSpec   `json:"spec,omitempty"`
	Status SyraxStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SyraxList contains a list of Syrax
type SyraxList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Syrax `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Syrax{}, &SyraxList{})
}
