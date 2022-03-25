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
	"strings"

	"github.com/vishvananda/netlink"
	"k8s.io/klog/v2"
)

type DummyInterfaceController interface {
	EnsureDummyInterface(ifName string, ifIP net.IP) error
	DeleteDummyInterface(ifName string) error
	ListDummyInterface(ifName string) ([]net.IP, error)
}

type dummyInterfaceController struct {
	netlink.Handle
}

// NewDummyInterfaceManager returns an instance for create/delete dummy net interface
func NewDummyInterfaceController() DummyInterfaceController {
	return &dummyInterfaceController{
		Handle: netlink.Handle{},
	}
}

// EnsureDummyInterface make sure the dummy net interface with specified name and ip exist
func (dic *dummyInterfaceController) EnsureDummyInterface(ifName string, ifIP net.IP) error {
	link, err := dic.LinkByName(ifName)
	if err == nil {
		addrs, err := dic.AddrList(link, 0)
		if err != nil {
			return err
		}

		for _, addr := range addrs {
			if addr.IP != nil && addr.IP.Equal(ifIP) {
				return nil
			}
		}

		klog.Infof("ip address for %s interface changed to %s", ifName, ifIP.String())
		return dic.AddrReplace(link, &netlink.Addr{IPNet: netlink.NewIPNet(ifIP)})
	}

	if strings.Contains(err.Error(), "Link not found") && link == nil {
		return dic.addDummyInterface(ifName, ifIP)
	}

	return err
}

// addDummyInterface creates a dummy net interface with the specified name and ip
func (dic *dummyInterfaceController) addDummyInterface(ifName string, ifIP net.IP) error {
	_, err := dic.LinkByName(ifName)
	if err == nil {
		return fmt.Errorf("Link %s exists", ifName)
	}

	dummy := &netlink.Dummy{
		LinkAttrs: netlink.LinkAttrs{Name: ifName},
	}
	err = dic.LinkAdd(dummy)
	if err != nil {
		return err
	}

	link, err := dic.LinkByName(ifName)
	if err != nil {
		return err
	}
	return dic.AddrAdd(link, &netlink.Addr{IPNet: netlink.NewIPNet(ifIP)})
}

// DeleteDummyInterface delete the dummy net interface with specified name
func (dic *dummyInterfaceController) DeleteDummyInterface(ifName string) error {
	link, err := dic.LinkByName(ifName)
	if err != nil {
		return err
	}
	return dic.LinkDel(link)
}

// ListDummyInterface list all ips for network interface specified by ifName
func (dic *dummyInterfaceController) ListDummyInterface(ifName string) ([]net.IP, error) {
	ips := make([]net.IP, 0)
	link, err := dic.LinkByName(ifName)
	if err != nil {
		return ips, err
	}

	addrs, err := dic.AddrList(link, 0)
	if err != nil {
		return ips, err
	}

	for _, addr := range addrs {
		if addr.IP != nil {
			ips = append(ips, addr.IP)
		}
	}

	return ips, nil
}
