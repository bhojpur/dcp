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
	"os"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	preflightconvert "github.com/bhojpur/dcp/pkg/node-servant/preflight-convert"
)

const (
	latestEngineImage      = "bhojpur/dcpsvr:latest"
	latestTunnelAgentImage = "bhojpur/tunnel-agent:latest"
)

// NewxPreflightConvertCmd generates a new preflight-convert check command
func NewxPreflightConvertCmd() *cobra.Command {
	o := preflightconvert.NewPreflightConvertOptions()
	cmd := &cobra.Command{
		Use:   "preflight-convert",
		Short: "",
		Run: func(cmd *cobra.Command, args []string) {
			if err := o.Complete(cmd.Flags()); err != nil {
				klog.Errorf("Fail to complete the preflight-convert option: %s", err)
				os.Exit(1)
			}
			preflighter := preflightconvert.NewPreflighterWithOptions(o)
			if err := preflighter.Do(); err != nil {
				klog.Errorf("Fail to run pre-flight checks: %s", err)
				os.Exit(1)
			}
			klog.Info("convert pre-flight checks success")
		},
		Args: cobra.NoArgs,
	}
	setFlags(cmd)

	return cmd
}

func setFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("kubeadm-conf-path", "k", "",
		"The path to kubelet service conf that is used by kubelet component to join the cluster on the work node."+
			"Support multiple values, will search in order until get the file.(e.g -k kbcfg1,kbcfg2)",
	)
	cmd.Flags().String("engine-image", latestEngineImage, "The Bhojpur DCP engine image.")
	cmd.Flags().String("tunnel-agent-image", latestTunnelAgentImage, "The tunnel-agent image.")
	cmd.Flags().BoolP("deploy-tunnel", "t", false, "If set, tunnel-agent will be deployed.")
	cmd.Flags().String("ignore-preflight-errors", "", "A list of checks whose errors will be shown as warnings. "+
		"Example: 'isprivilegeduser,imagepull'.Value 'all' ignores errors from all checks.",
	)
}
