/*
Copyright 2025 mamrezb.

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

// ServiceCheckerSpec defines the desired state of ServiceChecker.
type ServiceCheckerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// List of services to check
	Services []NamedService `json:"services,omitempty"`
}

// NamedService is a reference to a K8s Service
type NamedService struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Critical  bool   `json:"critical,omitempty"`
}

// ServiceCheckerStatus defines the observed state of ServiceChecker.
type ServiceCheckerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ServiceStatuses []ServiceStatus `json:"serviceStatuses,omitempty"`
}

type ServiceStatus struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Ready     bool   `json:"ready"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ServiceChecker is the Schema for the servicecheckers API.
type ServiceChecker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceCheckerSpec   `json:"spec,omitempty"`
	Status ServiceCheckerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ServiceCheckerList contains a list of ServiceChecker.
type ServiceCheckerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceChecker `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceChecker{}, &ServiceCheckerList{})
}
