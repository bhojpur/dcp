package adapter

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
	"k8s.io/apimachinery/pkg/runtime"

	alpha1 "github.com/bhojpur/dcp/pkg/appmanager/apis/apps/v1alpha1"
)

type Adapter interface {
	// NewResourceObject creates a empty pool object.
	NewResourceObject() runtime.Object
	// NewResourceListObject creates a empty pool list object.
	NewResourceListObject() runtime.Object
	// GetStatusObservedGeneration returns the observed generation of the pool.
	GetStatusObservedGeneration(pool metav1.Object) int64
	// GetDetails returns the replicas information of the pool status.
	GetDetails(pool metav1.Object) (replicasInfo ReplicasInfo, err error)
	// GetPoolFailure returns failure information of the pool.
	GetPoolFailure() *string
	// ApplyPoolTemplate updates the pool to the latest revision.
	ApplyPoolTemplate(ud *alpha1.UnitedDeployment, poolName, revision string, replicas int32, pool runtime.Object) error
	// IsExpected checks the pool is the expected revision or not.
	// If not, UnitedDeployment will call ApplyPoolTemplate to update it.
	IsExpected(pool metav1.Object, revision string) bool
	// PostUpdate does some works after pool updated
	PostUpdate(ud *alpha1.UnitedDeployment, pool runtime.Object, revision string) error
}

type ReplicasInfo struct {
	Replicas      int32
	ReadyReplicas int32
}
