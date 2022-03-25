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
	"context"
	sysnet "net"
	"net/url"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/bhojpur/dcp/pkg/cloud/agent/loadbalancer"
	"github.com/pkg/errors"
)

type Proxy interface {
	Update(addresses []string)
	SetAPIServerPort(ctx context.Context, port int) error
	SetSupervisorDefault(address string)
	SupervisorURL() string
	SupervisorAddresses() []string
	APIServerURL() string
	IsAPIServerLBEnabled() bool
}

// NewSupervisorProxy sets up a new proxy for retrieving supervisor and apiserver addresses.  If
// lbEnabled is true, a load-balancer is started on the requested port to connect to the supervisor
// address, and the address of this local load-balancer is returned instead of the actual supervisor
// and apiserver addresses.
// NOTE: This is a proxy in the API sense - it returns either actual server URLs, or the URL of the
// local load-balancer. It is not actually responsible for proxying requests at the network level;
// this is handled by the load-balancers that the proxy optionally steers connections towards.
func NewSupervisorProxy(ctx context.Context, lbEnabled bool, dataDir, supervisorURL string, lbServerPort int) (Proxy, error) {
	p := proxy{
		lbEnabled:            lbEnabled,
		dataDir:              dataDir,
		initialSupervisorURL: supervisorURL,
		supervisorURL:        supervisorURL,
		apiServerURL:         supervisorURL,
		lbServerPort:         lbServerPort,
	}

	if lbEnabled {
		lb, err := loadbalancer.New(ctx, dataDir, loadbalancer.SupervisorServiceName, supervisorURL, p.lbServerPort)
		if err != nil {
			return nil, err
		}
		p.supervisorLB = lb
		p.supervisorURL = lb.LoadBalancerServerURL()
		p.apiServerURL = p.supervisorURL
	}

	u, err := url.Parse(p.initialSupervisorURL)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse %s", p.initialSupervisorURL)
	}
	p.fallbackSupervisorAddress = u.Host
	p.supervisorPort = u.Port()

	return &p, nil
}

type proxy struct {
	dataDir          string
	lbEnabled        bool
	lbServerPort     int
	apiServerEnabled bool

	apiServerURL              string
	supervisorURL             string
	supervisorPort            string
	initialSupervisorURL      string
	fallbackSupervisorAddress string
	supervisorAddresses       []string

	apiServerLB  *loadbalancer.LoadBalancer
	supervisorLB *loadbalancer.LoadBalancer
}

func (p *proxy) Update(addresses []string) {
	apiServerAddresses := addresses
	supervisorAddresses := addresses

	if p.apiServerEnabled {
		supervisorAddresses = p.setSupervisorPort(supervisorAddresses)
	}
	if p.apiServerLB != nil {
		p.apiServerLB.Update(apiServerAddresses)
	}
	if p.supervisorLB != nil {
		p.supervisorLB.Update(supervisorAddresses)
	}
	p.supervisorAddresses = supervisorAddresses
}

func (p *proxy) setSupervisorPort(addresses []string) []string {
	var newAddresses []string
	for _, address := range addresses {
		h, _, err := sysnet.SplitHostPort(address)
		if err != nil {
			logrus.Errorf("Failed to parse address %s, dropping: %v", address, err)
			continue
		}
		newAddresses = append(newAddresses, sysnet.JoinHostPort(h, p.supervisorPort))
	}
	return newAddresses
}

// SetAPIServerPort configures the proxy to return a different set of addresses for the apiserver,
// for use in cases where the apiserver is not running on the same port as the supervisor. If
// load-balancing is enabled, another load-balancer is started on a port one below the supervisor
// load-balancer, and the address of this load-balancer is returned instead of the actual apiserver
// addresses.
func (p *proxy) SetAPIServerPort(ctx context.Context, port int) error {
	u, err := url.Parse(p.initialSupervisorURL)
	if err != nil {
		return errors.Wrapf(err, "failed to parse server URL %s", p.initialSupervisorURL)
	}
	u.Host = sysnet.JoinHostPort(u.Hostname(), strconv.Itoa(port))

	p.apiServerURL = u.String()
	p.apiServerEnabled = true

	if p.lbEnabled && p.apiServerLB == nil {
		lbServerPort := p.lbServerPort
		if lbServerPort != 0 {
			lbServerPort = lbServerPort - 1
		}
		lb, err := loadbalancer.New(ctx, p.dataDir, loadbalancer.APIServerServiceName, p.apiServerURL, lbServerPort)
		if err != nil {
			return err
		}
		p.apiServerURL = lb.LoadBalancerServerURL()
		p.apiServerLB = lb
	}

	return nil
}

// SetSupervisorDefault updates the default (fallback) address for the connection to the
// supervisor. This is most useful on Bhojpur DCP nodes without apiservers, where the local
// supervisor must be used to bootstrap the agent config, but then switched over to
// another node running an apiserver once one is available.
func (p *proxy) SetSupervisorDefault(address string) {
	host, port, err := sysnet.SplitHostPort(address)
	if err != nil {
		logrus.Errorf("Failed to parse address %s, dropping: %v", address, err)
		return
	}
	if p.apiServerEnabled {
		port = p.supervisorPort
		address = sysnet.JoinHostPort(host, port)
	}
	p.fallbackSupervisorAddress = address
	if p.supervisorLB == nil {
		p.supervisorURL = "https://" + address
	} else {
		p.supervisorLB.SetDefault(address)
	}
}

func (p *proxy) SupervisorURL() string {
	return p.supervisorURL
}

func (p *proxy) SupervisorAddresses() []string {
	if len(p.supervisorAddresses) > 0 {
		return p.supervisorAddresses
	}
	return []string{p.fallbackSupervisorAddress}
}

func (p *proxy) APIServerURL() string {
	return p.apiServerURL
}

func (p *proxy) IsAPIServerLBEnabled() bool {
	return p.apiServerLB != nil
}
