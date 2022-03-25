package filter

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

const (
	// MasterServiceFilterName filter is used to mutate the ClusterIP and https port of default/kubernetes service
	// in order to pods on edge nodes can access kube-apiserver directly by inClusterConfig.
	MasterServiceFilterName = "masterservice"

	// ServiceTopologyFilterName filter is used to reassemble endpointslice in order to make the service traffic
	// under the topology that defined by service.Annotation["bhojpur.net/topologyKeys"]
	ServiceTopologyFilterName = "servicetopology"

	// DiscardCloudServiceFilterName filter is used to discard cloud service(like loadBalancer service)
	// on kube-proxy list/watch service request from edge nodes.
	DiscardCloudServiceFilterName = "discardcloudservice"

	// SkipDiscardServiceAnnotation is annotation used by LB service.
	// If end users want to use specified LB service at the edge side,
	// End users should add annotation["bhojpur.net/skip-discard"]="true" for LB service.
	SkipDiscardServiceAnnotation = "bhojpur.net/skip-discard"

	// ingresscontroller filter is used to reassemble endpoints in order to make the data traffic be
	// load balanced only to the nodepool valid endpoints.
	IngressControllerFilterName = "ingresscontroller"
)

// DisabledInCloudMode contains the filters that should be disabled when Bhojpur DCP is working in cloud mode.
var DisabledInCloudMode = []string{DiscardCloudServiceFilterName}
