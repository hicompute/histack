/*
Copyright 2025.

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

// ClusterIPSpec defines the desired state of ClusterIP
type ClusterIPSpec struct {
	ClusterIPPool string `json:"clusterIPPool"`
	Interface     string `json:"containerInterface"`
	Address       string `json:"address"`
	Mac           string `json:"mac,omitempty"`
	// +kubebuilder:validation:Enum=v4;v6
	Family   string `json:"family"`
	Resource string `json:"resource"`
}

type ClusterIPHistory struct {
	Mac         string      `json:"mac"`
	Interface   string      `json:"interface,omitempty"`
	Resource    string      `json:"resource"`
	AllocatedAt metav1.Time `json:"allocatedAt"`
	ReleasedAt  metav1.Time `json:"releasedAt,omitempty"`
}

// ClusterIPStatus defines the observed state of ClusterIP.
type ClusterIPStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the ClusterIP resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// Standard condition types include:
	// - "Available": the resource is fully functional
	// - "Progressing": the resource is being created or updated
	// - "Degraded": the resource failed to reach or maintain its desired state
	//
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	History    []ClusterIPHistory `json:"history,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Resource",type=string,JSONPath=.spec.resource
// +kubebuilder:printcolumn:name="MAC",type=string,JSONPath=.spec.mac
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=.metadata.creationTimestamp
// +kubebuilder:selectablefield:JSONPath=.spec.resource
// +kubebuilder:selectablefield:JSONPath=.spec.containerInterface
// +kubebuilder:selectablefield:JSONPath=.spec.family
// +kubebuilder:selectablefield:JSONPath=.spec.mac
type ClusterIP struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of ClusterIP
	// +required
	Spec ClusterIPSpec `json:"spec"`

	// status defines the observed state of ClusterIP
	// +optional
	Status ClusterIPStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// ClusterIPList contains a list of ClusterIP
type ClusterIPList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterIP `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterIP{}, &ClusterIPList{})
}
