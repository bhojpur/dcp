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

	apirequest "k8s.io/apiserver/pkg/endpoints/request"

	"github.com/bhojpur/dcp/pkg/engine/util"
)

type filterChain []Interface

func (fc filterChain) Approve(comp, resource, verb string) bool {
	for _, f := range fc {
		if f.Approve(comp, resource, verb) {
			return true
		}
	}

	return false
}

func (fc filterChain) Filter(req *http.Request, rc io.ReadCloser, stopCh <-chan struct{}) (int, io.ReadCloser, error) {
	ctx := req.Context()
	comp, ok := util.ClientComponentFrom(ctx)
	if !ok {
		return 0, rc, nil
	}

	info, ok := apirequest.RequestInfoFrom(ctx)
	if !ok {
		return 0, rc, nil
	}

	for _, f := range fc {
		if !f.Approve(comp, info.Resource, info.Verb) {
			continue
		}

		return f.Filter(req, rc, stopCh)
	}

	return 0, rc, nil
}
