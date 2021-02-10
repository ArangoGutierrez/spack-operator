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

// SpackSpec defines the desired state of SpackPackage
// +k8s:openapi-gen=true
type SpackSpec struct {
	// ImageStream stores the stream where to push the built image
	ImageStream string `json:"imagestream,omitempty"`
	// Environment stores the spack.yaml env configuration file
	Environment string `json:"environment,omitempty"`
}

// SpackStatus defines the observed state of Spack
// +k8s:openapi-gen=true
type SpackStatus struct {
	State InstallStatus `json:"state"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Spack is the Schema for the Spack package builds API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Spack struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SpackSpec   `json:"spec,omitempty"`
	Status SpackStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SpackList contains a list of Spack
type SpackList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Spack `json:"items"`
}

// InstallStatus describes the state of installation of spack packages
type InstallStatus string

const (

	// EmptyStatus indicates that the spack package builds have not even been
	// created
	EmptyStatus InstallStatus = "empty"

	// AppliedStatus indicates that the spack package builds have been
	// created
	AppliedStatus InstallStatus = "applied"

	// ValidadtedPackage indicates that the spack package builds have been
	// validated
	ValidadtedPackage InstallStatus = "validated"

	// ErroredPackage indicates that the spack package builds status is
	// failing
	ErroredPackage InstallStatus = "error"
)

func init() {
	SchemeBuilder.Register(&Spack{}, &SpackList{})
}

// InstallStatus retrieves the status of a Spack CR
func (s *Spack) InstallStatus() InstallStatus {
	con := s.Status.State
	if len(con) == 0 {
		return EmptyStatus
	}
	return s.Status.State
}
