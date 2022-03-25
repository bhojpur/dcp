package system

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
	"io/ioutil"
	"os/exec"

	"github.com/opencontainers/selinux/go-selinux"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/pkg/client/constants"
	"github.com/bhojpur/dcp/pkg/client/util/edgenode"
)

const (
	ip_forward = "/proc/sys/net/ipv4/ip_forward"
	bridgenf   = "/proc/sys/net/bridge/bridge-nf-call-iptables"
	bridgenf6  = "/proc/sys/net/bridge/bridge-nf-call-ip6tables"

	kubernetsBridgeSetting = `
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1`
)

//setIpv4Forward turn on the node ipv4 forward.
func SetIpv4Forward() error {
	klog.Infof("Setting ipv4 forward")
	if err := ioutil.WriteFile(ip_forward, []byte("1"), 0644); err != nil {
		return fmt.Errorf("Write content 1 to file %s fail: %v ", ip_forward, err)
	}
	return nil
}

//setBridgeSetting turn on the node bridge-nf-call-iptables.
func SetBridgeSetting() error {
	klog.Info("Setting bridge settings for kubernetes.")
	if err := ioutil.WriteFile(constants.SysctlK8sConfig, []byte(kubernetsBridgeSetting), 0644); err != nil {
		return fmt.Errorf("Write file %s fail: %v ", constants.SysctlK8sConfig, err)
	}

	if exist, _ := edgenode.FileExists(bridgenf); !exist {
		cmd := exec.Command("bash", "-c", "modprobe br-netfilter")
		if err := edgenode.Exec(cmd); err != nil {
			return err
		}
	}
	if err := ioutil.WriteFile(bridgenf, []byte("1"), 0644); err != nil {
		return fmt.Errorf("Write file %s fail: %v ", bridgenf, err)
	}
	if err := ioutil.WriteFile(bridgenf6, []byte("1"), 0644); err != nil {
		return fmt.Errorf("Write file %s fail: %v ", bridgenf, err)
	}
	return nil
}

// setSELinux turn off the node selinux.
func SetSELinux() error {
	klog.Info("Disabling SELinux.")
	selinux.SetDisabled()
	return nil
}
