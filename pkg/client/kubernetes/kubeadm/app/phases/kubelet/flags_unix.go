//go:build !windows
// +build !windows

package kubelet

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
	"k8s.io/klog/v2"
	utilsexec "k8s.io/utils/exec"

	"github.com/bhojpur/dcp/cmd/grid/client/join/joindata"
	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/constants"
	kubeadmutil "github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/util"
	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/util/initsystem"
)

// buildKubeletArgMap takes a kubeletFlagsOpts object and builds based on that a string-string map with flags
// that should be given to the local Linux kubelet daemon.
func buildKubeletArgMap(data joindata.DcpJoinData) map[string]string {
	kubeletFlags := buildKubeletArgMapCommon(data)

	// TODO: Conditionally set `--cgroup-driver` to either `systemd` or `cgroupfs` for CRI other than Docker
	nodeReg := data.NodeRegistration()
	if nodeReg.CRISocket == constants.DefaultDockerCRISocket {
		driver, err := kubeadmutil.GetCgroupDriverDocker(utilsexec.New())
		if err != nil {
			klog.Warningf("cannot automatically assign a '--cgroup-driver' value when starting the Kubelet: %v\n", err)
		} else {
			kubeletFlags["cgroup-driver"] = driver
		}
	}

	initSystem, err := initsystem.GetInitSystem()
	if err != nil {
		klog.Warningf("cannot get init system: %v\n", err)
		return kubeletFlags
	}

	if initSystem.ServiceIsActive("systemd-resolved") {
		kubeletFlags["resolv-conf"] = "/run/systemd/resolve/resolv.conf"
	}

	return kubeletFlags
}
