package cmd

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
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/cmd/server/config"
	"github.com/bhojpur/dcp/cmd/server/options"
	"github.com/bhojpur/dcp/pkg/engine/cachemanager"
	"github.com/bhojpur/dcp/pkg/engine/certificate"
	"github.com/bhojpur/dcp/pkg/engine/certificate/hubself"
	"github.com/bhojpur/dcp/pkg/engine/gc"
	"github.com/bhojpur/dcp/pkg/engine/healthchecker"
	"github.com/bhojpur/dcp/pkg/engine/kubernetes/rest"
	"github.com/bhojpur/dcp/pkg/engine/network"
	"github.com/bhojpur/dcp/pkg/engine/proxy"
	"github.com/bhojpur/dcp/pkg/engine/server"
	"github.com/bhojpur/dcp/pkg/engine/transport"
	"github.com/bhojpur/dcp/pkg/engine/util"
	"github.com/bhojpur/dcp/pkg/projectinfo"
)

// NewCmdStartEngine creates a *cobra.Command object with default parameters
func NewCmdStartEngine(stopCh <-chan struct{}) *cobra.Command {
	engineOptions := options.NewEngineOptions()

	cmd := &cobra.Command{
		Use:   projectinfo.GetEngineName(),
		Short: "Launch " + projectinfo.GetEngineName(),
		Long:  "Launch " + projectinfo.GetEngineName(),
		Run: func(cmd *cobra.Command, args []string) {
			if engineOptions.Version {
				fmt.Printf("%s: %#v\n", projectinfo.GetEngineName(), projectinfo.Get())
				return
			}
			fmt.Printf("%s version: %#v\n", projectinfo.GetEngineName(), projectinfo.Get())

			cmd.Flags().VisitAll(func(flag *pflag.Flag) {
				klog.V(1).Infof("FLAG: --%s=%q", flag.Name, flag.Value)
			})
			if err := options.ValidateOptions(engineOptions); err != nil {
				klog.Fatalf("validate options: %v", err)
			}

			engineCfg, err := config.Complete(engineOptions)
			if err != nil {
				klog.Fatalf("complete %s configuration error, %v", projectinfo.GetEngineName(), err)
			}
			klog.Infof("%s cfg: %#+v", projectinfo.GetEngineName(), engineCfg)

			if err := Run(engineCfg, stopCh); err != nil {
				klog.Fatalf("run %s failed, %v", projectinfo.GetEngineName(), err)
			}
		},
	}

	engineOptions.AddFlags(cmd.Flags())
	return cmd
}

