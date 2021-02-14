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

// BuildSpec defines the desired state of a package
// +k8s:openapi-gen=true
type BuildSpec struct {
	// ImageStream stores the stream where to push the built image
	ImageStream string `json:"imagestream,omitempty"`
	// Environment stores the spack.yaml env configuration file
	Environment []SpackEnvionment `json:"environment,omitempty"`
}

// BuildStatus defines the observed state of a build
// +k8s:openapi-gen=true
type BuildStatus struct {
	State InstallStatus `json:"state"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Build is the Schema for the package builds API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Build struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec holds all the input necessary to produce a new package, and the conditions when
	// to trigger them.
	Spec BuildSpec `json:"spec,omitempty"`
	// status holds any relevant information about a build config
	// +optional
	Status BuildStatus `json:"status,omitempty"`
}

// SpackEnvionment holds the definition of a Spack Environment.
type SpackEnvionment struct {
	// Name of the Spack Environment profile to be used in buildConfig.
	Name *string `json:"name"`
	// Specification of the Spack Environment to be consumed by the Spack builder.
	Data *string `json:"data"`
}

// +kubebuilder:object:root=true

// BuildList contains a list of a build
type BuildList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Build `json:"items"`
}

// InstallStatus describes the state of installation of a package
type InstallStatus string

const (

	// EmptyStatus indicates that the package build have not even been
	// created
	EmptyStatus InstallStatus = "empty"

	// AppliedStatus indicates that the package build have been
	// created
	AppliedStatus InstallStatus = "applied"

	// ValidatedPackage indicates that the package build have been
	// validated
	ValidatedPackage InstallStatus = "validated"

	// ErroredPackage indicates that the package build status is
	// failing
	ErroredPackage InstallStatus = "error"
)

func init() {
	SchemeBuilder.Register(&Build{}, &BuildList{})
}

// InstallStatus retrieves the status of a Build CR
func (s *Build) InstallStatus() InstallStatus {
	con := s.Status.State
	if len(con) == 0 {
		return EmptyStatus
	}
	return s.Status.State
}
