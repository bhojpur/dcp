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
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/pkg/engine/cachemanager"
	"github.com/bhojpur/dcp/pkg/engine/certificate/interfaces"
	"github.com/bhojpur/dcp/pkg/engine/filter"
	"github.com/bhojpur/dcp/pkg/engine/healthchecker"
	"github.com/bhojpur/dcp/pkg/engine/transport"
	"github.com/bhojpur/dcp/pkg/engine/util"
)

type loadBalancerAlgo interface {
	PickOne() *RemoteProxy
	Name() string
}

type rrLoadBalancerAlgo struct {
	sync.Mutex
	backends []*RemoteProxy
	next     int
}

func (rr *rrLoadBalancerAlgo) Name() string {
	return "rr algorithm"
}

func (rr *rrLoadBalancerAlgo) PickOne() *RemoteProxy {
	if len(rr.backends) == 0 {
		return nil
	} else if len(rr.backends) == 1 {
		if rr.backends[0].IsHealthy() {
			return rr.backends[0]
		}
		return nil
	} else {
		// round robin
		rr.Lock()
		defer rr.Unlock()
		hasFound := false
		selected := rr.next
		for i := 0; i < len(rr.backends); i++ {
			selected = (rr.next + i) % len(rr.backends)
			if rr.backends[selected].IsHealthy() {
				hasFound = true
				break
			}
		}

		if hasFound {
			rr.next = (selected + 1) % len(rr.backends)
			return rr.backends[selected]
		}
	}

	return nil
}

type priorityLoadBalancerAlgo struct {
	sync.Mutex
	backends []*RemoteProxy
}

func (prio *priorityLoadBalancerAlgo) Name() string {
	return "priority algorithm"
}

func (prio *priorityLoadBalancerAlgo) PickOne() *RemoteProxy {
	if len(prio.backends) == 0 {
		return nil
	} else if len(prio.backends) == 1 {
		if prio.backends[0].IsHealthy() {
			return prio.backends[0]
		}
		return nil
	} else {
		prio.Lock()
		defer prio.Unlock()
		for i := 0; i < len(prio.backends); i++ {
			if prio.backends[i].IsHealthy() {
				return prio.backends[i]
			}
		}

		return nil
	}
}

// LoadBalancer is an interface for proxying http request to remote server
// based on the load balance mode(round-robin or priority)
type LoadBalancer interface {
	IsHealthy() bool
	ServeHTTP(rw http.ResponseWriter, req *http.Request)
}

type loadBalancer struct {
	backends    []*RemoteProxy
	algo        loadBalancerAlgo
	certManager interfaces.EngineCertificateManager
}

// NewLoadBalancer creates a loadbalancer for specified remote servers
func NewLoadBalancer(
	lbMode string,
	remoteServers []*url.URL,
	cacheMgr cachemanager.CacheManager,
	transportMgr transport.Interface,
	healthChecker healthchecker.HealthChecker,
	certManager interfaces.EngineCertificateManager,
	filterChain filter.Interface,
	stopCh <-chan struct{}) (LoadBalancer, error) {
	backends := make([]*RemoteProxy, 0, len(remoteServers))
	for i := range remoteServers {
		b, err := NewRemoteProxy(remoteServers[i], cacheMgr, transportMgr, healthChecker, filterChain, stopCh)
		if err != nil {
			klog.Errorf("could not new proxy backend(%s), %v", remoteServers[i].String(), err)
			continue
		}
		backends = append(backends, b)
	}
	if len(backends) == 0 {
		return nil, fmt.Errorf("no backends can be used by lb")
	}

	var algo loadBalancerAlgo
	switch lbMode {
	case "rr":
		algo = &rrLoadBalancerAlgo{backends: backends}
	case "priority":
		algo = &priorityLoadBalancerAlgo{backends: backends}
	default:
		algo = &rrLoadBalancerAlgo{backends: backends}
	}

	return &loadBalancer{
		backends:    backends,
		algo:        algo,
		certManager: certManager,
	}, nil
}

func (lb *loadBalancer) IsHealthy() bool {
	// both certificate is not expired and
	// have at least one healthy remote server,
	// load balancer can proxy the request to
	// remote server
	if lb.certManager.NotExpired() {
		for i := range lb.backends {
			if lb.backends[i].IsHealthy() {
				return true
			}
		}
	}
	return false
}

func (lb *loadBalancer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// pick a remote proxy based on the load balancing algorithm.
	rp := lb.algo.PickOne()
	if rp == nil {
		// exceptional case
		klog.Errorf("could not pick one healthy backends by %s for request %s", lb.algo.Name(), util.ReqString(req))
		http.Error(rw, "could not pick one healthy backends, try again to go through local proxy.", http.StatusInternalServerError)
		return
	}
	klog.V(3).Infof("picked backend %s by %s for request %s", rp.Name(), lb.algo.Name(), util.ReqString(req))
	rp.ServeHTTP(rw, req)
}
