package wraphandler

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
	"net/http"

	"k8s.io/klog/v2"

	hw "github.com/bhojpur/dcp/pkg/tunnel/handlerwrapper"
	"github.com/bhojpur/dcp/pkg/tunnel/handlerwrapper/initializer"
	"github.com/bhojpur/dcp/pkg/tunnel/handlerwrapper/localhostproxy"
	"github.com/bhojpur/dcp/pkg/tunnel/handlerwrapper/tracerequest"
)

func InitHandlerWrappers(mi initializer.MiddlewareInitializer) (hw.HandlerWrappers, error) {
	wrappers := make(hw.HandlerWrappers, 0)
	// register all of middleware here
	//
	// NOTE the register order decide the order in which
	// the middleware will be called
	//
	// e.g. there are two middleware mw1 and mw2
	// if the middlewares are registered in the following order,
	//
	// wrappers = append(wrappers, m2)
	// wrappers = append(wrappers, m1)
	//
	// then the middleware m2 will be called before the mw1
	wrappers = append(wrappers, tracerequest.NewTraceReqMiddleware())
	wrappers = append(wrappers, localhostproxy.NewLocalHostProxyMiddleware())

	// init all of wrappers
	for i := range wrappers {
		if err := mi.Initialize(wrappers[i]); err != nil {
			return wrappers, err
		}
	}

	return wrappers, nil
}

// WrapWrapHandler wraps the coreHandler with all of registered middleware
// and middleware will be initialized before wrap.
func WrapHandler(coreHandler http.Handler, wrappers hw.HandlerWrappers) (http.Handler, error) {
	handler := coreHandler
	klog.V(4).Infof("%d middlewares will be added into wrap handler", len(wrappers))
	if len(wrappers) == 0 {
		return handler, nil
	}
	for i := len(wrappers) - 1; i >= 0; i-- {
		handler = wrappers[i].WrapHandler(handler)
		klog.V(2).Infof("add %s into wrap handler", wrappers[i].Name())
	}
	return handler, nil
}
