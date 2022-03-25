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
	"io"

	v1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1beta1"
	"k8s.io/apimachinery/pkg/watch"
	listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/pkg/engine/filter"
	"github.com/bhojpur/dcp/pkg/engine/kubernetes/serializer"
	nodepoolv1alpha1 "github.com/bhojpur/dcp/pkg/appmanager/apis/apps/v1alpha1"
	appslisters "github.com/bhojpur/dcp/pkg/appmanager/client/listers/apps/v1alpha1"
)

const (
	AnnotationServiceTopologyKey           = "bhojpur.net/topologyKeys"
	AnnotationServiceTopologyValueNode     = "kubernetes.io/hostname"
	AnnotationServiceTopologyValueZone     = "kubernetes.io/zone"
	AnnotationServiceTopologyValueNodePool = "bhojpur.net/nodepool"
)

type serviceTopologyFilterHandler struct {
	nodeName       string
	serializer     *serializer.Serializer
	serviceLister  listers.ServiceLister
	nodePoolLister appslisters.NodePoolLister
	nodeGetter     filter.NodeGetter
}

func NewServiceTopologyFilterHandler(
	nodeName string,
	serializer *serializer.Serializer,
	serviceLister listers.ServiceLister,
	nodePoolLister appslisters.NodePoolLister,
	nodeGetter filter.NodeGetter) filter.Handler {
	return &serviceTopologyFilterHandler{
		nodeName:       nodeName,
		serializer:     serializer,
		serviceLister:  serviceLister,
		nodePoolLister: nodePoolLister,
		nodeGetter:     nodeGetter,
	}

}

//ObjectResponseFilter filter the endpointslice from get response object and return the bytes
func (fh *serviceTopologyFilterHandler) ObjectResponseFilter(b []byte) ([]byte, error) {
	eps, err := fh.serializer.Decode(b)
	if err != nil || eps == nil {
		klog.Errorf("skip filter, failed to decode response in ObjectResponseFilter of serviceTopologyFilterHandler, %v", err)
		return b, nil
	}

	if endpointSliceList, ok := eps.(*discovery.EndpointSliceList); ok {
		//filter endpointSlice
		var items []discovery.EndpointSlice
		for i := range endpointSliceList.Items {
			item := fh.reassembleEndpointSlice(&endpointSliceList.Items[i])
			if item != nil {
				items = append(items, *item)
			}
		}
		endpointSliceList.Items = items

		return fh.serializer.Encode(endpointSliceList)
	} else {
		return b, nil
	}
}

//FilterWatchObject filter the endpointslice from watch response object and return the bytes
func (fh *serviceTopologyFilterHandler) StreamResponseFilter(rc io.ReadCloser, ch chan watch.Event) error {
	defer func() {
		close(ch)
	}()

	d, err := fh.serializer.WatchDecoder(rc)
	if err != nil {
		klog.Errorf("StreamResponseFilter of serviceTopologyFilterHandler ended with error, %v", err)
		return err
	}

	for {
		watchType, obj, err := d.Decode()
		if err != nil {
			return err
		}

		var wEvent watch.Event
		wEvent.Type = watchType
		endpointSlice, ok := obj.(*discovery.EndpointSlice)
		if ok {
			item := fh.reassembleEndpointSlice(endpointSlice)
			if item == nil {
				continue
			}
			wEvent.Object = item
		} else {
			wEvent.Object = obj
		}

		klog.V(5).Infof("filter watch decode endpointSlice: type: %s, obj=%#+v", watchType, endpointSlice)
		ch <- wEvent
	}
}

// reassembleEndpointSlice will discard endpointslice for LB service and filter the endpoints out of endpointslice in terms of service Topology.
func (fh *serviceTopologyFilterHandler) reassembleEndpointSlice(endpointSlice *discovery.EndpointSlice) *discovery.EndpointSlice {
	var serviceTopologyType string
	// get the service Topology type
	if svcName, ok := endpointSlice.Labels[discovery.LabelServiceName]; ok {
		svc, err := fh.serviceLister.Services(endpointSlice.Namespace).Get(svcName)
		if err != nil {
			klog.Infof("skip reassemble endpointSlice, failed to get service %s/%s, err: %v", endpointSlice.Namespace, svcName, err)
			return endpointSlice
		}

		if serviceTopologyType, ok = svc.Annotations[AnnotationServiceTopologyKey]; !ok {
			klog.Infof("skip reassemble endpointSlice, service %s/%s has no annotation %s", endpointSlice.Namespace, svcName, AnnotationServiceTopologyKey)
			return endpointSlice
		}
	}

	var newEps []discovery.Endpoint
	// if type of service Topology is 'kubernetes.io/hostname'
	// filter the endpoint just on the local host
	if serviceTopologyType == AnnotationServiceTopologyValueNode {
		for i := range endpointSlice.Endpoints {
			if endpointSlice.Endpoints[i].Topology[v1.LabelHostname] == fh.nodeName {
				newEps = append(newEps, endpointSlice.Endpoints[i])
			}
		}
		endpointSlice.Endpoints = newEps
	} else if serviceTopologyType == AnnotationServiceTopologyValueNodePool || serviceTopologyType == AnnotationServiceTopologyValueZone {
		// if type of service Topology is bhojpur.net/nodepool
		// filter the endpoint just on the node which is in the same nodepool with current node
		currentNode, err := fh.nodeGetter(fh.nodeName)
		if err != nil {
			klog.Infof("skip reassemble endpointSlice, failed to get current node %s, err: %v", fh.nodeName, err)
			return endpointSlice
		}
		if nodePoolName, ok := currentNode.Labels[nodepoolv1alpha1.LabelCurrentNodePool]; ok {
			nodePool, err := fh.nodePoolLister.Get(nodePoolName)
			if err != nil {
				klog.Infof("skip reassemble endpointSlice, failed to get nodepool %s, err: %v", nodePoolName, err)
				return endpointSlice
			}
			for i := range endpointSlice.Endpoints {
				if inSameNodePool(endpointSlice.Endpoints[i].Topology[v1.LabelHostname], nodePool.Status.Nodes) {
					newEps = append(newEps, endpointSlice.Endpoints[i])
				}
			}
			endpointSlice.Endpoints = newEps
		}
	}
	return endpointSlice
}

func inSameNodePool(nodeName string, nodeList []string) bool {
	for _, n := range nodeList {
		if nodeName == n {
			return true
		}
	}
	return false
}
