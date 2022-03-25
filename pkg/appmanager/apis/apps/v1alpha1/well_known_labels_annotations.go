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

// UnitedDeployment related labels and annotations
const (
	// ControllerRevisionHashLabelKey is used to record the controller revision of current resource.
	ControllerRevisionHashLabelKey = "apps.bhojpur.net/controller-revision-hash"

	// PoolNameLabelKey is used to record the name of current pool.
	PoolNameLabelKey = "apps.bhojpur.net/pool-name"

	// SpecifiedDeleteKey indicates this object should be deleted, and the value could be the deletion option.
	SpecifiedDeleteKey = "apps.bhojpur.net/specified-delete"

	// AnnotationPatchKey indicates the patch for every sub pool
	AnnotationPatchKey = "apps.bhojpur.net/patch"

	AnnotationRefNodePool = "apps.bhojpur.net/ref-nodepool"
)

// NodePool related labels and annotations
const (
	// LabelDesiredNodePool indicates which nodepool the node want to join
	LabelDesiredNodePool = "apps.bhojpur.net/desired-nodepool"

	// LabelCurrentNodePool indicates which nodepool the node is currently
	// belonging to
	LabelCurrentNodePool = "apps.bhojpur.net/nodepool"

	// LabelCurrentDcpAppDaemon indicates which service the dcpappdaemon is currently
	// belonging to
	LabelCurrentDcpAppDaemon = "apps.bhojpur.net/dcpappdaemon"

	AnnotationPrevAttrs = "nodepool.bhojpur.net/previous-attributes"

	// DefaultCloudNodePoolName defines the name of the default cloud nodepool
	DefaultCloudNodePoolName = "default-nodepool"

	// DefaultEdgeNodePoolName defines the name of the default edge nodepool
	DefaultEdgeNodePoolName = "default-edge-nodepool"

	// ServiceTopologyKey is the toplogy key that will be attached to node,
	// the value will be the name of the nodepool
	ServiceTopologyKey = "topology.kubernetes.io/zone"
)
