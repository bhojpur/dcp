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
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/bhojpur/dcp/pkg/appmanager/apis/apps/v1alpha1"
)

type WorkloadControllor interface {
	ObjectKey(load *Workload) client.ObjectKey
	GetAllWorkloads(set *v1alpha1.DcpAppDaemon) ([]*Workload, error)
	CreateWorkload(set *v1alpha1.DcpAppDaemon, nodepool v1alpha1.NodePool, revision string) error
	UpdateWorkload(load *Workload, set *v1alpha1.DcpAppDaemon, nodepool v1alpha1.NodePool, revision string) error
	DeleteWorkload(set *v1alpha1.DcpAppDaemon, load *Workload) error
	GetTemplateType() v1alpha1.TemplateType
}
