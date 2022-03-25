package flannel

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
	"os"
	"path/filepath"
	"sync"

	"github.com/flannel-io/flannel/backend"
	"github.com/flannel-io/flannel/network"
	"github.com/flannel-io/flannel/pkg/ip"
	"github.com/flannel-io/flannel/subnet/kube"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	// Backends need to be imported for their init() to get executed and them to register
	_ "github.com/flannel-io/flannel/backend/extension"
	_ "github.com/flannel-io/flannel/backend/hostgw"
	_ "github.com/flannel-io/flannel/backend/ipsec"
	_ "github.com/flannel-io/flannel/backend/vxlan"
)

const (
	subnetFile = "/run/flannel/subnet.env"
)

func flannel(ctx context.Context, flannelIface *net.Interface, flannelConf, kubeConfigFile string, flannelIPv6Masq bool, netMode int) error {
	extIface, err := LookupExtInterface(flannelIface, netMode)
	if err != nil {
		return err
	}

	sm, err := kube.NewSubnetManager(ctx, "", kubeConfigFile, "flannel.alpha.coreos.com", flannelConf, false)
	if err != nil {
		return err
	}

	config, err := sm.GetNetworkConfig(ctx)
	if err != nil {
		return err
	}

	// Create a backend manager then use it to create the backend and register the network with it.
	bm := backend.NewManager(ctx, sm, extIface)

	be, err := bm.GetBackend(config.BackendType)
	if err != nil {
		return err
	}

	bn, err := be.RegisterNetwork(ctx, &sync.WaitGroup{}, config)
	if err != nil {
		return err
	}

	if netMode == (ipv4+ipv6) || netMode == ipv4 {
		go network.SetupAndEnsureIPTables(network.MasqRules(config.Network, bn.Lease()), 60)
		go network.SetupAndEnsureIPTables(network.ForwardRules(config.Network.String()), 50)
	}

	if flannelIPv6Masq && config.IPv6Network.String() != emptyIPv6Network {
		logrus.Debugf("Creating IPv6 masquerading iptables rules for %s network", config.IPv6Network.String())
		go network.SetupAndEnsureIP6Tables(network.MasqIP6Rules(config.IPv6Network, bn.Lease()), 60)
		go network.SetupAndEnsureIP6Tables(network.ForwardRules(config.IPv6Network.String()), 50)
	}

	if err := WriteSubnetFile(subnetFile, config.Network, config.IPv6Network, true, bn, netMode); err != nil {
		// Continue, even though it failed.
		logrus.Warningf("Failed to write flannel subnet file: %s", err)
	} else {
		logrus.Infof("Wrote flannel subnet file to %s", subnetFile)
	}

	// Start "Running" the backend network. This will block until the context is done so run in another goroutine.
	logrus.Info("Running flannel backend.")
	bn.Run(ctx)
	return nil
}

func LookupExtInterface(iface *net.Interface, netMode int) (*backend.ExternalInterface, error) {
	var ifaceAddr []net.IP
	var ifacev6Addr []net.IP
	var err error

	if iface == nil {
		logrus.Debug("No interface defined for flannel in the config. Fetching the default gateway interface")
		if netMode == ipv4 || netMode == (ipv4+ipv6) {
			if iface, err = ip.GetDefaultGatewayInterface(); err != nil {
				return nil, fmt.Errorf("failed to get default interface: %s", err)
			}
		} else {
			if iface, err = ip.GetDefaultV6GatewayInterface(); err != nil {
				return nil, fmt.Errorf("failed to get default interface: %s", err)
			}
		}
	}
	logrus.Debugf("The interface %s will be used by flannel", iface.Name)

	switch netMode {
	case ipv4:
		ifaceAddr, err = ip.GetInterfaceIP4Addrs(iface)
		if err != nil {
			return nil, fmt.Errorf("failed to find IPv4 address for interface %s", iface.Name)
		}
		logrus.Infof("The interface %s with ipv4 address %s will be used by flannel", iface.Name, ifaceAddr[0])
		ifacev6Addr = append(ifacev6Addr, nil)
	case ipv6:
		ifacev6Addr, err = ip.GetInterfaceIP6Addrs(iface)
		if err != nil {
			return nil, fmt.Errorf("failed to find IPv6 address for interface %s", iface.Name)
		}
		logrus.Infof("The interface %s with ipv6 address %s will be used by flannel", iface.Name, ifacev6Addr[0])
		ifaceAddr = append(ifaceAddr, nil)
	case (ipv4 + ipv6):
		ifaceAddr, err = ip.GetInterfaceIP4Addrs(iface)
		if err != nil {
			return nil, fmt.Errorf("failed to find IPv4 address for interface %s", iface.Name)
		}
		logrus.Infof("The interface %s with ipv4 address %s will be used by flannel", iface.Name, ifaceAddr[0])
		ifacev6Addr, err = ip.GetInterfaceIP6Addrs(iface)
		if err != nil {
			return nil, fmt.Errorf("failed to find IPv6 address for interface %s", iface.Name)
		}
		logrus.Infof("Using dual-stack mode. The ipv6 address %s will be used by flannel", ifacev6Addr[0])
	default:
		ifaceAddr = append(ifaceAddr, nil)
		ifacev6Addr = append(ifacev6Addr, nil)
	}

	if iface.MTU == 0 {
		return nil, fmt.Errorf("failed to determine MTU for %s interface", iface.Name)
	}

	return &backend.ExternalInterface{
		Iface:       iface,
		IfaceAddr:   ifaceAddr[0],
		IfaceV6Addr: ifacev6Addr[0],
		ExtAddr:     ifaceAddr[0],
		ExtV6Addr:   ifacev6Addr[0],
	}, nil
}

func WriteSubnetFile(path string, nw ip.IP4Net, nwv6 ip.IP6Net, ipMasq bool, bn backend.Network, netMode int) error {
	dir, name := filepath.Split(path)
	os.MkdirAll(dir, 0755)

	tempFile := filepath.Join(dir, "."+name)
	f, err := os.Create(tempFile)
	if err != nil {
		return err
	}

	// Write out the first usable IP by incrementing
	// sn.IP by one
	sn := bn.Lease().Subnet
	sn.IP++
	if netMode == ipv4 || netMode == (ipv4+ipv6) {
		fmt.Fprintf(f, "FLANNEL_NETWORK=%s\n", nw)
		fmt.Fprintf(f, "FLANNEL_SUBNET=%s\n", sn)
	}

	if nwv6.String() != emptyIPv6Network {
		snv6 := bn.Lease().IPv6Subnet
		snv6.IncrementIP()
		fmt.Fprintf(f, "FLANNEL_IPV6_NETWORK=%s\n", nwv6)
		fmt.Fprintf(f, "FLANNEL_IPV6_SUBNET=%s\n", snv6)
	}

	fmt.Fprintf(f, "FLANNEL_MTU=%d\n", bn.MTU())
	_, err = fmt.Fprintf(f, "FLANNEL_IPMASQ=%v\n", ipMasq)
	f.Close()
	if err != nil {
		return err
	}

	// rename(2) the temporary file to the desired location so that it becomes
	// atomically visible with the contents
	return os.Rename(tempFile, path)
	//TODO - is this safe? What if it's not on the same FS?
}
