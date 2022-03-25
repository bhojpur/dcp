package etcd

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
	"net/url"

	"github.com/bhojpur/dcp/pkg/cloud/agent/loadbalancer"
	"github.com/pkg/errors"
)

type Proxy interface {
	Update(addresses []string)
	ETCDURL() string
	ETCDAddresses() []string
	ETCDServerURL() string
}

// NewETCDProxy initializes a new proxy structure that contain a load balancer
// which listens on port 2379 and proxy between etcd cluster members
func NewETCDProxy(ctx context.Context, enabled bool, dataDir, etcdURL string) (Proxy, error) {
	u, err := url.Parse(etcdURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse etcd client URL")
	}

	e := &etcdproxy{
		dataDir:        dataDir,
		initialETCDURL: etcdURL,
		etcdURL:        etcdURL,
	}

	if enabled {
		lb, err := loadbalancer.New(ctx, dataDir, loadbalancer.ETCDServerServiceName, etcdURL, 2379)
		if err != nil {
			return nil, err
		}
		e.etcdLB = lb
		e.etcdLBURL = lb.LoadBalancerServerURL()
	}

	e.fallbackETCDAddress = u.Host
	e.etcdPort = u.Port()

	return e, nil
}

type etcdproxy struct {
	dataDir   string
	etcdLBURL string

	initialETCDURL      string
	etcdURL             string
	etcdPort            string
	fallbackETCDAddress string
	etcdAddresses       []string
	etcdLB              *loadbalancer.LoadBalancer
}

func (e *etcdproxy) Update(addresses []string) {
	if e.etcdLB != nil {
		e.etcdLB.Update(addresses)
	}
}

func (e *etcdproxy) ETCDURL() string {
	return e.etcdURL
}

func (e *etcdproxy) ETCDAddresses() []string {
	if len(e.etcdAddresses) > 0 {
		return e.etcdAddresses
	}
	return []string{e.fallbackETCDAddress}
}

func (e *etcdproxy) ETCDServerURL() string {
	return e.etcdURL
}
