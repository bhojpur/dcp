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
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/pkg/node-servant/revert"
)

// NewRevertCmd generates a new revert command
func NewRevertCmd() *cobra.Command {
	o := revert.NewRevertOptions()
	cmd := &cobra.Command{
		Use:   "revert",
		Short: "",
		Run: func(cmd *cobra.Command, args []string) {
			if err := o.Complete(cmd.Flags()); err != nil {
				klog.Fatalf("fail to complete the revert option: %s", err)
			}

			r := revert.NewReverterWithOptions(o)
			if err := r.Do(); err != nil {
				klog.Fatalf("fail to revert the Bhojpur DCP node to a kubernetes node: %s", err)
			}
			klog.Info("revert success")
		},
		Args: cobra.NoArgs,
	}
	setFlags(cmd)

	return cmd
}

// setFlags sets flags.
func setFlags(cmd *cobra.Command) {
	cmd.Flags().String("kubeadm-conf-path", "",
		"The path to kubelet service conf that is used by kubelet component to join the cluster on the edge node.")
}
