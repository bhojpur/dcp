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
	"fmt"
	"io"
	"net/http"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	dcpinformers "github.com/bhojpur/dcp/pkg/appmanager/client/informers/externalversions"
	appslisters "github.com/bhojpur/dcp/pkg/appmanager/client/listers/apps/v1alpha1"
	"github.com/bhojpur/dcp/pkg/engine/cachemanager"
	"github.com/bhojpur/dcp/pkg/engine/filter"
	filterutil "github.com/bhojpur/dcp/pkg/engine/filter/util"
	"github.com/bhojpur/dcp/pkg/engine/kubernetes/serializer"
	"github.com/bhojpur/dcp/pkg/engine/util"
)

// Register registers a filter
func Register(filters *filter.Filters) {
	filters.Register(filter.ServiceTopologyFilterName, func() (filter.Interface, error) {
		return NewFilter(), nil
	})
}

func NewFilter() *serviceTopologyFilter {
	return &serviceTopologyFilter{
		Approver:    filter.NewApprover("kube-proxy", "endpointslices", []string{"list", "watch"}...),
		workingMode: util.WorkingModeEdge,
		stopCh:      make(chan struct{}),
	}
}

type serviceTopologyFilter struct {
	*filter.Approver
	serviceLister     listers.ServiceLister
	serviceSynced     cache.InformerSynced
	nodepoolLister    appslisters.NodePoolLister
	nodePoolSynced    cache.InformerSynced
	nodeGetter        filter.NodeGetter
	nodeSynced        cache.InformerSynced
	nodeName          string
	workingMode       util.WorkingMode
	serializerManager *serializer.SerializerManager
	stopCh            chan struct{}
}

func (ssf *serviceTopologyFilter) SetWorkingMode(mode util.WorkingMode) error {
	ssf.workingMode = mode
	return nil
}

func (ssf *serviceTopologyFilter) SetSharedInformerFactory(factory informers.SharedInformerFactory) error {
	ssf.serviceLister = factory.Core().V1().Services().Lister()
	ssf.serviceSynced = factory.Core().V1().Services().Informer().HasSynced

	if ssf.workingMode == util.WorkingModeCloud {
		klog.Infof("prepare list/watch to sync node(%s) for cloud working mode", ssf.nodeName)
		ssf.nodeSynced = factory.Core().V1().Nodes().Informer().HasSynced
		ssf.nodeGetter = factory.Core().V1().Nodes().Lister().Get
	}

	return nil
}

func (ssf *serviceTopologyFilter) SetDcpSharedInformerFactory(dcpFactory dcpinformers.SharedInformerFactory) error {
	ssf.nodepoolLister = dcpFactory.Apps().V1alpha1().NodePools().Lister()
	ssf.nodePoolSynced = dcpFactory.Apps().V1alpha1().NodePools().Informer().HasSynced

	return nil
}

func (ssf *serviceTopologyFilter) SetNodeName(nodeName string) error {
	ssf.nodeName = nodeName

	return nil
}

func (ssf *serviceTopologyFilter) SetStorageWrapper(s cachemanager.StorageWrapper) error {
	if len(ssf.nodeName) == 0 {
		return fmt.Errorf("node name for serviceTopologyFilter is not ready")
	}

	// hub agent will list/watch node from kube-apiserver when hub agent work as cloud mode
	if ssf.workingMode == util.WorkingModeCloud {
		return nil
	}
	klog.Infof("prepare local disk storage to sync node(%s) for edge working mode", ssf.nodeName)

	nodeKey := fmt.Sprintf("kubelet/nodes/%s", ssf.nodeName)
	ssf.nodeSynced = func() bool {
		obj, err := s.Get(nodeKey)
		if err != nil || obj == nil {
			return false
		}

		if _, ok := obj.(*v1.Node); !ok {
			return false
		}

		return true
	}

	ssf.nodeGetter = func(name string) (*v1.Node, error) {
		obj, err := s.Get(fmt.Sprintf("kubelet/nodes/%s", name))
		if err != nil {
			return nil, err
		} else if obj == nil {
			return nil, fmt.Errorf("node(%s) is not ready", name)
		}

		if node, ok := obj.(*v1.Node); ok {
			return node, nil
		}

		return nil, fmt.Errorf("node(%s) is not found", name)
	}

	return nil
}

func (ssf *serviceTopologyFilter) SetSerializerManager(s *serializer.SerializerManager) error {
	ssf.serializerManager = s
	return nil
}

func (ssf *serviceTopologyFilter) Approve(comp, resource, verb string) bool {
	if !ssf.Approver.Approve(comp, resource, verb) {
		return false
	}

	if ok := cache.WaitForCacheSync(ssf.stopCh, ssf.nodeSynced, ssf.serviceSynced, ssf.nodePoolSynced); !ok {
		return false
	}

	return true
}

func (ssf *serviceTopologyFilter) Filter(req *http.Request, rc io.ReadCloser, stopCh <-chan struct{}) (int, io.ReadCloser, error) {
	s := filterutil.CreateSerializer(req, ssf.serializerManager)
	if s == nil {
		klog.Errorf("skip filter, failed to create serializer in serviceTopologyFilter")
		return 0, rc, nil
	}

	handler := NewServiceTopologyFilterHandler(ssf.nodeName, s, ssf.serviceLister, ssf.nodepoolLister, ssf.nodeGetter)
	return filter.NewFilterReadCloser(req, rc, handler, s, filter.ServiceTopologyFilterName, stopCh)
}
