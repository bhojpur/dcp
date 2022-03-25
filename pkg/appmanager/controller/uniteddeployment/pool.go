package uniteddeployment

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
	"github.com/bhojpur/dcp/pkg/appmanager/controller/uniteddeployment/adapter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	unitv1alpha1 "github.com/bhojpur/dcp/pkg/appmanager/apis/apps/v1alpha1"
)

// Pool stores the details of a pool resource owned by one UnitedDeployment.
type Pool struct {
	Name      string
	Namespace string
	Spec      PoolSpec
	Status    PoolStatus
}

// PoolSpec stores the spec details of the Pool
type PoolSpec struct {
	PoolRef metav1.Object
}

// PoolStatus stores the observed state of the Pool.
type PoolStatus struct {
	ObservedGeneration int64
	adapter.ReplicasInfo
	PatchInfo string
}

// ResourceRef stores the Pool resource it represents.
type ResourceRef struct {
	Resources []metav1.Object
}

// ControlInterface defines the interface that UnitedDeployment uses to list, create, update, and delete Pools.
type ControlInterface interface {
	// GetAllPools returns the pools which are managed by the UnitedDeployment.
	GetAllPools(ud *unitv1alpha1.UnitedDeployment) ([]*Pool, error)
	// CreatePool creates the pool depending on the inputs.
	CreatePool(ud *unitv1alpha1.UnitedDeployment, unit string, revision string, replicas int32) error
	// UpdatePool updates the target pool with the input information.
	UpdatePool(pool *Pool, ud *unitv1alpha1.UnitedDeployment, revision string, replicas int32) error
	// DeletePool is used to delete the input pool.
	DeletePool(*Pool) error
	// GetPoolFailure extracts the pool failure message to expose on UnitedDeployment status.
	GetPoolFailure(*Pool) *string
	// IsExpected check the pool is the expected revision
	IsExpected(pool *Pool, revision string) bool
}
