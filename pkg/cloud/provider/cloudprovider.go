package provider

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

	"github.com/bhojpur/dcp/pkg/cloud/version"
	"k8s.io/client-go/informers"
	informercorev1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"
	cloudprovider "k8s.io/cloud-provider"
)

type dcp struct {
	nodeInformer          informercorev1.NodeInformer
	nodeInformerHasSynced cache.InformerSynced
}

var _ cloudprovider.Interface = &dcp{}
var _ cloudprovider.InformerUser = &dcp{}

func init() {
	cloudprovider.RegisterCloudProvider(version.Program, func(config io.Reader) (cloudprovider.Interface, error) {
		return &dcp{}, nil
	})
}

func (k *dcp) Initialize(clientBuilder cloudprovider.ControllerClientBuilder, stop <-chan struct{}) {
}

func (k *dcp) SetInformers(informerFactory informers.SharedInformerFactory) {
	k.nodeInformer = informerFactory.Core().V1().Nodes()
	k.nodeInformerHasSynced = k.nodeInformer.Informer().HasSynced
}

func (k *dcp) Instances() (cloudprovider.Instances, bool) {
	return k, true
}

func (k *dcp) InstancesV2() (cloudprovider.InstancesV2, bool) {
	return nil, false
}

func (k *dcp) LoadBalancer() (cloudprovider.LoadBalancer, bool) {
	return nil, false
}

func (k *dcp) Zones() (cloudprovider.Zones, bool) {
	return nil, false
}

func (k *dcp) Clusters() (cloudprovider.Clusters, bool) {
	return nil, false
}

func (k *dcp) Routes() (cloudprovider.Routes, bool) {
	return nil, false
}

func (k *dcp) ProviderName() string {
	return version.Program
}

func (k *dcp) HasClusterID() bool {
	return true
}
