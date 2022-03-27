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
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/cmd/grid/server/config"
	"github.com/bhojpur/dcp/pkg/engine/certificate/interfaces"
	"github.com/bhojpur/dcp/pkg/engine/kubernetes/rest"
	"github.com/bhojpur/dcp/pkg/profile"
	"github.com/bhojpur/dcp/pkg/projectinfo"
	"github.com/bhojpur/dcp/pkg/utils/certmanager"
)

// Server is an interface for providing http service for Bhojpur DCP server engine
type Server interface {
	Run()
}

// engineServer includes dcpServer and proxyServer,
// and dcpServer handles requests by engine agent itself, like profiling, metrics, healthz
// and proxyServer does not handle requests locally and proxy requests to kube-apiserver
type engineServer struct {
	dcpServer              *http.Server
	proxyServer            *http.Server
	secureProxyServer      *http.Server
	dummyProxyServer       *http.Server
	dummySecureProxyServer *http.Server
}

// NewEngineServer creates a Server object
func NewEngineServer(cfg *config.EngineConfiguration,
	certificateMgr interfaces.EngineCertificateManager,
	proxyHandler http.Handler) (Server, error) {
	dcpMux := mux.NewRouter()
	registerHandlers(dcpMux, cfg, certificateMgr)
	dcpServer := &http.Server{
		Addr:           cfg.EngineServerAddr,
		Handler:        dcpMux,
		MaxHeaderBytes: 1 << 20,
	}

	proxyServer := &http.Server{
		Addr:    cfg.EngineProxyServerAddr,
		Handler: proxyHandler,
	}

	secureProxyServer := &http.Server{
		Addr:           cfg.EngineProxyServerSecureAddr,
		Handler:        proxyHandler,
		TLSConfig:      cfg.TLSConfig,
		TLSNextProto:   make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
		MaxHeaderBytes: 1 << 20,
	}

	var dummyProxyServer, secureDummyProxyServer *http.Server
	if cfg.EnableDummyIf {
		if _, err := net.InterfaceByName(cfg.HubAgentDummyIfName); err != nil {
			return nil, err
		}

		dummyProxyServer = &http.Server{
			Addr:           cfg.EngineProxyServerDummyAddr,
			Handler:        proxyHandler,
			MaxHeaderBytes: 1 << 20,
		}

		secureDummyProxyServer = &http.Server{
			Addr:           cfg.EngineProxyServerSecureDummyAddr,
			Handler:        proxyHandler,
			TLSConfig:      cfg.TLSConfig,
			TLSNextProto:   make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
			MaxHeaderBytes: 1 << 20,
		}
	}

	return &engineServer{
		dcpServer:              dcpServer,
		proxyServer:            proxyServer,
		secureProxyServer:      secureProxyServer,
		dummyProxyServer:       dummyProxyServer,
		dummySecureProxyServer: secureDummyProxyServer,
	}, nil
}

// Run will start Bhojpur DCP server engine and proxy server
func (s *engineServer) Run() {
	go func() {
		err := s.dcpServer.ListenAndServe()
		if err != nil {
			panic(err)
		}
	}()

	if s.dummyProxyServer != nil {
		go func() {
			err := s.dummyProxyServer.ListenAndServe()
			if err != nil {
				panic(err)
			}
		}()
		go func() {
			err := s.dummySecureProxyServer.ListenAndServeTLS("", "")
			if err != nil {
				panic(err)
			}
		}()
	}

	go func() {
		err := s.secureProxyServer.ListenAndServeTLS("", "")
		if err != nil {
			panic(err)
		}
	}()

	err := s.proxyServer.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

// registerHandler registers handlers for engineServer, and engineServer can handle requests like profiling, healthz, update token.
func registerHandlers(c *mux.Router, cfg *config.EngineConfiguration, certificateMgr interfaces.EngineCertificateManager) {
	// register handlers for update join token
	c.Handle("/v1/token", updateTokenHandler(certificateMgr)).Methods("POST", "PUT")

	// register handler for health check
	c.HandleFunc("/v1/healthz", healthz).Methods("GET")

	// register handler for profile
	if cfg.EnableProfiling {
		profile.Install(c)
	}

	// register handler for metrics
	c.Handle("/metrics", promhttp.Handler())
}

// healthz returns ok for healthz request
func healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

// create a certificate manager for the Bhojpur DCP server engine and run the CSR approver for both Bhojpur DCP
// and generate a TLS configuration
func GenUseCertMgrAndTLSConfig(restConfigMgr *rest.RestConfigManager, certificateMgr interfaces.EngineCertificateManager, certDir, proxyServerSecureDummyAddr string, stopCh <-chan struct{}) (*tls.Config, error) {
	cfg := restConfigMgr.GetRestConfig(false)
	if cfg == nil {
		return nil, fmt.Errorf("failed to prepare rest config based on hub agent client certificate")
	}

	clientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	// create a certificate manager for the Bhojpur DCP server engine and run the CSR approver for both Bhojpur DCP
	serverCertMgr, err := certmanager.NewEngineServerCertManager(clientSet, certDir, proxyServerSecureDummyAddr)
	if err != nil {
		return nil, err
	}
	serverCertMgr.Start()

	// generate the TLS configuration based on the latest certificate
	rootCert, err := certmanager.GenCertPoolUseCA(certificateMgr.GetCaFile())
	if err != nil {
		klog.Errorf("could not generate a x509 CertPool based on the given CA file, %v", err)
		return nil, err
	}
	tlsCfg, err := certmanager.GenTLSConfigUseCertMgrAndCertPool(serverCertMgr, rootCert)
	if err != nil {
		return nil, err
	}

	// waiting for the certificate is generated
	_ = wait.PollUntil(5*time.Second, func() (bool, error) {
		// keep polling until the certificate is signed
		if serverCertMgr.Current() != nil {
			return true, nil
		}
		klog.Infof("waiting for the master to sign the %s certificate", projectinfo.GetEngineName())
		return false, nil
	}, stopCh)

	return tlsCfg, nil
}