// Run runs the EngineConfiguration. This should never exit
func Run(cfg *config.EngineConfiguration, stopCh <-chan struct{}) error {
	trace := 1
	klog.Infof("%d. register cert managers", trace)
	cmr := certificate.NewCertificateManagerRegistry()
	hubself.Register(cmr)
	trace++

	klog.Infof("%d. create cert manager with %s mode", trace, cfg.CertMgrMode)
	certManager, err := cmr.New(cfg.CertMgrMode, cfg)
	if err != nil {
		return fmt.Errorf("could not create certificate manager, %v", err)
	}
	trace++

	klog.Infof("%d. new transport manager", trace)
	transportManager, err := transport.NewTransportManager(certManager, stopCh)
	if err != nil {
		return fmt.Errorf("could not new transport manager, %v", err)
	}
	trace++

	var healthChecker healthchecker.HealthChecker
	if cfg.WorkingMode == util.WorkingModeEdge {
		klog.Infof("%d. create health checker for remote servers ", trace)
		healthChecker, err = healthchecker.NewHealthChecker(cfg, transportManager, stopCh)
		if err != nil {
			return fmt.Errorf("could not new health checker, %v", err)
		}
	} else {
		klog.Infof("%d. disable health checker for node %s because it is a cloud node", trace, cfg.NodeName)
		// In cloud mode, health checker is not needed.
		// This fake checker will always report that the remote server is healthy.
		healthChecker = healthchecker.NewFakeChecker(true, make(map[string]int))
	}
	healthChecker.Run()
	trace++

	klog.Infof("%d. new restConfig manager for %s mode", trace, cfg.CertMgrMode)
	restConfigMgr, err := rest.NewRestConfigManager(cfg, certManager, healthChecker)
	if err != nil {
		return fmt.Errorf("could not new restConfig manager, %v", err)
	}
	trace++

	klog.Infof("%d. create tls config for secure servers ", trace)
	cfg.TLSConfig, err = server.GenUseCertMgrAndTLSConfig(restConfigMgr, certManager, filepath.Join(cfg.RootDir, "pki"), cfg.EngineProxyServerSecureDummyAddr, stopCh)
	if err != nil {
		return fmt.Errorf("could not create tls config, %v", err)
	}
	trace++

	var cacheMgr cachemanager.CacheManager
	if cfg.WorkingMode == util.WorkingModeEdge {
		klog.Infof("%d. new cache manager with storage wrapper and serializer manager", trace)
		cacheMgr, err = cachemanager.NewCacheManager(cfg.StorageWrapper, cfg.SerializerManager, cfg.RESTMapperManager, cfg.SharedFactory)
		if err != nil {
			return fmt.Errorf("could not new cache manager, %v", err)
		}
	} else {
		klog.Infof("%d. disable cache manager for node %s because it is a cloud node", trace, cfg.NodeName)
	}
	trace++

	if cfg.WorkingMode == util.WorkingModeEdge {
		klog.Infof("%d. new gc manager for node %s, and gc frequency is a random time between %d min and %d min", trace, cfg.NodeName, cfg.GCFrequency, 3*cfg.GCFrequency)
		gcMgr, err := gc.NewGCManager(cfg, restConfigMgr, stopCh)
		if err != nil {
			return fmt.Errorf("could not new gc manager, %v", err)
		}
		gcMgr.Run()
	} else {
		klog.Infof("%d. disable gc manager for node %s because it is a cloud node", trace, cfg.NodeName)
	}
	trace++

	klog.Infof("%d. new reverse proxy handler for remote servers", trace)
	dcpProxyHandler, err := proxy.NewReverseProxyHandler(cfg, cacheMgr, transportManager, healthChecker, certManager, stopCh)
	if err != nil {
		return fmt.Errorf("could not create reverse proxy handler, %v", err)
	}
	trace++

	if cfg.EnableDummyIf {
		klog.Infof("%d. create dummy network interface %s and init iptables manager", trace, cfg.HubAgentDummyIfName)
		networkMgr, err := network.NewNetworkManager(cfg)
		if err != nil {
			return fmt.Errorf("could not create network manager, %v", err)
		}
		networkMgr.Run(stopCh)
		trace++
		klog.Infof("%d. new %s server and begin to serve, dummy proxy server: %s, secure dummy proxy server: %s", trace, projectinfo.GetEngineName(), cfg.EngineProxyServerDummyAddr, cfg.EngineProxyServerSecureDummyAddr)
	}

	// start shared informers before start Bhojpur DCP server engine
	cfg.SharedFactory.Start(stopCh)
	cfg.DcpSharedFactory.Start(stopCh)

	klog.Infof("%d. new %s server and begin to serve, proxy server: %s, secure proxy server: %s, hub server: %s", trace, projectinfo.GetEngineName(), cfg.EngineProxyServerAddr, cfg.EngineProxyServerSecureAddr, cfg.EngineServerAddr)
	s, err := server.NewEngineServer(cfg, certManager, dcpProxyHandler)
	if err != nil {
		return fmt.Errorf("could not create Bhojpur DCP engine server, %v", err)
	}
	s.Run()
	klog.Infof("Bhojpur DCP engine agent exited")
	return nil
}
