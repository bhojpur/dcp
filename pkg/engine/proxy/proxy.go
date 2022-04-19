package proxy

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
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/endpoints/filters"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/server"

	"github.com/bhojpur/dcp/cmd/grid/dcpsvr/config"
	"github.com/bhojpur/dcp/pkg/engine/cachemanager"
	"github.com/bhojpur/dcp/pkg/engine/certificate/interfaces"
	"github.com/bhojpur/dcp/pkg/engine/healthchecker"
	"github.com/bhojpur/dcp/pkg/engine/proxy/local"
	"github.com/bhojpur/dcp/pkg/engine/proxy/remote"
	"github.com/bhojpur/dcp/pkg/engine/proxy/util"
	"github.com/bhojpur/dcp/pkg/engine/transport"
	hubutil "github.com/bhojpur/dcp/pkg/engine/util"
)

type dcpReverseProxy struct {
	resolver            apirequest.RequestInfoResolver
	loadBalancer        remote.LoadBalancer
	checker             healthchecker.HealthChecker
	localProxy          *local.LocalProxy
	cacheMgr            cachemanager.CacheManager
	maxRequestsInFlight int
	stopCh              <-chan struct{}
}

// NewReverseProxyHandler creates a http handler for proxying
// all of incoming requests.
func NewReverseProxyHandler(
	engineCfg *config.EngineConfiguration,
	cacheMgr cachemanager.CacheManager,
	transportMgr transport.Interface,
	healthChecker healthchecker.HealthChecker,
	certManager interfaces.EngineCertificateManager,
	stopCh <-chan struct{}) (http.Handler, error) {
	cfg := &server.Config{
		LegacyAPIGroupPrefixes: sets.NewString(server.DefaultLegacyAPIPrefix),
	}
	resolver := server.NewRequestInfoResolver(cfg)

	lb, err := remote.NewLoadBalancer(
		engineCfg.LBMode,
		engineCfg.RemoteServers,
		cacheMgr,
		transportMgr,
		healthChecker,
		certManager,
		engineCfg.FilterChain,
		stopCh)
	if err != nil {
		return nil, err
	}

	var localProxy *local.LocalProxy
	// When Bhojpur DCP is working in cloud mode, cacheMgr will be set to nil which means the local cache is disabled,
	// so we don't need to create a LocalProxy.
	if cacheMgr != nil {
		localProxy = local.NewLocalProxy(cacheMgr, lb.IsHealthy)
	}

	dcpProxy := &dcpReverseProxy{
		resolver:            resolver,
		loadBalancer:        lb,
		checker:             healthChecker,
		localProxy:          localProxy,
		cacheMgr:            cacheMgr,
		maxRequestsInFlight: engineCfg.MaxRequestInFlight,
		stopCh:              stopCh,
	}

	return dcpProxy.buildHandlerChain(dcpProxy), nil
}

func (p *dcpReverseProxy) buildHandlerChain(handler http.Handler) http.Handler {
	handler = util.WithRequestTrace(handler)
	handler = util.WithRequestContentType(handler)
	if p.cacheMgr != nil {
		handler = util.WithCacheHeaderCheck(handler)
	}
	handler = util.WithRequestTimeout(handler)
	if p.cacheMgr != nil {
		handler = util.WithListRequestSelector(handler)
	}
	handler = util.WithMaxInFlightLimit(handler, p.maxRequestsInFlight)
	handler = util.WithRequestClientComponent(handler)
	handler = filters.WithRequestInfo(handler, p.resolver)
	return handler
}

func (p *dcpReverseProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	isKubeletLeaseReq := hubutil.IsKubeletLeaseReq(req)
	if !isKubeletLeaseReq && p.loadBalancer.IsHealthy() || p.localProxy == nil {
		p.loadBalancer.ServeHTTP(rw, req)
	} else {
		if isKubeletLeaseReq {
			p.checker.UpdateLastKubeletLeaseReqTime(time.Now())
		}
		p.localProxy.ServeHTTP(rw, req)
	}
}
