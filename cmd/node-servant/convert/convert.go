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
	"time"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	nodeconverter "github.com/bhojpur/dcp/pkg/node-servant/convert"
)

const (
	// defaultEngineHealthCheckTimeout defines the default timeout for Engine health check phase
	defaultEngineHealthCheckTimeout = 2 * time.Minute
)

// NewConvertCmd generates a new convert command
func NewConvertCmd() *cobra.Command {
	o := nodeconverter.NewConvertOptions()
	cmd := &cobra.Command{
		Use:   "convert --working-mode",
		Short: "",
		Run: func(cmd *cobra.Command, args []string) {
			if err := o.Complete(cmd.Flags()); err != nil {
				klog.Fatalf("fail to complete the convert option: %s", err)
			}

			converter := nodeconverter.NewConverterWithOptions(o)
			if err := converter.Do(); err != nil {
				klog.Fatalf("fail to convert the kubernetes node to a Bhojpur DCP node: %s", err)
			}
			klog.Info("convert success")
		},
		Args: cobra.NoArgs,
	}
	setFlags(cmd)

	return cmd
}

// setFlags sets flags.
func setFlags(cmd *cobra.Command) {
	cmd.Flags().String("engine-image", "bhojpur/dcpsvr:latest",
		"The Bhojpur DCP Engine image.")
	cmd.Flags().Duration("engine-healthcheck-timeout", defaultEngineHealthCheckTimeout,
		"The timeout for Bhojpur DCP engine health check.")
	cmd.Flags().StringP("kubeadm-conf-path", "k", "",
		"The path to kubelet service conf that is used by kubelet component to join the cluster on the work node."+
			"Support multiple values, will search in order until get the file.(e.g -k kbcfg1,kbcfg2)",
	)
	cmd.Flags().String("join-token", "", "The token used by Bhojpur DCP engine for joining the cluster.")
	cmd.Flags().String("working-mode", "edge", "The node type cloud/edge, effect Bhojpur DCP engine workingMode.")
}
