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

import (
	"io"
	"net/http"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type FilterInitializer interface {
	Initialize(filter Interface) error
}

// Interface of data filtering framework.
type Interface interface {
	// Approve is used to determine whether the data returned
	// from the cloud needs to enter the filtering framework for processing.
	Approve(comp, resource, verb string) bool

	// Filter is used to filter data returned from the cloud.
	Filter(req *http.Request, rc io.ReadCloser, stopCh <-chan struct{}) (int, io.ReadCloser, error)
}

// Handler customizes data filtering processing interface for each handler.
// In the data filtering framework, data is mainly divided into two types:
// 	Object data: data returned by list/get request.
// 	Streaming data: The data returned by the watch request will be continuously pushed to the edge by the cloud.
type Handler interface {
	// StreamResponseFilter is used to filter processing of streaming data.
	StreamResponseFilter(rc io.ReadCloser, ch chan watch.Event) error

	// ObjectResponseFilter is used to filter processing of object data.
	ObjectResponseFilter(b []byte) ([]byte, error)
}

type NodeGetter func(name string) (*v1.Node, error)
