package v1alpha1

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NodePoolType string

const (
	Edge  NodePoolType = "Edge"
	Cloud NodePoolType = "Cloud"
)

// NodePoolSpec defines the desired state of NodePool
type NodePoolSpec struct {
	// The type of the NodePool
	// +optional
	Type NodePoolType `json:"type,omitempty"`

	// A label query over nodes to consider for adding to the pool
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	// If specified, the Labels will be added to all nodes.
	// NOTE: existing labels with samy keys on the nodes will be overwritten.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// If specified, the Annotations will be added to all nodes.
	// NOTE: existing labels with samy keys on the nodes will be overwritten.
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// If specified, the Taints will be added to all nodes.
	// +optional
	Taints []v1.Taint `json:"taints,omitempty"`
}

// NodePoolStatus defines the observed state of NodePool
type NodePoolStatus struct {
	// Total number of ready nodes in the pool.
	// +optional
	ReadyNodeNum int32 `json:"readyNodeNum"`

	// Total number of unready nodes in the pool.
	// +optional
	UnreadyNodeNum int32 `json:"unreadyNodeNum"`

	// The list of nodes' names in the pool
	// +optional
	Nodes []string `json:"nodes,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,path=nodepools,shortName=np,categories=all
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type",description="The type of nodepool"
// +kubebuilder:printcolumn:name="ReadyNodes",type="integer",JSONPath=".status.readyNodeNum",description="The number of ready nodes in the pool"
// +kubebuilder:printcolumn:name="NotReadyNodes",type="integer",JSONPath=".status.unreadyNodeNum"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +genclient:nonNamespaced

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient
// NodePool is the Schema for the nodepools API
type NodePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodePoolSpec   `json:"spec,omitempty"`
	Status NodePoolStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NodePoolList contains a list of NodePool
type NodePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodePool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodePool{}, &NodePoolList{})
}
