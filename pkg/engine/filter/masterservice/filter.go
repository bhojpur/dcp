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
	"net"
	"net/http"
	"strconv"

	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/pkg/engine/filter"
	filterutil "github.com/bhojpur/dcp/pkg/engine/filter/util"
	"github.com/bhojpur/dcp/pkg/engine/kubernetes/serializer"
)

// Register registers a filter
func Register(filters *filter.Filters) {
	filters.Register(filter.MasterServiceFilterName, func() (filter.Interface, error) {
		return NewFilter(), nil
	})
}

func NewFilter() *masterServiceFilter {
	return &masterServiceFilter{
		Approver: filter.NewApprover("kubelet", "services", []string{"list", "watch"}...),
		stopCh:   make(chan struct{}),
	}
}

type masterServiceFilter struct {
	*filter.Approver
	serializerManager *serializer.SerializerManager
	host              string
	port              int32
	stopCh            chan struct{}
}

func (msf *masterServiceFilter) SetSerializerManager(s *serializer.SerializerManager) error {
	msf.serializerManager = s
	return nil
}

func (msf *masterServiceFilter) SetMasterServiceAddr(addr string) error {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	msf.host = host
	port, err := strconv.ParseInt(portStr, 10, 32)
	if err != nil {
		return err
	}
	msf.port = int32(port)
	return nil
}

func (msf *masterServiceFilter) Filter(req *http.Request, rc io.ReadCloser, stopCh <-chan struct{}) (int, io.ReadCloser, error) {
	s := filterutil.CreateSerializer(req, msf.serializerManager)
	if s == nil {
		klog.Errorf("skip filter, failed to create serializer in masterServiceFilter")
		return 0, rc, nil
	}

	handler := NewMasterServiceFilterHandler(req, s, msf.host, msf.port)
	return filter.NewFilterReadCloser(req, rc, handler, s, filter.MasterServiceFilterName, stopCh)
}
