package netutil

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

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func GetIPFromInterface(ifaceName string) string {
	ip, err := getIPFromInterface(ifaceName)
	if err != nil {
		logrus.Warn(errors.Wrap(err, "unable to get global unicast ip from interface name"))
	} else {
		logrus.Infof("Found ip %s from iface %s", ip, ifaceName)
	}
	return ip
}

func getIPFromInterface(ifaceName string) (string, error) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return "", err
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return "", err
	}
	if iface.Flags&net.FlagUp == 0 {
		return "", fmt.Errorf("the interface %s is not up", ifaceName)
	}

	globalUnicasts := []string{}
	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			return "", errors.Wrapf(err, "unable to parse CIDR for interface %s", iface.Name)
		}
		// skipping if not ipv4
		if ip.To4() == nil {
			continue
		}
		if ip.IsGlobalUnicast() {
			globalUnicasts = append(globalUnicasts, ip.String())
		}
	}

	if len(globalUnicasts) > 1 {
		return "", fmt.Errorf("multiple global unicast addresses defined for %s, please set ip from one of %v", ifaceName, globalUnicasts)
	}
	if len(globalUnicasts) == 1 {
		return globalUnicasts[0], nil
	}

	return "", fmt.Errorf("can't find ip for interface %s", ifaceName)
}
