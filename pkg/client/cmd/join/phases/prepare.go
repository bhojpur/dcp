package phases

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
	"os"
	"path/filepath"

	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/pkg/client/cmd/join/joindata"
	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/cmd/options"
	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/cmd/phases/workflow"
	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/constants"
	"github.com/bhojpur/dcp/pkg/client/util/kubernetes"
	"github.com/bhojpur/dcp/pkg/client/util/system"
)

// NewPreparePhase creates a client workflow phase that initialize the node environment.
func NewPreparePhase() workflow.Phase {
	return workflow.Phase{
		Name:  "Initialize system environment.",
		Short: "Initialize system environment.",
		Run:   runPrepare,
		InheritFlags: []string{
			options.TokenStr,
		},
	}
}

//runPrepare executes the node initialization process.
func runPrepare(c workflow.RunData) error {
	data, ok := c.(joindata.DcpJoinData)
	if !ok {
		return fmt.Errorf("Prepare phase invoked with an invalid data struct. ")
	}

	// cleanup at first
	staticPodsPath := filepath.Join(constants.KubernetesDir, constants.ManifestsSubDirName)
	if err := os.RemoveAll(staticPodsPath); err != nil {
		klog.Warningf("remove %s: %v", staticPodsPath, err)
	}

	if err := system.SetIpv4Forward(); err != nil {
		return err
	}
	if err := system.SetBridgeSetting(); err != nil {
		return err
	}
	if err := system.SetSELinux(); err != nil {
		return err
	}
	if err := kubernetes.CheckAndInstallKubelet(data.KubernetesResourceServer(), data.KubernetesVersion()); err != nil {
		return err
	}
	if err := kubernetes.SetKubeletService(); err != nil {
		return err
	}
	if err := kubernetes.SetKubeletUnitConfig(); err != nil {
		return err
	}
	if err := kubernetes.SetKubeletConfigForNode(); err != nil {
		return err
	}
	if err := kubernetes.SetKubeletCaCert(data.TLSBootstrapCfg()); err != nil {
		return err
	}
	return nil
}
