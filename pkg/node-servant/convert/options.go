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
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"

	enutil "github.com/bhojpur/dcp/pkg/client/util/edgenode"
	hubutil "github.com/bhojpur/dcp/pkg/engine/util"
	"github.com/bhojpur/dcp/pkg/node-servant/components"
)

// Options has the information that required by convert operation
type Options struct {
	engineImage              string
	engineHealthCheckTimeout time.Duration
	workingMode              hubutil.WorkingMode

	joinToken        string
	kubeadmConfPaths []string
	bhojpurDir       string
}

// NewConvertOptions creates a new Options
func NewConvertOptions() *Options {
	return &Options{
		kubeadmConfPaths: components.GetDefaultKubeadmConfPath(),
	}
}

// Complete completes all the required options.
func (o *Options) Complete(flags *pflag.FlagSet) error {
	engineImage, err := flags.GetString("dcpsvr-image")
	if err != nil {
		return err
	}
	o.engineImage = engineImage

	engineHealthCheckTimeout, err := flags.GetDuration("dcpsvr-healthcheck-timeout")
	if err != nil {
		return err
	}
	o.engineHealthCheckTimeout = engineHealthCheckTimeout

	kubeadmConfPaths, err := flags.GetString("kubeadm-conf-path")
	if err != nil {
		return err
	}
	if kubeadmConfPaths != "" {
		o.kubeadmConfPaths = strings.Split(kubeadmConfPaths, ",")
	}

	joinToken, err := flags.GetString("join-token")
	if err != nil {
		return err
	}
	if joinToken == "" {
		return fmt.Errorf("get joinToken empty")
	}
	o.joinToken = joinToken

	bhojpurDir := os.Getenv("BHOJPUR_DCP_DIR")
	if bhojpurDir == "" {
		bhojpurDir = enutil.BhojpurDir
	}
	o.bhojpurDir = bhojpurDir

	workingMode, err := flags.GetString("working-mode")
	if err != nil {
		return err
	}

	wm := hubutil.WorkingMode(workingMode)
	if !hubutil.IsSupportedWorkingMode(wm) {
		return fmt.Errorf("invalid working mode: %s", workingMode)
	}
	o.workingMode = wm

	return nil
}
