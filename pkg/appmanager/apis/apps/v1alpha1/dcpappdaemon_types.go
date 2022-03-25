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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DcpAppDaemonConditionType indicates valid conditions type of a DcpAppDaemon.
type DcpAppDaemonConditionType string

const (
	// WorkLoadProvisioned means all the expected workload are provisioned
	WorkLoadProvisioned DcpAppDaemonConditionType = "WorkLoadProvisioned"
	// WorkLoadUpdated means all the workload are updated.
	WorkLoadUpdated DcpAppDaemonConditionType = "WorkLoadUpdated"
	// WorkLoadFailure is added to a UnitedDeployment when one of its workload has failure during its own reconciling.
	WorkLoadFailure DcpAppDaemonConditionType = "WorkLoadFailure"
)

// DcpAppDaemonSpec defines the desired state of DcpAppDaemon.
type DcpAppDaemonSpec struct {
	// Selector is a label query over pods that should match the replica count.
	// It must match the pod template's labels.
	Selector *metav1.LabelSelector `json:"selector"`

	// WorkloadTemplate describes the pool that will be created.
	// +optional
	WorkloadTemplate WorkloadTemplate `json:"workloadTemplate,omitempty"`

	// NodePoolSelector is a label query over nodepool that should match the replica count.
	// It must match the nodepool's labels.
	NodePoolSelector *metav1.LabelSelector `json:"nodepoolSelector"`

	// Indicates the number of histories to be conserved.
	// If unspecified, defaults to 10.
	// +optional
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit,omitempty"`
}

// DcpAppDaemonStatus defines the observed state of DcpAppDaemon.
type DcpAppDaemonStatus struct {
	// ObservedGeneration is the most recent generation observed for this DcpAppDaemon. It corresponds to the
	// DcpAppDaemon's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Count of hash collisions for the DcpAppDaemon. The DcpAppDaemon controller
	// uses this field as a collision avoidance mechanism when it needs to
	// create the name for the newest ControllerRevision.
	// +optional
	CollisionCount *int32 `json:"collisionCount,omitempty"`

	// CurrentRevision, if not empty, indicates the current version of the DcpAppDaemon.
	CurrentRevision string `json:"currentRevision"`

	// Represents the latest available observations of a DcpAppDaemon's current state.
	// +optional
	Conditions []DcpAppDaemonCondition `json:"conditions,omitempty"`

	// TemplateType indicates the type of PoolTemplate
	TemplateType TemplateType `json:"templateType"`

	// NodePools indicates the list of node pools selected by DcpAppDaemon
	NodePools []string `json:"nodepools,omitempty"`
}

// DcpAppDaemonCondition describes current state of a DcpAppDaemon.
type DcpAppDaemonCondition struct {
	// Type of in place set condition.
	Type DcpAppDaemonConditionType `json:"type,omitempty"`

	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status,omitempty"`

	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`

	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=yad
// +kubebuilder:printcolumn:name="WorkloadTemplate",type="string",JSONPath=".status.templateType",description="The WorkloadTemplate Type."
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp",description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC."

// DcpAppDaemon is the Schema for the DcpAppDaemon API
type DcpAppDaemon struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DcpAppDaemonSpec   `json:"spec,omitempty"`
	Status DcpAppDaemonStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DcpAppDaemonList contains a list of DcpAppDaemon
type DcpAppDaemonList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DcpAppDaemon `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DcpAppDaemon{}, &DcpAppDaemonList{})
}
