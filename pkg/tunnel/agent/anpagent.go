package agent

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
	"crypto/tls"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"k8s.io/klog/v2"
	anpagent "sigs.k8s.io/apiserver-network-proxy/pkg/agent"

	"github.com/bhojpur/dcp/pkg/projectinfo"
)

// anpTunnelAgent implements the TunnelAgent using the
// apiserver-network-proxy package
type anpTunnelAgent struct {
	tlsCfg           *tls.Config
	tunnelServerAddr string
	nodeName         string
	agentIdentifiers string
}

var _ TunnelAgent = &anpTunnelAgent{}

// RunAgent runs the tunnel-agent which will try to connect tunnel-server
func (ata *anpTunnelAgent) Run(stopChan <-chan struct{}) {
	dialOption := grpc.WithTransportCredentials(credentials.NewTLS(ata.tlsCfg))
	cc := &anpagent.ClientSetConfig{
		Address:                 ata.tunnelServerAddr,
		AgentID:                 ata.nodeName,
		AgentIdentifiers:        ata.agentIdentifiers,
		SyncInterval:            5 * time.Second,
		ProbeInterval:           5 * time.Second,
		DialOptions:             []grpc.DialOption{dialOption},
		ServiceAccountTokenPath: "",
	}

	cs := cc.NewAgentClientSet(stopChan)
	cs.Serve()
	klog.Infof("start serving gRPC request redirected from %s: %s",
		projectinfo.GetServerName(), ata.tunnelServerAddr)
}
