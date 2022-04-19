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
	"bufio"
	"errors"
	"strings"

	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/cmd/options"
	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/cmd/phases/workflow"
	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/preflight"
)

// NewPreflightPhase creates a kubeadm workflow phase implements preflight checks for reset
func NewPreflightPhase() workflow.Phase {
	return workflow.Phase{
		Name:    "preflight",
		Aliases: []string{"pre-flight"},
		Short:   "Run reset pre-flight checks",
		Long:    "Run pre-flight checks for kubeadm reset.",
		Run:     runPreflight,
		InheritFlags: []string{
			options.IgnorePreflightErrors,
			options.ForceReset,
		},
	}
}

// runPreflight executes preflight checks logic.
func runPreflight(c workflow.RunData) error {
	r, ok := c.(resetData)
	if !ok {
		return errors.New("preflight phase invoked with an invalid data struct")
	}

	if !r.ForceReset() {
		klog.Infoln("[reset] WARNING: Changes made to this host by 'kubeadm init' or 'kubeadm join' will be reverted.")
		klog.Info("[reset] Are you sure you want to proceed? [y/N]: ")
		s := bufio.NewScanner(r.InputReader())
		s.Scan()
		if err := s.Err(); err != nil {
			return err
		}
		if strings.ToLower(s.Text()) != "y" {
			return errors.New("aborted reset operation")
		}
	}

	klog.Infoln("[preflight] Running pre-flight checks")
	return preflight.RunRootCheckOnly(r.IgnorePreflightErrors())
}
