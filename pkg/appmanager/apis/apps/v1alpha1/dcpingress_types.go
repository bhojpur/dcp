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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Define the default nodepool ingress related values
const (
	// DefaultIngressControllerReplicasPerPool defines the default ingress controller replicas per pool
	DefaultIngressControllerReplicasPerPool int32 = 1
	// NginxIngressControllerVersion defines the nginx ingress controller version
	NginxIngressControllerVersion = "0.48.1"
	// SingletonDcpIngressInstanceName defines the singleton instance name of DcpIngress
	SingletonDcpIngressInstanceName = "ingress-singleton"
	// DcpIngressFinalizer is used to cleanup ingress resources when singleton DcpIngress CR is deleted
	DcpIngressFinalizer = "ingress.operator.bhojpur.net"
)

type IngressNotReadyType string

const (
	IngressPending IngressNotReadyType = "Pending"
	IngressFailure IngressNotReadyType = "Failure"
)

// IngressPool defines the details of a Pool for ingress
type IngressPool struct {
	// Indicates the pool name.
	Name string `json:"name"`

	// Pool specific configuration will be supported in future.
}

// IngressNotReadyConditionInfo defines the details info of an ingress not ready Pool
type IngressNotReadyConditionInfo struct {
	// Type of ingress not ready condition.
	Type IngressNotReadyType `json:"type,omitempty"`

	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`

	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

// IngressNotReadyPool defines the condition details of an ingress not ready Pool
type IngressNotReadyPool struct {
	// Indicates the pool name.
	Name string `json:"name"`

	// Info of ingress not ready condition.
	Info *IngressNotReadyConditionInfo `json:"poolinfo,omitempty"`
}

// DcpIngressSpec defines the desired state of DcpIngress
type DcpIngressSpec struct {
	// Indicates the number of the ingress controllers to be deployed under all the specified nodepools.
	// +optional
	Replicas int32 `json:"ingress_controller_replicas_per_pool,omitempty"`

	// Indicates all the nodepools on which to enable ingress.
	// +optional
	Pools []IngressPool `json:"pools,omitempty"`
}

// DcpIngressCondition describes current state of a DcpIngress
type DcpIngressCondition struct {
	// Indicates the pools that ingress controller is deployed successfully.
	IngressReadyPools []string `json:"ingressreadypools,omitempty"`

	// Indicates the pools that ingress controller is being deployed or deployed failed.
	IngressNotReadyPools []IngressNotReadyPool `json:"ingressunreadypools,omitempty"`
}

// DcpIngressStatus defines the observed state of DcpIngress
type DcpIngressStatus struct {
	// Indicates the number of the ingress controllers deployed under all the specified nodepools.
	// +optional
	Replicas int32 `json:"ingress_controller_replicas_per_pool,omitempty"`

	// Indicates all the nodepools on which to enable ingress.
	// +optional
	Conditions DcpIngressCondition `json:"conditions,omitempty"`

	// Indicates the nginx ingress controller version deployed under all the specified nodepools.
	// +optional
	Version string `json:"nginx_ingress_controller_version,omitempty"`

	// Total number of ready pools on which ingress is enabled.
	// +optional
	ReadyNum int32 `json:"readyNum"`

	// Total number of unready pools on which ingress is enabling or enable failed.
	// +optional
	UnreadyNum int32 `json:"unreadyNum"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,path=dcpingresses,shortName=ying,categories=all
// +kubebuilder:printcolumn:name="Nginx-Ingress-Version",type="string",JSONPath=".status.nginx_ingress_controller_version",description="The nginx ingress controller version"
// +kubebuilder:printcolumn:name="Replicas-Per-Pool",type="integer",JSONPath=".status.ingress_controller_replicas_per_pool",description="The nginx ingress controller replicas per pool"
// +kubebuilder:printcolumn:name="ReadyNum",type="integer",JSONPath=".status.readyNum",description="The number of pools on which ingress is enabled"
// +kubebuilder:printcolumn:name="NotReadyNum",type="integer",JSONPath=".status.unreadyNum",description="The number of pools on which ingress is enabling or enable failed"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +genclient:nonNamespaced
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DcpIngress is the Schema for the Bhojpur DCP ingresses API
type DcpIngress struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DcpIngressSpec   `json:"spec,omitempty"`
	Status DcpIngressStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DcpIngressList contains a list of DcpIngress
type DcpIngressList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DcpIngress `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DcpIngress{}, &DcpIngressList{})
}
