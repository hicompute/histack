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

// ClusterIPPoolSpec defines the desired state of ClusterIPPool
type ClusterIPPoolSpec struct {
	// +kubebuilder:validation:Enum=v4;v6
	IPFamily string `json:"ipFamily"`
	CIDR     string `json:"cidr"`
	Gateway  string `json:"gateway,omitempty"`
}

// ClusterIPPoolStatus defines the observed state of ClusterIPPool.
type ClusterIPPoolStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the ClusterIPPool resource.
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
	Conditions   []metav1.Condition `json:"conditions,omitempty"`
	TotalIPs     uint64             `json:"totalIPs,omitempty"`
	AllocatedIPs uint64             `json:"allocatedIPs,omitempty"`
	FreeIPs      uint64             `json:"freeIPs,omitempty"`
	// +kubebuilder:default=0
	NextIndex          uint64   `json:"nextIndex"`
	ReleasedClusterIPs []string `json:"releasedClusterIPs,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:selectablefield:JSONPath=.spec.ipFamily
type ClusterIPPool struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of ClusterIPPool
	// +required
	Spec ClusterIPPoolSpec `json:"spec"`

	// status defines the observed state of ClusterIPPool
	// +optional
	Status ClusterIPPoolStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// ClusterIPPoolList contains a list of ClusterIPPool
type ClusterIPPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterIPPool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterIPPool{}, &ClusterIPPoolList{})
}
