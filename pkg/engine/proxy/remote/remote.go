package remote

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
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/apimachinery/pkg/util/proxy"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/pkg/engine/cachemanager"
	"github.com/bhojpur/dcp/pkg/engine/filter"
	"github.com/bhojpur/dcp/pkg/engine/healthchecker"
	"github.com/bhojpur/dcp/pkg/engine/transport"
	"github.com/bhojpur/dcp/pkg/engine/util"
)

// RemoteProxy is an reverse proxy for remote server
type RemoteProxy struct {
	checker              healthchecker.HealthChecker
	reverseProxy         *httputil.ReverseProxy
	cacheMgr             cachemanager.CacheManager
	remoteServer         *url.URL
	filterChain          filter.Interface
	currentTransport     http.RoundTripper
	bearerTransport      http.RoundTripper
	upgradeHandler       *proxy.UpgradeAwareHandler
	bearerUpgradeHandler *proxy.UpgradeAwareHandler
	stopCh               <-chan struct{}
}

type responder struct{}

func (r *responder) Error(w http.ResponseWriter, req *http.Request, err error) {
	klog.Errorf("failed while proxying request %s, %v", req.URL, err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// NewRemoteProxy creates an *RemoteProxy object, and will be used by LoadBalancer
func NewRemoteProxy(remoteServer *url.URL,
	cacheMgr cachemanager.CacheManager,
	transportMgr transport.Interface,
	healthChecker healthchecker.HealthChecker,
	filterChain filter.Interface,
	stopCh <-chan struct{}) (*RemoteProxy, error) {
	currentTransport := transportMgr.CurrentTransport()
	if currentTransport == nil {
		return nil, fmt.Errorf("could not get current transport when init proxy backend(%s)", remoteServer.String())
	}
	bearerTransport := transportMgr.BearerTransport()
	if bearerTransport == nil {
		return nil, fmt.Errorf("could not get bearer transport when init proxy backend(%s)", remoteServer.String())
	}

	upgradeAwareHandler := proxy.NewUpgradeAwareHandler(remoteServer, currentTransport, false, true, &responder{})
	upgradeAwareHandler.UseRequestLocation = true
	bearerUpgradeAwareHandler := proxy.NewUpgradeAwareHandler(remoteServer, bearerTransport, false, true, &responder{})
	bearerUpgradeAwareHandler.UseRequestLocation = true

	proxyBackend := &RemoteProxy{
		checker:              healthChecker,
		reverseProxy:         httputil.NewSingleHostReverseProxy(remoteServer),
		cacheMgr:             cacheMgr,
		remoteServer:         remoteServer,
		filterChain:          filterChain,
		currentTransport:     currentTransport,
		bearerTransport:      bearerTransport,
		upgradeHandler:       upgradeAwareHandler,
		bearerUpgradeHandler: bearerUpgradeAwareHandler,
		stopCh:               stopCh,
	}

	proxyBackend.reverseProxy.Transport = proxyBackend
	proxyBackend.reverseProxy.ModifyResponse = proxyBackend.modifyResponse
	proxyBackend.reverseProxy.FlushInterval = -1
	proxyBackend.reverseProxy.ErrorHandler = proxyBackend.errorHandler

	return proxyBackend, nil
}

// Name represents the address of remote server
func (rp *RemoteProxy) Name() string {
	return rp.remoteServer.String()
}

func (rp *RemoteProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if httpstream.IsUpgradeRequest(req) {
		klog.V(5).Infof("get upgrade request %s", req.URL)
		if isBearerRequest(req) {
			rp.bearerUpgradeHandler.ServeHTTP(rw, req)
		} else {
			rp.upgradeHandler.ServeHTTP(rw, req)
		}
		return
	}

	rp.reverseProxy.ServeHTTP(rw, req)
}

// IsHealthy returns healthy status of remote server
func (rp *RemoteProxy) IsHealthy() bool {
	return rp.checker.IsHealthy(rp.remoteServer)
}

func (rp *RemoteProxy) modifyResponse(resp *http.Response) error {
	if resp == nil || resp.Request == nil {
		klog.Infof("no request info in response, skip cache response")
		return nil
	}

	req := resp.Request
	ctx := req.Context()

	// re-added transfer-encoding=chunked response header for watch request
	info, exists := apirequest.RequestInfoFrom(ctx)
	if exists {
		if info.Verb == "watch" {
			klog.V(5).Infof("add transfer-encoding=chunked header into response for req %s", util.ReqString(req))
			h := resp.Header
			if hv := h.Get("Transfer-Encoding"); hv == "" {
				h.Add("Transfer-Encoding", "chunked")
			}
		}
	}

	if resp.StatusCode >= http.StatusOK && resp.StatusCode <= http.StatusPartialContent {
		// prepare response content type
		reqContentType, _ := util.ReqContentTypeFrom(ctx)
		respContentType := resp.Header.Get("Content-Type")
		if len(respContentType) == 0 {
			respContentType = reqContentType
		}
		ctx = util.WithRespContentType(ctx, respContentType)
		req = req.WithContext(ctx)

		// filter response data
		if rp.filterChain != nil {
			size, filterRc, err := rp.filterChain.Filter(req, resp.Body, rp.stopCh)
			if err != nil {
				klog.Errorf("failed to filter response for %s, %v", util.ReqString(req), err)
				return err
			}
			resp.Body = filterRc
			if size > 0 {
				resp.ContentLength = int64(size)
				resp.Header.Set("Content-Length", fmt.Sprint(size))
			}
		}

		// cache resp with storage interface
		if rp.cacheMgr != nil && rp.cacheMgr.CanCacheFor(req) {
			rc, prc := util.NewDualReadCloser(req, resp.Body, true)
			go func(req *http.Request, prc io.ReadCloser, stopCh <-chan struct{}) {
				err := rp.cacheMgr.CacheResponse(req, prc, stopCh)
				if err != nil && err != io.EOF && err != context.Canceled {
					klog.Errorf("%s response cache ended with error, %v", util.ReqString(req), err)
				}
			}(req, prc, rp.stopCh)

			resp.Body = rc
		}
	} else if resp.StatusCode == http.StatusNotFound && info.Verb == "list" && rp.cacheMgr != nil {
		// 404 Not Found: The CRD may have been unregistered and should be updated locally as well.
		// Other types of requests may return a 404 response for other reasons (for example, getting a pod that doesn't exist).
		// And the main purpose is to return 404 when list an unregistered resource locally, so here only consider the list request.
		gvr := schema.GroupVersionResource{
			Group:    info.APIGroup,
			Version:  info.APIVersion,
			Resource: info.Resource,
		}

		err := rp.cacheMgr.DeleteKindFor(gvr)
		if err != nil {
			klog.Errorf("failed: %v", err)
		}
	}
	return nil
}

func (rp *RemoteProxy) errorHandler(rw http.ResponseWriter, req *http.Request, err error) {
	klog.Errorf("remote proxy error handler: %s, %v", util.ReqString(req), err)
	if rp.cacheMgr == nil || !rp.cacheMgr.CanCacheFor(req) {
		rw.WriteHeader(http.StatusBadGateway)
		return
	}

	ctx := req.Context()
	if info, ok := apirequest.RequestInfoFrom(ctx); ok {
		if info.Verb == "get" || info.Verb == "list" {
			if obj, err := rp.cacheMgr.QueryCache(req); err == nil {
				util.WriteObject(http.StatusOK, obj, rw, req)
				return
			}
		}
	}
	rw.WriteHeader(http.StatusBadGateway)
}

// RoundTrip is used to implement http.RoundTripper for RemoteProxy.
func (rp *RemoteProxy) RoundTrip(req *http.Request) (*http.Response, error) {
	// when edge client(like kube-proxy, flannel, etc) use service account(default InClusterConfig) to access Bhojpur DCP,
	// Authorization header will be set in request. and when edge client(like kubelet) use x509 certificate to access
	// Bhojpur DCP engine, Authorization header in request will be empty.
	if isBearerRequest(req) {
		return rp.bearerTransport.RoundTrip(req)
	}

	return rp.currentTransport.RoundTrip(req)
}

func isBearerRequest(req *http.Request) bool {
	auth := strings.TrimSpace(req.Header.Get("Authorization"))
	if auth != "" {
		parts := strings.Split(auth, " ")
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			klog.V(5).Infof("request: %s with bearer token: %s", util.ReqString(req), parts[1])
			return true
		}
	}
	return false
}
