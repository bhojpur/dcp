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

	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/pkg/client/constants"
	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/cmd/phases/workflow"
	"github.com/bhojpur/dcp/pkg/client/util/edgenode"
)

func NewCleanFilePhase() workflow.Phase {
	return workflow.Phase{
		Name:  "Clean up the directories and files related to Bhojpur DCP.",
		Short: "Clean up the directories and files related to Bhojpur DCP.",
		Run:   runCleanfile,
	}
}

func runCleanfile(c workflow.RunData) error {
	for _, comp := range []string{"kubectl", "kubeadm", "kubelet"} {
		target := fmt.Sprintf("/usr/bin/%s", comp)
		if err := os.RemoveAll(target); err != nil {
			klog.Warningf("Clean file %s fail: %v, please clean it manually.", target, err)
		}
	}

	for _, file := range []string{constants.KubeletWorkdir,
		constants.TunnelAgentWorkdir,
		constants.TunnelServerWorkdir,
		constants.EngineWorkdir,
		edgenode.KubeletSvcPath,
		constants.KubeletServiceFilepath,
		constants.KubeCniDir,
		constants.KubeletConfigureDir,
		constants.SysctlK8sConfig} {
		if err := os.RemoveAll(file); err != nil {
			klog.Warningf("Clean file %s fail: %v, please clean it manually.", file, err)
		}
	}
	return nil
}
