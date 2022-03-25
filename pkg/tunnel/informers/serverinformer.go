package informers

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
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/bhojpur/dcp/pkg/projectinfo"
	"github.com/bhojpur/dcp/pkg/tunnel/constants"
	"github.com/bhojpur/dcp/pkg/tunnel/util"
)

// RegisterInformersForTunnelServer registers shared informers that tunnel server use.
func RegisterInformersForTunnelServer(informerFactory informers.SharedInformerFactory) {
	// add node informers
	informerFactory.Core().V1().Nodes()

	// add service informers
	informerFactory.InformerFor(&corev1.Service{}, newServiceInformer)

	// add configMap informers
	informerFactory.InformerFor(&corev1.ConfigMap{}, newConfigMapInformer)

	// add endpoints informers
	informerFactory.InformerFor(&corev1.Endpoints{}, newEndPointsInformer)
}

// newServiceInformer creates a shared index informers that returns services related to tunnel
func newServiceInformer(cs clientset.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	// this informers will be used by coreDNSRecordController and certificate manager,
	// so it should return x-tunnel-server-svc and x-tunnel-server-internal-svc
	selector := fmt.Sprintf("name=%v", projectinfo.TunnelServerLabel())
	tweakListOptions := func(options *metav1.ListOptions) {
		options.LabelSelector = selector
	}
	return coreinformers.NewFilteredServiceInformer(cs, constants.TunnelServerServiceNs, resyncPeriod, nil, tweakListOptions)
}

// newConfigMapInformer creates a shared index informers that returns only interested configmaps
func newConfigMapInformer(cs clientset.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	selector := fmt.Sprintf("metadata.name=%v", util.TunnelServerDnatConfigMapName)
	tweakListOptions := func(options *metav1.ListOptions) {
		options.FieldSelector = selector
	}
	return coreinformers.NewFilteredConfigMapInformer(cs, util.TunnelServerDnatConfigMapNs, resyncPeriod, nil, tweakListOptions)
}

// newEndPointsInformer creates a shared index informers that returns only interested endpoints
func newEndPointsInformer(cs clientset.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	selector := fmt.Sprintf("metadata.name=%v", constants.TunnelEndpointsName)
	tweakListOptions := func(options *metav1.ListOptions) {
		options.FieldSelector = selector
	}
	return coreinformers.NewFilteredEndpointsInformer(cs, constants.TunnelEndpointsNs, resyncPeriod, nil, tweakListOptions)
}
