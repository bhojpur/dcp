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
	"strings"

	"github.com/spf13/pflag"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/bhojpur/dcp/pkg/node-servant/components"
)

const (
	kubeAdmFlagsEnvFile = "/var/lib/kubelet/kubeadm-flags.env"
)

// Options has the information that required by preflight-convert operation
type Options struct {
	KubeadmConfPaths      []string
	EngineImage           string
	TunnelAgentImage      string
	DeployTunnel          bool
	IgnorePreflightErrors sets.String

	KubeAdmFlagsEnvFile string
	ImagePullPolicy     v1.PullPolicy
	CRISocket           string
}

func (o *Options) GetCRISocket() string {
	return o.CRISocket
}

func (o *Options) GetImageList() []string {
	imgs := []string{}

	imgs = append(imgs, o.EngineImage)
	if o.DeployTunnel {
		imgs = append(imgs, o.TunnelAgentImage)
	}
	return imgs
}

func (o *Options) GetImagePullPolicy() v1.PullPolicy {
	return o.ImagePullPolicy
}

func (o *Options) GetKubeadmConfPaths() []string {
	return o.KubeadmConfPaths
}

func (o *Options) GetKubeAdmFlagsEnvFile() string {
	return o.KubeAdmFlagsEnvFile
}

// NewPreflightConvertOptions creates a new Options
func NewPreflightConvertOptions() *Options {
	return &Options{
		KubeadmConfPaths:      components.GetDefaultKubeadmConfPath(),
		IgnorePreflightErrors: sets.NewString(),
		KubeAdmFlagsEnvFile:   kubeAdmFlagsEnvFile,
		ImagePullPolicy:       v1.PullIfNotPresent,
	}
}

// Complete completes all the required options.
func (o *Options) Complete(flags *pflag.FlagSet) error {

	kubeadmConfPaths, err := flags.GetString("kubeadm-conf-path")
	if err != nil {
		return err
	}
	if kubeadmConfPaths != "" {
		o.KubeadmConfPaths = strings.Split(kubeadmConfPaths, ",")
	}

	engineImage, err := flags.GetString("dcpsvr-image")
	if err != nil {
		return err
	}
	o.EngineImage = engineImage

	tunnelAgentImage, err := flags.GetString("tunnel-agent-image")
	if err != nil {
		return err
	}
	o.TunnelAgentImage = tunnelAgentImage

	dt, err := flags.GetBool("deploy-tunnel")
	if err != nil {
		return err
	}
	o.DeployTunnel = dt

	ipStr, err := flags.GetString("ignore-preflight-errors")
	if err != nil {
		return err
	}
	if ipStr != "" {
		ipStr = strings.ToLower(ipStr)
		o.IgnorePreflightErrors = sets.NewString(strings.Split(ipStr, ",")...)
	}

	CRISocket, err := components.DetectCRISocket()
	if err != nil {
		return err
	}
	o.CRISocket = CRISocket
	return nil
}
