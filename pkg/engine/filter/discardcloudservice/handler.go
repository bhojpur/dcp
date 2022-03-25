package discardcloudservice

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

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/pkg/engine/filter"
	"github.com/bhojpur/dcp/pkg/engine/kubernetes/serializer"
)

var (
	cloudClusterIPService = map[string]struct{}{
		"kube-system/x-tunnel-server-internal-svc": {},
	}
)

type discardCloudServiceFilterHandler struct {
	serializer *serializer.Serializer
}

func NewDiscardCloudServiceFilterHandler(serializer *serializer.Serializer) filter.Handler {
	return &discardCloudServiceFilterHandler{
		serializer: serializer,
	}
}

// ObjectResponseFilter remove the cloud service(like LoadBalancer service) from response object
func (fh *discardCloudServiceFilterHandler) ObjectResponseFilter(b []byte) ([]byte, error) {
	list, err := fh.serializer.Decode(b)
	if err != nil || list == nil {
		klog.Errorf("skip filter, failed to decode response in ObjectResponseFilter of discardCloudServiceFilterHandler %v", err)
		return b, nil
	}

	serviceList, ok := list.(*v1.ServiceList)
	if ok {
		var svcNew []v1.Service
		for i := range serviceList.Items {
			nsName := fmt.Sprintf("%s/%s", serviceList.Items[i].Namespace, serviceList.Items[i].Name)
			// remove lb service
			if serviceList.Items[i].Spec.Type == v1.ServiceTypeLoadBalancer {
				if serviceList.Items[i].Annotations[filter.SkipDiscardServiceAnnotation] != "true" {
					klog.V(2).Infof("load balancer service(%s) is discarded in ObjectResponseFilter of discardCloudServiceFilterHandler", nsName)
					continue
				}
			}

			// remove cloud clusterIP service
			if _, ok := cloudClusterIPService[nsName]; ok {
				klog.V(2).Infof("clusterIP service(%s) is discarded in ObjectResponseFilter of discardCloudServiceFilterHandler", nsName)
				continue
			}

			svcNew = append(svcNew, serviceList.Items[i])
		}
		serviceList.Items = svcNew
		return fh.serializer.Encode(serviceList)
	}

	return b, nil
}

// StreamResponseFilter filter the cloud service(like LoadBalancer service) from watch stream response
func (fh *discardCloudServiceFilterHandler) StreamResponseFilter(rc io.ReadCloser, ch chan watch.Event) error {
	defer func() {
		close(ch)
	}()

	d, err := fh.serializer.WatchDecoder(rc)
	if err != nil {
		klog.Errorf("StreamResponseFilter for discardCloudServiceFilterHandler ended with error, %v", err)
		return err
	}

	for {
		watchType, obj, err := d.Decode()
		if err != nil {
			return err
		}

		service, ok := obj.(*v1.Service)
		if ok {
			nsName := fmt.Sprintf("%s/%s", service.Namespace, service.Name)
			// remove cloud LoadBalancer service
			if service.Spec.Type == v1.ServiceTypeLoadBalancer {
				if service.Annotations[filter.SkipDiscardServiceAnnotation] != "true" {
					klog.V(2).Infof("load balancer service(%s) is discarded in StreamResponseFilter of discardCloudServiceFilterHandler", nsName)
					continue
				}
			}

			// remove cloud clusterIP service
			if _, ok := cloudClusterIPService[nsName]; ok {
				klog.V(2).Infof("clusterIP service(%s) is discarded in StreamResponseFilter of discardCloudServiceFilterHandler", nsName)
				continue
			}
		}

		var wEvent watch.Event
		wEvent.Type = watchType
		wEvent.Object = obj
		ch <- wEvent
	}
}
