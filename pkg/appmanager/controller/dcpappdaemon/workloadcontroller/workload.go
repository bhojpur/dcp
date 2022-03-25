package workloadcontroller

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

	unitv1alpha1 "github.com/bhojpur/dcp/pkg/appmanager/apis/apps/v1alpha1"
)

type Workload struct {
	Name      string
	Namespace string
	Kind      string
	Spec      WorkloadSpec
	Status    WorkloadStatus
}

// WorkloadSpec stores the spec details of the workload
type WorkloadSpec struct {
	Ref          metav1.Object
	Toleration   []corev1.Toleration
	NodeSelector map[string]string
}

// WorkloadStatus stores the observed state of the Workload.
type WorkloadStatus struct {
}

func (w *Workload) GetRevision() string {
	return w.Spec.Ref.GetLabels()[unitv1alpha1.ControllerRevisionHashLabelKey]
}

func (w *Workload) GetNodePoolName() string {
	return w.Spec.Ref.GetAnnotations()[unitv1alpha1.AnnotationRefNodePool]
}

func (w *Workload) GetToleration() []corev1.Toleration {
	return w.Spec.Toleration
}

func (w *Workload) GetNodeSelector() map[string]string {
	return w.Spec.NodeSelector
}

func (w *Workload) GetKind() string {
	return w.Kind
}
