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
	"sync"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/cmd/grid/tunnel-server/config"
	"github.com/bhojpur/dcp/cmd/grid/tunnel-server/options"
	"github.com/bhojpur/dcp/pkg/projectinfo"
	"github.com/bhojpur/dcp/pkg/tunnel/handlerwrapper/initializer"
	"github.com/bhojpur/dcp/pkg/tunnel/handlerwrapper/wraphandler"
	"github.com/bhojpur/dcp/pkg/tunnel/informers"
	"github.com/bhojpur/dcp/pkg/tunnel/server"
	"github.com/bhojpur/dcp/pkg/tunnel/trafficforward/dns"
	"github.com/bhojpur/dcp/pkg/tunnel/trafficforward/iptables"
	"github.com/bhojpur/dcp/pkg/tunnel/util"
	"github.com/bhojpur/dcp/pkg/utils/certmanager"
)

// NewTunnelServerCommand creates a new tunnel-server command
func NewTunnelServerCommand(stopCh <-chan struct{}) *cobra.Command {
	serverOptions := options.NewServerOptions()

	cmd := &cobra.Command{
		Use:   "Launch Bhojpur DCP " + projectinfo.GetServerName(),
		Short: projectinfo.GetServerName() + " sends requests to " + projectinfo.GetAgentName(),
		RunE: func(c *cobra.Command, args []string) error {
			if serverOptions.Version {
				fmt.Printf("%s: %#v\n", projectinfo.GetServerName(), projectinfo.Get())
				return nil
			}
			klog.Infof("%s version: %#v", projectinfo.GetServerName(), projectinfo.Get())

			if err := serverOptions.Validate(); err != nil {
				return err
			}

			cfg, err := serverOptions.Config()
			if err != nil {
				return err
			}
			if err := Run(cfg.Complete(), stopCh); err != nil {
				return err
			}
			return nil
		},
		Args: cobra.NoArgs,
	}

	serverOptions.AddFlags(cmd.Flags())

	return cmd
}

// run starts the tunel-server
func Run(cfg *config.CompletedConfig, stopCh <-chan struct{}) error {
	var wg sync.WaitGroup
	// register informers that tunnel server need
	informers.RegisterInformersForTunnelServer(cfg.SharedInformerFactory)

	// 0. start the DNS controller
	if cfg.EnableDNSController {
		dnsController, err := dns.NewCoreDNSRecordController(cfg.Client,
			cfg.SharedInformerFactory,
			cfg.ListenInsecureAddrForMaster,
			cfg.ListenAddrForMaster,
			cfg.DNSSyncPeriod)
		if err != nil {
			return fmt.Errorf("fail to create a new dnsController, %v", err)
		}
		go dnsController.Run(stopCh)
	}
	// 1. start the IP table manager
	if cfg.EnableIptables {
		iptablesMgr := iptables.NewIptablesManager(cfg.Client,
			cfg.SharedInformerFactory.Core().V1().Nodes(),
			cfg.ListenAddrForMaster,
			cfg.ListenInsecureAddrForMaster,
			cfg.IptablesSyncPeriod)
		if iptablesMgr == nil {
			return fmt.Errorf("fail to create a new IptableManager")
		}
		wg.Add(1)
		go iptablesMgr.Run(stopCh, &wg)
	}

	// 2. create a certificate manager for the tunnel server and run the
	// csr approver for both tunnel-server and tunnel-agent
	serverCertMgr, err := certmanager.NewTunnelServerCertManager(cfg.Client, cfg.SharedInformerFactory, cfg.CertDir, cfg.CertDNSNames, cfg.CertIPs, stopCh)
	if err != nil {
		return err
	}
	serverCertMgr.Start()

	// 3. create handler wrappers
	mInitializer := initializer.NewMiddlewareInitializer(cfg.SharedInformerFactory)
	wrappers, err := wraphandler.InitHandlerWrappers(mInitializer)
	if err != nil {
		klog.Errorf("failed to init handler wrappers, %v", err)
		return err
	}

	// after all of informers are configured completed, start the shared index informer
	cfg.SharedInformerFactory.Start(stopCh)

	// 4. waiting for the certificate is generated
	_ = wait.PollUntil(5*time.Second, func() (bool, error) {
		// keep polling until the certificate is signed
		if serverCertMgr.Current() != nil {
			return true, nil
		}
		klog.Infof("waiting for the master to sign the %s certificate", projectinfo.GetServerName())
		return false, nil
	}, stopCh)

	// 5. generate the TLS configuration based on the latest certificate
	tlsCfg, err := certmanager.GenTLSConfigUseCertMgrAndCertPool(serverCertMgr, cfg.RootCert)
	if err != nil {
		return err
	}

	// 6. start the server
	ts := server.NewTunnelServer(
		cfg.EgressSelectorEnabled,
		cfg.InterceptorServerUDSFile,
		cfg.ListenAddrForMaster,
		cfg.ListenInsecureAddrForMaster,
		cfg.ListenAddrForAgent,
		cfg.ServerCount,
		tlsCfg,
		wrappers,
		cfg.ProxyStrategy)
	if err := ts.Run(); err != nil {
		return err
	}

	// 7. start meta server
	util.RunMetaServer(cfg.ListenMetaAddr)

	<-stopCh
	wg.Wait()
	return nil
}
