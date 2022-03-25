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
	"time"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/certificate"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/cmd/tunnel-agent/config"
	"github.com/bhojpur/dcp/cmd/tunnel-agent/options"
	"github.com/bhojpur/dcp/pkg/projectinfo"
	"github.com/bhojpur/dcp/pkg/tunnel/agent"
	"github.com/bhojpur/dcp/pkg/tunnel/constants"
	"github.com/bhojpur/dcp/pkg/tunnel/server/serveraddr"
	"github.com/bhojpur/dcp/pkg/tunnel/util"
	"github.com/bhojpur/dcp/pkg/utils/certmanager"
)

// NewTunnelAgentCommand creates a new tunnel-agent command
func NewTunnelAgentCommand(stopCh <-chan struct{}) *cobra.Command {
	agentOptions := options.NewAgentOptions()

	cmd := &cobra.Command{
		Short: fmt.Sprintf("Launch %s", projectinfo.GetAgentName()),
		RunE: func(c *cobra.Command, args []string) error {
			if agentOptions.Version {
				fmt.Printf("%s: %#v\n", projectinfo.GetAgentName(), projectinfo.Get())
				return nil
			}
			klog.Infof("%s version: %#v", projectinfo.GetAgentName(), projectinfo.Get())

			if err := agentOptions.Validate(); err != nil {
				return err
			}

			cfg, err := agentOptions.Config()
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

	agentOptions.AddFlags(cmd.Flags())
	return cmd
}

// Run starts the tunel-agent
func Run(cfg *config.CompletedConfig, stopCh <-chan struct{}) error {
	var (
		tunnelServerAddr string
		err              error
		agentCertMgr     certificate.Manager
	)

	// 1. get the address of the tunnel-server
	tunnelServerAddr = cfg.TunnelServerAddr
	if tunnelServerAddr == "" {
		if tunnelServerAddr, err = serveraddr.GetTunnelServerAddr(cfg.Client); err != nil {
			return err
		}
	}
	klog.Infof("%s address: %s", projectinfo.GetServerName(), tunnelServerAddr)

	// 2. create a certificate manager
	agentCertMgr, err =
		certmanager.NewTunnelAgentCertManager(cfg.Client, cfg.CertDir)
	if err != nil {
		return err
	}
	agentCertMgr.Start()

	// 2.1. waiting for the certificate is generated
	_ = wait.PollUntil(5*time.Second, func() (bool, error) {
		if agentCertMgr.Current() != nil {
			return true, nil
		}
		klog.Infof("certificate %s not signed, waiting...",
			projectinfo.GetAgentName())
		return false, nil
	}, stopCh)
	klog.Infof("certificate %s ok", projectinfo.GetAgentName())

	// 3. generate a TLS configuration for securing the connection to server
	tlsCfg, err := certmanager.GenTLSConfigUseCertMgrAndCA(agentCertMgr,
		tunnelServerAddr, constants.TunnelCAFile)
	if err != nil {
		return err
	}

	// 4. start the tunnel-agent
	ta := agent.NewTunnelAgent(tlsCfg, tunnelServerAddr, cfg.NodeName, cfg.AgentIdentifiers)
	ta.Run(stopCh)

	// 5. start meta server
	util.RunMetaServer(cfg.AgentMetaAddr)

	<-stopCh
	return nil
}
