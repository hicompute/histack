package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
	Conditions         []metav1.Condition `json:"conditions"`
	TotalIPs           string             `json:"totalIPs"`
	AllocatedIPs       string             `json:"allocatedIPs"`
	FreeIPs            string             `json:"freeIPs"`
	NextIndex          string             `json:"nextIndex"`
	ReleasedClusterIPs []string           `json:"releasedClusterIPs,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:selectablefield:JSONPath=.spec.ipFamily
type ClusterIPPool struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata"`

	// spec defines the desired state of ClusterIPPool
	// +required
	Spec ClusterIPPoolSpec `json:"spec"`

	// status defines the observed state of ClusterIPPool
	// +optional
	Status ClusterIPPoolStatus `json:"status"`
}

// +kubebuilder:object:root=true

// ClusterIPPoolList contains a list of ClusterIPPool
type ClusterIPPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []ClusterIPPool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterIPPool{}, &ClusterIPPoolList{})
}
