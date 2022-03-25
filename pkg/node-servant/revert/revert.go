package revert

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
	"time"

	"github.com/bhojpur/dcp/pkg/engine/util"
	"github.com/bhojpur/dcp/pkg/node-servant/components"
)

// NodeReverter do the revert job
type nodeReverter struct {
	Options
}

// NewReverterWithOptions creates nodeReverter
func NewReverterWithOptions(o *Options) *nodeReverter {
	return &nodeReverter{
		*o,
	}
}

// Do, do the convert job
// shall be implemented as idempotent, can execute multiple times with no side-affect.
func (n *nodeReverter) Do() error {

	if err := n.revertKubelet(); err != nil {
		return err
	}
	if err := n.uninstallEngine(); err != nil {
		return err
	}

	return nil
}

func (n *nodeReverter) revertKubelet() error {
	op := components.NewKubeletOperator(n.dcpDir)
	return op.UndoRedirectTrafficToEngine()
}

func (n *nodeReverter) uninstallEngine() error {
	op := components.NewEngineOperator("", "", "",
		util.WorkingModeCloud, time.Duration(1)) // params is not important here
	return op.UnInstall()
}
