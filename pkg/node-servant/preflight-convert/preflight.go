package preflight_convert

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

	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/pkg/preflight"
)

// ConvertPreflighter do the preflight-convert-convert job
type ConvertPreflighter struct {
	Options
}

// NewPreflighterWithOptions create nodePreflighter
func NewPreflighterWithOptions(o *Options) *ConvertPreflighter {
	return &ConvertPreflighter{
		*o,
	}
}

func (n *ConvertPreflighter) Do() error {
	klog.Infof("[preflight-convert] Running node-servant pre-flight checks")
	if err := preflight.RunConvertNodeChecks(n, n.IgnorePreflightErrors, n.DeployTunnel); err != nil {
		return err
	}

	fmt.Println("[preflight-convert] Pulling images required for converting a Kubernetes cluster to a Bhojpur DCP cluster")
	fmt.Println("[preflight-convert] This might take a minute or two, depending on the speed of your internet connection")
	if err := preflight.RunPullImagesCheck(n, n.IgnorePreflightErrors); err != nil {
		return err
	}

	return nil
}
