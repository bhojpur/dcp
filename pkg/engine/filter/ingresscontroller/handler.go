package ingresscontroller

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
	"io"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/pkg/engine/filter"
	"github.com/bhojpur/dcp/pkg/engine/kubernetes/serializer"
	nodepoolv1alpha1 "github.com/bhojpur/dcp/pkg/appmanager/apis/apps/v1alpha1"
	appslisters "github.com/bhojpur/dcp/pkg/appmanager/client/listers/apps/v1alpha1"
)

type ingressControllerFilterHandler struct {
	nodeName       string
	serializer     *serializer.Serializer
	serviceLister  listers.ServiceLister
	nodePoolLister appslisters.NodePoolLister
	nodeGetter     filter.NodeGetter
}

func NewIngressControllerFilterHandler(
	nodeName string,
	serializer *serializer.Serializer,
	serviceLister listers.ServiceLister,
	nodePoolLister appslisters.NodePoolLister,
	nodeGetter filter.NodeGetter) filter.Handler {
	return &ingressControllerFilterHandler{
		nodeName:       nodeName,
		serializer:     serializer,
		serviceLister:  serviceLister,
		nodePoolLister: nodePoolLister,
		nodeGetter:     nodeGetter,
	}
}

// ObjectResponseFilter filter the endpoints from get response object and return the bytes
func (fh *ingressControllerFilterHandler) ObjectResponseFilter(b []byte) ([]byte, error) {
	eps, err := fh.serializer.Decode(b)
	if err != nil || eps == nil {
		klog.Errorf("skip filter, failed to decode response in ObjectResponseFilter of ingressControllerFilterHandler, %v", err)
		return b, nil
	}

	endpointsList, ok := eps.(*v1.EndpointsList)
	if !ok {
		return b, nil
	}
	// filter endpoints
	var items []v1.Endpoints
	for i := range endpointsList.Items {
		item := fh.reassembleEndpoint(&endpointsList.Items[i])
		if item != nil {
			items = append(items, *item)
		}
	}
	endpointsList.Items = items

	return fh.serializer.Encode(endpointsList)
}

// FilterWatchObject filter the endpoints from watch response object and return the bytes
func (fh *ingressControllerFilterHandler) StreamResponseFilter(rc io.ReadCloser, ch chan watch.Event) error {
	defer func() {
		close(ch)
	}()

	d, err := fh.serializer.WatchDecoder(rc)
	if err != nil {
		klog.Errorf("StreamResponseFilter of ingressControllerFilterHandler ended with error, %v", err)
		return err
	}
	for {
		watchType, obj, err := d.Decode()
		if err != nil {
			return err
		}
		var wEvent watch.Event
		wEvent.Type = watchType
		endpoints, ok := obj.(*v1.Endpoints)
		if ok {
			item := fh.reassembleEndpoint(endpoints)
			if item == nil {
				continue
			}
			wEvent.Object = item
		} else {
			wEvent.Object = obj
		}
		klog.V(5).Infof("filter watch decode endpoint: type: %s, obj=%#+v", watchType, endpoints)
		ch <- wEvent
	}
}

// reassembleEndpoints will filter the valid endpoints to its nodepool
func (fh *ingressControllerFilterHandler) reassembleEndpoint(endpoints *v1.Endpoints) *v1.Endpoints {
	svcName := endpoints.Name
	_, err := fh.serviceLister.Services(endpoints.Namespace).Get(svcName)
	if err != nil {
		klog.Infof("skip reassemble endpoints, failed to get service %s/%s, err: %v", endpoints.Namespace, svcName, err)
		return endpoints
	}
	// filter the endpoints on the node which is in the same nodepool with current node
	currentNode, err := fh.nodeGetter(fh.nodeName)
	if err != nil {
		klog.Infof("skip reassemble endpoints, failed to get current node %s, err: %v", fh.nodeName, err)
		return endpoints
	}
	if nodePoolName, ok := currentNode.Labels[nodepoolv1alpha1.LabelCurrentNodePool]; ok {
		nodePool, err := fh.nodePoolLister.Get(nodePoolName)
		if err != nil {
			klog.Infof("skip reassemble endpoints, failed to get nodepool %s, err: %v", nodePoolName, err)
			return endpoints
		}
		var newEpSubsets []v1.EndpointSubset
		for i := range endpoints.Subsets {
			endpoints.Subsets[i].Addresses = filterValidEndpointsAddr(endpoints.Subsets[i].Addresses, nodePool)
			endpoints.Subsets[i].NotReadyAddresses = filterValidEndpointsAddr(endpoints.Subsets[i].NotReadyAddresses, nodePool)
			if endpoints.Subsets[i].Addresses != nil || endpoints.Subsets[i].NotReadyAddresses != nil {
				newEpSubsets = append(newEpSubsets, endpoints.Subsets[i])
			}
		}
		endpoints.Subsets = newEpSubsets
		if len(endpoints.Subsets) == 0 {
			// this endpoints has no nodepool valid addresses for ingress controller, return nil to ignore it
			return nil
		}
	}
	return endpoints
}

func filterValidEndpointsAddr(addresses []v1.EndpointAddress, nodePool *nodepoolv1alpha1.NodePool) []v1.EndpointAddress {
	var newEpAddresses []v1.EndpointAddress
	for i := range addresses {
		nodeName := addresses[i].NodeName
		if nodeName == nil {
			// ignore endpoints whose NodeName is not set, for example "kubernetes"
			continue
		}
		if inSameNodePool(*nodeName, nodePool.Status.Nodes) {
			newEpAddresses = append(newEpAddresses, addresses[i])
			klog.Infof("endpoints address/%s with nodeName/%s is valid to nodepool/%s", addresses[i].IP, *nodeName, nodePool.Name)
		}
	}
	return newEpAddresses
}

func inSameNodePool(nodeName string, nodeList []string) bool {
	for _, n := range nodeList {
		if nodeName == n {
			return true
		}
	}
	return false
}
