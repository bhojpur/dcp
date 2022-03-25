package healthchecker

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
	"testing"

	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
)

func TestNodeLeaseManager_Update(t *testing.T) {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
			UID:  types.UID("foo-uid"),
		},
	}

	lease := &coordinationv1.Lease{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "coordination.k8s.io/v1",
			Kind:       "Lease",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "foo",
			Namespace:       "kube-node-lease",
			ResourceVersion: "115883910",
		},
	}

	gr := schema.GroupResource{Group: "v1", Resource: "lease"}
	noConnectionUpdateErr := apierrors.NewServerTimeout(gr, "put", 1)

	cases := []struct {
		desc          string
		updateReactor func(action clienttesting.Action) (bool, runtime.Object, error)
		getReactor    func(action clienttesting.Action) (bool, runtime.Object, error)
		success       bool
	}{
		{
			desc: "update success",
			updateReactor: func(action clienttesting.Action) (bool, runtime.Object, error) {
				return true, lease, nil
			},
			getReactor: func(action clienttesting.Action) (bool, runtime.Object, error) {
				return true, lease, nil
			},
			success: true,
		},
		{
			desc: "update fail",
			updateReactor: func(action clienttesting.Action) (bool, runtime.Object, error) {
				return true, nil, noConnectionUpdateErr
			},
			getReactor: func(action clienttesting.Action) (bool, runtime.Object, error) {
				return true, nil, noConnectionUpdateErr
			},
			success: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {

			cl := fake.NewSimpleClientset(node)
			cl.PrependReactor("update", "leases", tc.updateReactor)
			cl.PrependReactor("get", "leases", tc.getReactor)
			cl.PrependReactor("create", "leases", tc.updateReactor)
			nl := NewNodeLease(cl, "foo", 40, 3)
			if _, err := nl.Update(nil); tc.success != (err == nil) {
				t.Fatalf("got success %v, expected %v", err == nil, tc.success)
			}
		})
	}

}
