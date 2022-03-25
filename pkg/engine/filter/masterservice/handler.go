package masterservice

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
	"net/http"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/pkg/engine/filter"
	"github.com/bhojpur/dcp/pkg/engine/kubernetes/serializer"
	"github.com/bhojpur/dcp/pkg/engine/util"
)

const (
	MasterServiceNamespace = "default"
	MasterServiceName      = "kubernetes"
	MasterServicePortName  = "https"
)

type masterServiceFilterHandler struct {
	req        *http.Request
	serializer *serializer.Serializer
	host       string
	port       int32
}

func NewMasterServiceFilterHandler(
	req *http.Request,
	serializer *serializer.Serializer,
	host string,
	port int32) filter.Handler {
	return &masterServiceFilterHandler{
		req:        req,
		serializer: serializer,
		host:       host,
		port:       port,
	}
}

// ObjectResponseFilter mutate master service(default/kubernetes) in the ServiceList object
func (fh *masterServiceFilterHandler) ObjectResponseFilter(b []byte) ([]byte, error) {
	list, err := fh.serializer.Decode(b)
	if err != nil || list == nil {
		klog.Errorf("skip filter, failed to decode response in ObjectResponseFilter of masterServiceFilterHandler, %v", err)
		return b, nil
	}

	// return data un-mutated if not ServiceList
	serviceList, ok := list.(*v1.ServiceList)
	if !ok {
		return b, nil
	}

	// mutate master service
	for i := range serviceList.Items {
		if serviceList.Items[i].Namespace == MasterServiceNamespace && serviceList.Items[i].Name == MasterServiceName {
			serviceList.Items[i].Spec.ClusterIP = fh.host
			for j := range serviceList.Items[i].Spec.Ports {
				if serviceList.Items[i].Spec.Ports[j].Name == MasterServicePortName {
					serviceList.Items[i].Spec.Ports[j].Port = fh.port
					break
				}
			}
			klog.V(2).Infof("mutate master service into ClusterIP:Port=%s:%d for request %s", fh.host, fh.port, util.ReqString(fh.req))
			break
		}
	}

	// return the mutated serviceList
	return fh.serializer.Encode(serviceList)
}

//StreamResponseFilter mutate master service(default/kubernetes) in Watch Stream
func (fh *masterServiceFilterHandler) StreamResponseFilter(rc io.ReadCloser, ch chan watch.Event) error {
	defer func() {
		close(ch)
	}()

	d, err := fh.serializer.WatchDecoder(rc)
	if err != nil {
		klog.Errorf("StreamResponseFilter for master service ended with error, %v", err)
		return err
	}

	for {
		watchType, obj, err := d.Decode()
		if err != nil {
			//klog.V(2).Infof("%s %s watch decode ended with: %v", comp, info.Path, err)
			return err
		}

		var wEvent watch.Event
		wEvent.Type = watchType
		// return data un-mutated if not Service
		service, ok := obj.(*v1.Service)
		if ok && service.Namespace == MasterServiceNamespace && service.Name == MasterServiceName {
			service.Spec.ClusterIP = fh.host
			for j := range service.Spec.Ports {
				if service.Spec.Ports[j].Name == MasterServicePortName {
					service.Spec.Ports[j].Port = fh.port
					break
				}
			}
			klog.V(2).Infof("mutate master service into ClusterIP:Port=%s:%d for request %s", fh.host, fh.port, util.ReqString(fh.req))
			wEvent.Object = service
		} else {
			accessor := meta.NewAccessor()
			ns, _ := accessor.Namespace(obj)
			name, _ := accessor.Name(obj)
			kind, _ := accessor.Kind(obj)
			klog.V(2).Infof("skip filter, not master service(%s: %s/%s) for request %s", kind, ns, name, util.ReqString(fh.req))
			wEvent.Object = obj
		}

		ch <- wEvent
	}
}
