package server

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
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"k8s.io/klog/v2"
	anpserver "sigs.k8s.io/apiserver-network-proxy/pkg/server"
	anpagent "sigs.k8s.io/apiserver-network-proxy/proto/agent"

	"github.com/bhojpur/dcp/pkg/tunnel/constants"
	hw "github.com/bhojpur/dcp/pkg/tunnel/handlerwrapper"
	wh "github.com/bhojpur/dcp/pkg/tunnel/handlerwrapper/wraphandler"
)

// anpTunnelServer implements the TunnelServer interface using the
// apiserver-network-proxy package
type anpTunnelServer struct {
	egressSelectorEnabled    bool
	interceptorServerUDSFile string
	serverMasterAddr         string
	serverMasterInsecureAddr string
	serverAgentAddr          string
	serverCount              int
	tlsCfg                   *tls.Config
	wrappers                 hw.HandlerWrappers
	proxyStrategy            string
}

var _ TunnelServer = &anpTunnelServer{}

// Run runs the tunnel-server
func (ats *anpTunnelServer) Run() error {
	proxyServer := anpserver.NewProxyServer(uuid.New().String(),
		[]anpserver.ProxyStrategy{anpserver.ProxyStrategy(ats.proxyStrategy)},
		ats.serverCount,
		&anpserver.AgentTokenAuthenticationOptions{})
	// 1. start the proxier
	proxierErr := runProxier(
		&anpserver.Tunnel{Server: proxyServer},
		ats.egressSelectorEnabled,
		ats.interceptorServerUDSFile,
		ats.tlsCfg)
	if proxierErr != nil {
		return fmt.Errorf("fail to run the proxier: %s", proxierErr)
	}

	wrappedHandler, err := wh.WrapHandler(
		NewRequestInterceptor(ats.interceptorServerUDSFile, ats.tlsCfg),
		ats.wrappers,
	)
	if err != nil {
		return fmt.Errorf("fail to wrap handler: %v", err)
	}

	// 2. start the master server
	masterServerErr := runMasterServer(
		wrappedHandler,
		ats.egressSelectorEnabled,
		ats.serverMasterAddr,
		ats.serverMasterInsecureAddr,
		ats.tlsCfg)
	if masterServerErr != nil {
		return fmt.Errorf("fail to run master server: %s", masterServerErr)
	}

	// 3. start the agent server
	agentServerErr := runAgentServer(ats.tlsCfg, ats.serverAgentAddr, proxyServer)
	if agentServerErr != nil {
		return fmt.Errorf("fail to run agent server: %s", agentServerErr)
	}

	return nil
}

// runProxier starts a proxy server that redirects requests received from
// apiserver to corresponding tunel-agent
func runProxier(handler http.Handler,
	egressSelectorEnabled bool,
	udsSockFile string,
	tlsConfig *tls.Config) error {
	klog.Info("start handling request from interceptor")
	if egressSelectorEnabled {
		// TODO will support egress selector for apiserver version > 1.18
		return errors.New("DOESN'T SUPPROT EGRESS SELECTOR YET")
	}
	// request will be sent from request interceptor on the same host,
	// so we use UDS protocol to avoide sending request through kernel
	// network stack.
	go func() {
		server := &http.Server{
			Handler:     handler,
			ReadTimeout: constants.TunnelANPProxierReadTimeoutSec * time.Second,
		}
		unixListener, err := net.Listen("unix", udsSockFile)
		if err != nil {
			klog.Errorf("proxier fail to serving request through uds: %s", err)
		}
		defer unixListener.Close()
		if err := server.Serve(unixListener); err != nil {
			klog.Errorf("proxier fail to serving request through uds: %s", err)
		}
	}()

	return nil
}

// runMasterServer runs an https server to handle requests from apiserver
func runMasterServer(handler http.Handler,
	egressSelectorEnabled bool,
	masterServerAddr,
	masterServerInsecureAddr string,
	tlsCfg *tls.Config) error {
	if egressSelectorEnabled {
		return errors.New("DOESN'T SUPPORT EGRESS SELECTOR YET")
	}
	go func() {
		klog.Infof("start handling https request from master at %s", masterServerAddr)
		server := http.Server{
			Addr:         masterServerAddr,
			Handler:      handler,
			ReadTimeout:  constants.TunnelANPInterceptorReadTimeoutSec * time.Second,
			TLSConfig:    tlsCfg,
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
		}
		if err := server.ListenAndServeTLS("", ""); err != nil {
			klog.Errorf("failed to serve https request from master: %v", err)
		}
	}()

	go func() {
		klog.Infof("start handling http request from master at %s", masterServerInsecureAddr)
		server := http.Server{
			Addr:         masterServerInsecureAddr,
			ReadTimeout:  constants.TunnelANPInterceptorReadTimeoutSec * time.Second,
			Handler:      handler,
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
		}
		if err := server.ListenAndServe(); err != nil {
			klog.Errorf("failed to serve http request from master: %v", err)
		}
	}()

	return nil
}

// runAgentServer runs a grpc server that handles connections from the tunnel-agent
// NOTE agent server is responsible for managing grpc connection tunnel-server
// and tunnel-agent, and the proxy server is responsible for redirecting requests
// to corresponding tunnel-agent
func runAgentServer(tlsCfg *tls.Config,
	agentServerAddr string,
	proxyServer *anpserver.ProxyServer) error {
	serverOption := grpc.Creds(credentials.NewTLS(tlsCfg))

	ka := keepalive.ServerParameters{
		// Ping the client if it is idle for `Time` seconds to ensure the
		// connection is still active
		Time: constants.TunnelANPGrpcKeepAliveTimeSec * time.Second,
		// Wait `Timeout` second for the ping ack before assuming the
		// connection is dead
		Timeout: constants.TunnelANPGrpcKeepAliveTimeoutSec * time.Second,
	}

	grpcServer := grpc.NewServer(serverOption,
		grpc.KeepaliveParams(ka))

	anpagent.RegisterAgentServiceServer(grpcServer, proxyServer)
	listener, err := net.Listen("tcp", agentServerAddr)
	klog.Info("start handling connection from agents")
	if err != nil {
		return fmt.Errorf("fail to listen to agent on %s: %s", agentServerAddr, err)
	}
	go grpcServer.Serve(listener)
	return nil
}
