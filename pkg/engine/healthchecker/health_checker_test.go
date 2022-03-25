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
	"net/url"
	"os"
	"testing"
	"time"

	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientfake "k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/pkg/engine/cachemanager"
	"github.com/bhojpur/dcp/pkg/engine/storage/disk"
)

var (
	rootDir = "/tmp/healthz"
)

func TestHealthyCheckrWithHealthyServer(t *testing.T) {
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
		remoteServers []*url.URL
		updateReactor []func(action clienttesting.Action) (bool, runtime.Object, error)
		getReactor    []func(action clienttesting.Action) (bool, runtime.Object, error)
		isHealthy     [][]bool
	}{
		{
			desc: "healthy",
			remoteServers: []*url.URL{
				{Host: "127.0.0.1:18080"},
			},
			updateReactor: []func(action clienttesting.Action) (bool, runtime.Object, error){
				func(action clienttesting.Action) (bool, runtime.Object, error) {
					return true, lease, nil
				},
			},
			getReactor: []func(action clienttesting.Action) (bool, runtime.Object, error){func(action clienttesting.Action) (bool, runtime.Object, error) {
				return true, lease, nil
			}},
			isHealthy: [][]bool{{true}},
		},
		{
			desc: "unhealthy",
			remoteServers: []*url.URL{
				{Host: "127.0.0.1:18080"},
			},
			updateReactor: []func(action clienttesting.Action) (bool, runtime.Object, error){
				func(action clienttesting.Action) (bool, runtime.Object, error) {
					return true, nil, noConnectionUpdateErr
				},
			},
			getReactor: []func(action clienttesting.Action) (bool, runtime.Object, error){
				func(action clienttesting.Action) (bool, runtime.Object, error) {
					return true, nil, noConnectionUpdateErr
				},
			},
			isHealthy: [][]bool{{false}},
		},
		{
			desc: "two-healthy",
			remoteServers: []*url.URL{
				{Host: "127.0.0.1:18080"},
				{Host: "127.0.0.1:18081"},
			},
			updateReactor: []func(action clienttesting.Action) (bool, runtime.Object, error){
				func(action clienttesting.Action) (bool, runtime.Object, error) {
					return true, lease, nil
				},
				func(action clienttesting.Action) (bool, runtime.Object, error) {
					return true, lease, nil
				},
			},
			getReactor: []func(action clienttesting.Action) (bool, runtime.Object, error){
				func(action clienttesting.Action) (bool, runtime.Object, error) {
					return true, lease, nil
				},
				func(action clienttesting.Action) (bool, runtime.Object, error) {
					return true, lease, nil
				},
			},
			isHealthy: [][]bool{{true, true}},
		},
		{
			desc: "two-unhealthy",
			remoteServers: []*url.URL{
				{Host: "127.0.0.1:18080"},
				{Host: "127.0.0.1:18081"},
			},
			updateReactor: []func(action clienttesting.Action) (bool, runtime.Object, error){
				func(action clienttesting.Action) (bool, runtime.Object, error) {
					return true, nil, noConnectionUpdateErr
				},
				func(action clienttesting.Action) (bool, runtime.Object, error) {
					return true, nil, noConnectionUpdateErr
				},
			},
			getReactor: []func(action clienttesting.Action) (bool, runtime.Object, error){
				func(action clienttesting.Action) (bool, runtime.Object, error) {
					return true, nil, noConnectionUpdateErr
				},
				func(action clienttesting.Action) (bool, runtime.Object, error) {
					return true, nil, noConnectionUpdateErr
				},
			},
			isHealthy: [][]bool{{false, false}},
		},
		{
			desc: "one-healthy one-unhealthy",
			remoteServers: []*url.URL{
				{Host: "127.0.0.1:18080"},
				{Host: "127.0.0.1:18081"},
			},
			updateReactor: []func(action clienttesting.Action) (bool, runtime.Object, error){
				func(action clienttesting.Action) (bool, runtime.Object, error) {
					return true, nil, noConnectionUpdateErr
				},
				func(action clienttesting.Action) (bool, runtime.Object, error) {
					return true, lease, nil
				},
			},
			getReactor: []func(action clienttesting.Action) (bool, runtime.Object, error){
				func(action clienttesting.Action) (bool, runtime.Object, error) {
					return true, nil, noConnectionUpdateErr
				},
				func(action clienttesting.Action) (bool, runtime.Object, error) {
					return true, lease, nil
				},
			},
			isHealthy: [][]bool{{false, true}},
		},
		{
			desc: "healthy to unhealthy",
			remoteServers: []*url.URL{
				{Host: "127.0.0.1:18080"},
			},
			updateReactor: []func(action clienttesting.Action) (bool, runtime.Object, error){
				func() func(action clienttesting.Action) (bool, runtime.Object, error) {
					i := 0
					return func(action clienttesting.Action) (bool, runtime.Object, error) {
						i++
						switch i {
						case 1:
							return true, lease, nil
						default:
							return true, nil, noConnectionUpdateErr
						}
					}
				}(),
			},
			getReactor: []func(action clienttesting.Action) (bool, runtime.Object, error){
				func() func(action clienttesting.Action) (bool, runtime.Object, error) {
					i := 0
					return func(action clienttesting.Action) (bool, runtime.Object, error) {
						i++
						switch i {
						case 1:
							return true, lease, nil
						default:
							return true, nil, noConnectionUpdateErr
						}
					}
				}(),
			},
			isHealthy: [][]bool{{true}, {false}},
		},
		{
			desc: "unhealthy to healthy",
			remoteServers: []*url.URL{
				{Host: "127.0.0.1:18080"},
			},
			updateReactor: []func(action clienttesting.Action) (bool, runtime.Object, error){
				func() func(action clienttesting.Action) (bool, runtime.Object, error) {
					i := 0
					return func(action clienttesting.Action) (bool, runtime.Object, error) {
						i++
						switch i {
						case 1:
							return true, nil, noConnectionUpdateErr
						default:
							return true, lease, nil
						}
					}
				}(),
			},
			getReactor: []func(action clienttesting.Action) (bool, runtime.Object, error){
				func() func(action clienttesting.Action) (bool, runtime.Object, error) {
					i := 0
					return func(action clienttesting.Action) (bool, runtime.Object, error) {
						i++
						switch i {
						case 1:
							return true, nil, noConnectionUpdateErr
						default:
							return true, lease, nil
						}
					}
				}(),
			},
			isHealthy: [][]bool{{false}, {true}},
		},
	}

	store, err := disk.NewDiskStorage(rootDir)
	if err != nil {
		t.Errorf("failed to create disk storage, %v", err)
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			stopCh := make(chan struct{})
			hcm := &healthCheckerManager{
				checkers:          make(map[string]*checker),
				remoteServers:     tc.remoteServers,
				remoteServerIndex: 0,
				sw:                cachemanager.NewStorageWrapper(store),
				stopCh:            stopCh,
			}

			for i, server := range tc.remoteServers {
				cl := clientfake.NewSimpleClientset(node)
				cl.PrependReactor("update", "leases", tc.updateReactor[i])
				cl.PrependReactor("get", "leases", tc.getReactor[i])
				cl.PrependReactor("create", "leases", tc.updateReactor[i])
				nl := NewNodeLease(cl, "foo", defaultLeaseDurationSeconds, 3)
				c := &checker{
					remoteServer:     server,
					clusterHealthy:   tc.isHealthy[0][i],
					healthyThreshold: 2,
					healthyCnt:       0,
					nodeLease:        nl,
					lastTime:         time.Now(),
					getLastNodeLease: hcm.getLastNodeLease,
					setLastNodeLease: hcm.setLastNodeLease,
				}
				hcm.checkers[server.String()] = c
			}

			hcm.Run()

			for i := range tc.isHealthy {
				klog.Infof("begin sleep 16s: %v", time.Now())
				time.Sleep(16 * time.Second)
				for j, server := range tc.remoteServers {
					if hcm.IsHealthy(server) != tc.isHealthy[i][j] {
						t.Fatalf("got %v, expected %v", hcm.IsHealthy(server), tc.isHealthy[i][j])
					}
				}
			}

			close(stopCh)
		})
	}

	if err := os.RemoveAll(rootDir); err != nil {
		t.Errorf("Got error %v, unable to remove path %s", err, rootDir)
	}
}
