package config

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
	"crypto/x509"
	"fmt"
	"net"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"

	"github.com/bhojpur/dcp/pkg/projectinfo"
	"github.com/bhojpur/dcp/pkg/tunnel/constants"
)

// Config is the main context object for tunel-server
type Config struct {
	EgressSelectorEnabled       bool
	EnableIptables              bool
	EnableDNSController         bool
	IptablesSyncPeriod          int
	DNSSyncPeriod               int
	CertDNSNames                []string
	CertIPs                     []net.IP
	CertDir                     string
	ListenAddrForAgent          string
	ListenAddrForMaster         string
	ListenInsecureAddrForMaster string
	ListenMetaAddr              string
	RootCert                    *x509.CertPool
	Client                      kubernetes.Interface
	SharedInformerFactory       informers.SharedInformerFactory
	ServerCount                 int
	ProxyStrategy               string
	InterceptorServerUDSFile    string
}

type completedConfig struct {
	*Config
}

// CompletedConfig same as Config, just to swap private object.
type CompletedConfig struct {
	// Embed a private pointer that cannot be instantiated outside of this package.
	*completedConfig
}

// Complete fills in any fields not set that are required to have valid data. It's mutating the receiver.
func (c *Config) Complete() *CompletedConfig {
	cc := completedConfig{c}

	if cc.InterceptorServerUDSFile == "" {
		cc.InterceptorServerUDSFile = "/tmp/interceptor-proxier.sock"
	}
	if cc.CertDir == "" {
		cc.CertDir = fmt.Sprintf(constants.TunnelServerCertDir, projectinfo.GetServerName())
	}
	return &CompletedConfig{&cc}
}
