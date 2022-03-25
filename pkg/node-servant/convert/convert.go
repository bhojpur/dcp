package convert

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

	"github.com/bhojpur/dcp/pkg/engine/util"
	"github.com/bhojpur/dcp/pkg/node-servant/components"
)

// NodeConverter do the convert job
type nodeConverter struct {
	Options
}

// NewConverterWithOptions create nodeConverter
func NewConverterWithOptions(o *Options) *nodeConverter {
	return &nodeConverter{
		*o,
	}
}

// Do, do the convert job.
// shall be implemented as idempotent, can execute multiple times with no side-affect.
func (n *nodeConverter) Do() error {
	if err := n.validateOptions(); err != nil {
		return err
	}

	if err := n.installEngine(); err != nil {
		return err
	}
	if err := n.convertKubelet(); err != nil {
		return err
	}

	return nil
}

func (n *nodeConverter) validateOptions() error {
	if !util.IsSupportedWorkingMode(n.workingMode) {
		return fmt.Errorf("workingMode must be pointed out as cloud or edge. got %s", n.workingMode)
	}

	return nil
}

func (n *nodeConverter) installEngine() error {
	apiServerAddress, err := components.GetApiServerAddress(n.kubeadmConfPaths)
	if err != nil {
		return err
	}
	if apiServerAddress == "" {
		return fmt.Errorf("get apiServerAddress empty")
	}
	op := components.NewEngineOperator(apiServerAddress, n.engineImage, n.joinToken,
		n.workingMode, n.engineHealthCheckTimeout)
	return op.Install()
}

func (n *nodeConverter) convertKubelet() error {
	op := components.NewKubeletOperator(n.bhojpurDir)
	return op.RedirectTrafficToEngine()
}
