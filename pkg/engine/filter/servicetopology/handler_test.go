package servicetopology

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
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	nodepoolv1alpha1 "github.com/bhojpur/dcp/pkg/appmanager/apis/apps/v1alpha1"
	dcpfake "github.com/bhojpur/dcp/pkg/appmanager/client/clientset/versioned/fake"
	dcpinformers "github.com/bhojpur/dcp/pkg/appmanager/client/informers/externalversions"
)

func TestReassembleEndpointSlice(t *testing.T) {
	currentNodeName := "node1"

	testcases := map[string]struct {
		endpointSlice *discovery.EndpointSlice
		kubeClient    *k8sfake.Clientset
		dcpClient     *dcpfake.Clientset
		expectResult  *discovery.EndpointSlice
	}{
		"service with annotation bhojpur.net/topologyKeys: kubernetes.io/hostname": {
			endpointSlice: &discovery.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "svc1-np7sf",
					Namespace: "default",
					Labels: map[string]string{
						discovery.LabelServiceName: "svc1",
					},
				},
				Endpoints: []discovery.Endpoint{
					{
						Addresses: []string{
							"10.244.1.2",
						},
						Topology: map[string]string{
							corev1.LabelHostname: currentNodeName,
						},
					},
					{
						Addresses: []string{
							"10.244.1.3",
						},
						Topology: map[string]string{
							corev1.LabelHostname: "node2",
						},
					},
					{
						Addresses: []string{
							"10.244.1.4",
						},
						Topology: map[string]string{
							corev1.LabelHostname: currentNodeName,
						},
					},
					{
						Addresses: []string{
							"10.244.1.5",
						},
						Topology: map[string]string{
							corev1.LabelHostname: "node3",
						},
					},
				},
			},
			kubeClient: k8sfake.NewSimpleClientset(
				&corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: currentNodeName,
						Labels: map[string]string{
							nodepoolv1alpha1.LabelCurrentNodePool: "hangzhou",
						},
					},
				},
				&corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node2",
						Labels: map[string]string{
							nodepoolv1alpha1.LabelCurrentNodePool: "shanghai",
						},
					},
				},
				&corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node3",
						Labels: map[string]string{
							nodepoolv1alpha1.LabelCurrentNodePool: "hangzhou",
						},
					},
				},
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc1",
						Namespace: "default",
						Annotations: map[string]string{
							AnnotationServiceTopologyKey: AnnotationServiceTopologyValueNode,
						},
					},
				},
			),
			dcpClient: dcpfake.NewSimpleClientset(
				&nodepoolv1alpha1.NodePool{
					ObjectMeta: metav1.ObjectMeta{
						Name: "hangzhou",
					},
					Spec: nodepoolv1alpha1.NodePoolSpec{
						Type: nodepoolv1alpha1.Edge,
					},
					Status: nodepoolv1alpha1.NodePoolStatus{
						Nodes: []string{
							currentNodeName,
							"node3",
						},
					},
				},
				&nodepoolv1alpha1.NodePool{
					ObjectMeta: metav1.ObjectMeta{
						Name: "shanghai",
					},
					Spec: nodepoolv1alpha1.NodePoolSpec{
						Type: nodepoolv1alpha1.Edge,
					},
					Status: nodepoolv1alpha1.NodePoolStatus{
						Nodes: []string{
							"node2",
						},
					},
				},
			),
			expectResult: &discovery.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "svc1-np7sf",
					Namespace: "default",
					Labels: map[string]string{
						discovery.LabelServiceName: "svc1",
					},
				},
				Endpoints: []discovery.Endpoint{
					{
						Addresses: []string{
							"10.244.1.2",
						},
						Topology: map[string]string{
							corev1.LabelHostname: currentNodeName,
						},
					},
					{
						Addresses: []string{
							"10.244.1.4",
						},
						Topology: map[string]string{
							corev1.LabelHostname: currentNodeName,
						},
					},
				},
			},
		},
		"service with annotation bhojpur.net/topologyKeys: bhojpur.net/nodepool": {
			endpointSlice: &discovery.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "svc1-np7sf",
					Namespace: "default",
					Labels: map[string]string{
						discovery.LabelServiceName: "svc1",
					},
				},
				Endpoints: []discovery.Endpoint{
					{
						Addresses: []string{
							"10.244.1.2",
						},
						Topology: map[string]string{
							corev1.LabelHostname: currentNodeName,
						},
					},
					{
						Addresses: []string{
							"10.244.1.3",
						},
						Topology: map[string]string{
							corev1.LabelHostname: "node2",
						},
					},
					{
						Addresses: []string{
							"10.244.1.4",
						},
						Topology: map[string]string{
							corev1.LabelHostname: currentNodeName,
						},
					},
					{
						Addresses: []string{
							"10.244.1.5",
						},
						Topology: map[string]string{
							corev1.LabelHostname: "node3",
						},
					},
				},
			},
			kubeClient: k8sfake.NewSimpleClientset(
				&corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: currentNodeName,
						Labels: map[string]string{
							nodepoolv1alpha1.LabelCurrentNodePool: "hangzhou",
						},
					},
				},
				&corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node2",
						Labels: map[string]string{
							nodepoolv1alpha1.LabelCurrentNodePool: "shanghai",
						},
					},
				},
				&corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node3",
						Labels: map[string]string{
							nodepoolv1alpha1.LabelCurrentNodePool: "hangzhou",
						},
					},
				},
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc1",
						Namespace: "default",
						Annotations: map[string]string{
							AnnotationServiceTopologyKey: AnnotationServiceTopologyValueNodePool,
						},
					},
				},
			),
			dcpClient: dcpfake.NewSimpleClientset(
				&nodepoolv1alpha1.NodePool{
					ObjectMeta: metav1.ObjectMeta{
						Name: "hangzhou",
					},
					Spec: nodepoolv1alpha1.NodePoolSpec{
						Type: nodepoolv1alpha1.Edge,
					},
					Status: nodepoolv1alpha1.NodePoolStatus{
						Nodes: []string{
							currentNodeName,
							"node3",
						},
					},
				},
				&nodepoolv1alpha1.NodePool{
					ObjectMeta: metav1.ObjectMeta{
						Name: "shanghai",
					},
					Spec: nodepoolv1alpha1.NodePoolSpec{
						Type: nodepoolv1alpha1.Edge,
					},
					Status: nodepoolv1alpha1.NodePoolStatus{
						Nodes: []string{
							"node2",
						},
					},
				},
			),
			expectResult: &discovery.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "svc1-np7sf",
					Namespace: "default",
					Labels: map[string]string{
						discovery.LabelServiceName: "svc1",
					},
				},
				Endpoints: []discovery.Endpoint{
					{
						Addresses: []string{
							"10.244.1.2",
						},
						Topology: map[string]string{
							corev1.LabelHostname: currentNodeName,
						},
					},
					{
						Addresses: []string{
							"10.244.1.4",
						},
						Topology: map[string]string{
							corev1.LabelHostname: currentNodeName,
						},
					},
					{
						Addresses: []string{
							"10.244.1.5",
						},
						Topology: map[string]string{
							corev1.LabelHostname: "node3",
						},
					},
				},
			},
		},
		"service with annotation bhojpur.net/topologyKeys: kubernetes.io/zone": {
			endpointSlice: &discovery.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "svc1-np7sf",
					Namespace: "default",
					Labels: map[string]string{
						discovery.LabelServiceName: "svc1",
					},
				},
				Endpoints: []discovery.Endpoint{
					{
						Addresses: []string{
							"10.244.1.2",
						},
						Topology: map[string]string{
							corev1.LabelHostname: currentNodeName,
						},
					},
					{
						Addresses: []string{
							"10.244.1.3",
						},
						Topology: map[string]string{
							corev1.LabelHostname: "node2",
						},
					},
					{
						Addresses: []string{
							"10.244.1.4",
						},
						Topology: map[string]string{
							corev1.LabelHostname: currentNodeName,
						},
					},
					{
						Addresses: []string{
							"10.244.1.5",
						},
						Topology: map[string]string{
							corev1.LabelHostname: "node3",
						},
					},
				},
			},
			kubeClient: k8sfake.NewSimpleClientset(
				&corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: currentNodeName,
						Labels: map[string]string{
							nodepoolv1alpha1.LabelCurrentNodePool: "hangzhou",
						},
					},
				},
				&corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node2",
						Labels: map[string]string{
							nodepoolv1alpha1.LabelCurrentNodePool: "shanghai",
						},
					},
				},
				&corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node3",
						Labels: map[string]string{
							nodepoolv1alpha1.LabelCurrentNodePool: "hangzhou",
						},
					},
				},
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc1",
						Namespace: "default",
						Annotations: map[string]string{
							AnnotationServiceTopologyKey: AnnotationServiceTopologyValueNodePool,
						},
					},
				},
			),
			dcpClient: dcpfake.NewSimpleClientset(
				&nodepoolv1alpha1.NodePool{
					ObjectMeta: metav1.ObjectMeta{
						Name: "hangzhou",
					},
					Spec: nodepoolv1alpha1.NodePoolSpec{
						Type: nodepoolv1alpha1.Edge,
					},
					Status: nodepoolv1alpha1.NodePoolStatus{
						Nodes: []string{
							currentNodeName,
							"node3",
						},
					},
				},
				&nodepoolv1alpha1.NodePool{
					ObjectMeta: metav1.ObjectMeta{
						Name: "shanghai",
					},
					Spec: nodepoolv1alpha1.NodePoolSpec{
						Type: nodepoolv1alpha1.Edge,
					},
					Status: nodepoolv1alpha1.NodePoolStatus{
						Nodes: []string{
							"node2",
						},
					},
				},
			),
			expectResult: &discovery.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "svc1-np7sf",
					Namespace: "default",
					Labels: map[string]string{
						discovery.LabelServiceName: "svc1",
					},
				},
				Endpoints: []discovery.Endpoint{
					{
						Addresses: []string{
							"10.244.1.2",
						},
						Topology: map[string]string{
							corev1.LabelHostname: currentNodeName,
						},
					},
					{
						Addresses: []string{
							"10.244.1.4",
						},
						Topology: map[string]string{
							corev1.LabelHostname: currentNodeName,
						},
					},
					{
						Addresses: []string{
							"10.244.1.5",
						},
						Topology: map[string]string{
							corev1.LabelHostname: "node3",
						},
					},
				},
			},
		},
		"service without annotation bhojpur.net/topologyKeys": {
			endpointSlice: &discovery.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "svc1-np7sf",
					Namespace: "default",
					Labels: map[string]string{
						discovery.LabelServiceName: "svc1",
					},
				},
				Endpoints: []discovery.Endpoint{
					{
						Addresses: []string{
							"10.244.1.2",
						},
						Topology: map[string]string{
							corev1.LabelHostname: currentNodeName,
						},
					},
					{
						Addresses: []string{
							"10.244.1.3",
						},
						Topology: map[string]string{
							corev1.LabelHostname: "node2",
						},
					},
					{
						Addresses: []string{
							"10.244.1.4",
						},
						Topology: map[string]string{
							corev1.LabelHostname: currentNodeName,
						},
					},
					{
						Addresses: []string{
							"10.244.1.5",
						},
						Topology: map[string]string{
							corev1.LabelHostname: "node3",
						},
					},
				},
			},
			kubeClient: k8sfake.NewSimpleClientset(
				&corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: currentNodeName,
						Labels: map[string]string{
							nodepoolv1alpha1.LabelCurrentNodePool: "hangzhou",
						},
					},
				},
				&corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node2",
						Labels: map[string]string{
							nodepoolv1alpha1.LabelCurrentNodePool: "shanghai",
						},
					},
				},
				&corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node3",
						Labels: map[string]string{
							nodepoolv1alpha1.LabelCurrentNodePool: "hangzhou",
						},
					},
				},
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "svc1",
						Namespace:   "default",
						Annotations: map[string]string{},
					},
				},
			),
			dcpClient: dcpfake.NewSimpleClientset(
				&nodepoolv1alpha1.NodePool{
					ObjectMeta: metav1.ObjectMeta{
						Name: "hangzhou",
					},
					Spec: nodepoolv1alpha1.NodePoolSpec{
						Type: nodepoolv1alpha1.Edge,
					},
					Status: nodepoolv1alpha1.NodePoolStatus{
						Nodes: []string{
							currentNodeName,
							"node3",
						},
					},
				},
				&nodepoolv1alpha1.NodePool{
					ObjectMeta: metav1.ObjectMeta{
						Name: "shanghai",
					},
					Spec: nodepoolv1alpha1.NodePoolSpec{
						Type: nodepoolv1alpha1.Edge,
					},
					Status: nodepoolv1alpha1.NodePoolStatus{
						Nodes: []string{
							"node2",
						},
					},
				},
			),
			expectResult: &discovery.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "svc1-np7sf",
					Namespace: "default",
					Labels: map[string]string{
						discovery.LabelServiceName: "svc1",
					},
				},
				Endpoints: []discovery.Endpoint{
					{
						Addresses: []string{
							"10.244.1.2",
						},
						Topology: map[string]string{
							corev1.LabelHostname: currentNodeName,
						},
					},
					{
						Addresses: []string{
							"10.244.1.3",
						},
						Topology: map[string]string{
							corev1.LabelHostname: "node2",
						},
					},
					{
						Addresses: []string{
							"10.244.1.4",
						},
						Topology: map[string]string{
							corev1.LabelHostname: currentNodeName,
						},
					},
					{
						Addresses: []string{
							"10.244.1.5",
						},
						Topology: map[string]string{
							corev1.LabelHostname: "node3",
						},
					},
				},
			},
		},
		"currentNode is not in any nodepool": {
			endpointSlice: &discovery.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "svc1-np7sf",
					Namespace: "default",
					Labels: map[string]string{
						discovery.LabelServiceName: "svc1",
					},
				},
				Endpoints: []discovery.Endpoint{
					{
						Addresses: []string{
							"10.244.1.2",
						},
						Topology: map[string]string{
							corev1.LabelHostname: currentNodeName,
						},
					},
					{
						Addresses: []string{
							"10.244.1.3",
						},
						Topology: map[string]string{
							corev1.LabelHostname: "node2",
						},
					},
					{
						Addresses: []string{
							"10.244.1.4",
						},
						Topology: map[string]string{
							corev1.LabelHostname: currentNodeName,
						},
					},
					{
						Addresses: []string{
							"10.244.1.5",
						},
						Topology: map[string]string{
							corev1.LabelHostname: "node3",
						},
					},
				},
			},
			kubeClient: k8sfake.NewSimpleClientset(
				&corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name:   currentNodeName,
						Labels: map[string]string{},
					},
				},
				&corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node2",
						Labels: map[string]string{
							nodepoolv1alpha1.LabelCurrentNodePool: "shanghai",
						},
					},
				},
				&corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node3",
						Labels: map[string]string{
							nodepoolv1alpha1.LabelCurrentNodePool: "hangzhou",
						},
					},
				},
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "svc1",
						Namespace:   "default",
						Annotations: map[string]string{},
					},
				},
			),
			dcpClient: dcpfake.NewSimpleClientset(
				&nodepoolv1alpha1.NodePool{
					ObjectMeta: metav1.ObjectMeta{
						Name: "hangzhou",
					},
					Spec: nodepoolv1alpha1.NodePoolSpec{
						Type: nodepoolv1alpha1.Edge,
					},
					Status: nodepoolv1alpha1.NodePoolStatus{
						Nodes: []string{
							"node3",
						},
					},
				},
				&nodepoolv1alpha1.NodePool{
					ObjectMeta: metav1.ObjectMeta{
						Name: "shanghai",
					},
					Spec: nodepoolv1alpha1.NodePoolSpec{
						Type: nodepoolv1alpha1.Edge,
					},
					Status: nodepoolv1alpha1.NodePoolStatus{
						Nodes: []string{
							"node2",
						},
					},
				},
			),
			expectResult: &discovery.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "svc1-np7sf",
					Namespace: "default",
					Labels: map[string]string{
						discovery.LabelServiceName: "svc1",
					},
				},
				Endpoints: []discovery.Endpoint{
					{
						Addresses: []string{
							"10.244.1.2",
						},
						Topology: map[string]string{
							corev1.LabelHostname: currentNodeName,
						},
					},
					{
						Addresses: []string{
							"10.244.1.3",
						},
						Topology: map[string]string{
							corev1.LabelHostname: "node2",
						},
					},
					{
						Addresses: []string{
							"10.244.1.4",
						},
						Topology: map[string]string{
							corev1.LabelHostname: currentNodeName,
						},
					},
					{
						Addresses: []string{
							"10.244.1.5",
						},
						Topology: map[string]string{
							corev1.LabelHostname: "node3",
						},
					},
				},
			},
		},
	}

	for k, tt := range testcases {
		t.Run(k, func(t *testing.T) {
			tt.kubeClient.DiscoveryV1beta1().EndpointSlices("default").Create(context.TODO(), tt.endpointSlice, metav1.CreateOptions{})

			factory := informers.NewSharedInformerFactory(tt.kubeClient, 24*time.Hour)
			serviceInformer := factory.Core().V1().Services()
			serviceInformer.Informer()
			serviceLister := serviceInformer.Lister()

			stopper := make(chan struct{})
			defer close(stopper)
			factory.Start(stopper)
			factory.WaitForCacheSync(stopper)

			dcpFactory := dcpinformers.NewSharedInformerFactory(tt.dcpClient, 24*time.Hour)
			nodePoolInformer := dcpFactory.Apps().V1alpha1().NodePools()
			nodePoolLister := nodePoolInformer.Lister()

			stopper2 := make(chan struct{})
			defer close(stopper2)
			dcpFactory.Start(stopper2)
			dcpFactory.WaitForCacheSync(stopper2)

			nodeGetter := func(name string) (*corev1.Node, error) {
				return tt.kubeClient.CoreV1().Nodes().Get(context.TODO(), name, metav1.GetOptions{})
			}

			fh := &serviceTopologyFilterHandler{
				nodeName:       currentNodeName,
				serviceLister:  serviceLister,
				nodePoolLister: nodePoolLister,
				nodeGetter:     nodeGetter,
			}

			reassembledEndpointSlice := fh.reassembleEndpointSlice(tt.endpointSlice)

			if !isEqualEndpointSlice(reassembledEndpointSlice, tt.expectResult) {
				t.Errorf("reassembleEndpointSlice got error, expected: \n%v\nbut got: \n%v\n", tt.expectResult, reassembledEndpointSlice)
			}

		})
	}
}

// isEqualEndpointSlice is used to determine whether two endpointSlice are equal.
// Note that this function can only be used in this test.
func isEqualEndpointSlice(endpointSlice1, endpointSlice2 *discovery.EndpointSlice) bool {
	if endpointSlice1.Name != endpointSlice2.Name ||
		endpointSlice1.Namespace != endpointSlice2.Namespace ||
		endpointSlice1.Labels[discovery.LabelServiceName] != endpointSlice2.Labels[discovery.LabelServiceName] {
		return false
	}

	endpoints1 := endpointSlice1.Endpoints
	endpoints2 := endpointSlice2.Endpoints
	if len(endpoints1) != len(endpoints2) {
		return false
	}

	for i := 0; i < len(endpoints1); i++ {
		if !isEqualStrings(endpoints1[i].Addresses, endpoints2[i].Addresses) {
			return false
		}

		if endpoints1[i].Topology[corev1.LabelHostname] != endpoints2[i].Topology[corev1.LabelHostname] {
			return false
		}
	}

	return true
}

func isEqualStrings(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i := 0; i < len(s1); i++ {
		if s1[i] != s2[i] {
			return false
		}
	}

	return true
}
