package network

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
	"net"
	"time"

	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/cmd/grid/dcpsvr/config"
)

const (
	SyncNetworkPeriod = 60
)

type NetworkManager struct {
	ifController    DummyInterfaceController
	iptablesManager *IptablesManager
	dummyIfIP       net.IP
	dummyIfName     string
	enableIptables  bool
}

func NewNetworkManager(cfg *config.EngineConfiguration) (*NetworkManager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration for hub agent is nil")
	}

	ip, port, err := net.SplitHostPort(cfg.EngineProxyServerDummyAddr)
	if err != nil {
		return nil, err
	}
	m := &NetworkManager{
		ifController:    NewDummyInterfaceController(),
		iptablesManager: NewIptablesManager(ip, port),
		dummyIfIP:       net.ParseIP(ip),
		dummyIfName:     cfg.HubAgentDummyIfName,
		enableIptables:  cfg.EnableIptables,
	}
	// secure port
	_, securePort, err := net.SplitHostPort(cfg.EngineProxyServerSecureDummyAddr)
	if err != nil {
		return nil, err
	}
	m.iptablesManager.rules = append(m.iptablesManager.rules, makeupIptablesRules(ip, securePort)...)
	if err = m.configureNetwork(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *NetworkManager) Run(stopCh <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(SyncNetworkPeriod * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-stopCh:
				klog.Infof("exit network manager run goroutine normally")
				m.iptablesManager.CleanUpIptablesRules()
				err := m.ifController.DeleteDummyInterface(m.dummyIfName)
				if err != nil {
					klog.Errorf("failed to delete dummy interface %s, %v", m.dummyIfName, err)
				}
				return
			case <-ticker.C:
				if err := m.configureNetwork(); err != nil {
					// do nothing here
				}
			}
		}
	}()
}

func (m *NetworkManager) configureNetwork() error {
	err := m.ifController.EnsureDummyInterface(m.dummyIfName, m.dummyIfIP)
	if err != nil {
		klog.Errorf("ensure dummy interface failed, %v", err)
		return err
	}

	if m.enableIptables {
		err := m.iptablesManager.EnsureIptablesRules()
		if err != nil {
			klog.Errorf("ensure iptables for dummy interface failed, %v", err)
			return err
		}
	}

	return nil
}
