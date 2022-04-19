package clusterinfo

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
	"context"
	"fmt"
	"io"
	"os"

	ct "github.com/daviddengcn/go-colortext"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	kubeutil "github.com/bhojpur/dcp/pkg/client/util/kubernetes"
	"github.com/bhojpur/dcp/pkg/projectinfo"
)

// ClusterInfoOptions has the information that required by cluster-info operation
type ClusterInfoOptions struct {
	clientSet    *kubernetes.Clientset
	CloudNodes   []string
	EdgeNodes    []string
	ClusterNodes []string
	OtherNodes   []string
}

// NewClusterInfoOptions creates a new ClusterInfoOptions
func NewClusterInfoOptions() *ClusterInfoOptions {
	return &ClusterInfoOptions{
		CloudNodes:   []string{},
		EdgeNodes:    []string{},
		ClusterNodes: []string{},
		OtherNodes:   []string{},
	}
}

// NewClusterInfoCmd generates a new cluster-info command
func NewClusterInfoCmd() *cobra.Command {
	o := NewClusterInfoOptions()
	cmd := &cobra.Command{
		Use:   "cluster-info",
		Short: "list cloud nodes and edge nodes in cluster",
		Run: func(cmd *cobra.Command, _ []string) {
			if err := o.Complete(cmd.Flags()); err != nil {
				klog.Fatalf("fail to complete the cluster-info option: %s", err)
			}
			if err := o.Run(); err != nil {
				klog.Fatalf("fail to run cluster-info cmd: %s", err)
			}
		},
		Args: cobra.NoArgs,
	}

	return cmd
}

// Complete completes all the required options
func (o *ClusterInfoOptions) Complete(flags *pflag.FlagSet) error {
	var err error
	o.clientSet, err = kubeutil.GenClientSet(flags)
	if err != nil {
		return err
	}
	return nil
}

// Validate makes sure provided values for ClusterInfoOptions are valid
func (o *ClusterInfoOptions) Validate() error {
	return nil
}

func (o *ClusterInfoOptions) Run() (err error) {
	key := projectinfo.GetEdgeWorkerLabelKey()
	Nodes, err := o.clientSet.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, node := range Nodes.Items {
		o.ClusterNodes = append(o.ClusterNodes, node.Name)
		if node.Labels[key] == "false" {
			o.CloudNodes = append(o.CloudNodes, node.Name)
		} else if node.Labels[key] == "true" {
			o.EdgeNodes = append(o.EdgeNodes, node.Name)
		} else {
			o.OtherNodes = append(o.OtherNodes, node.Name)
		}

	}
	printClusterInfo(os.Stdout, "Bhojpur DCP cluster", o.ClusterNodes)
	printClusterInfo(os.Stdout, "Bhojpur DCP cloud", o.CloudNodes)
	printClusterInfo(os.Stdout, "Bhojpur DCP edge", o.EdgeNodes)
	printClusterInfo(os.Stdout, "other", o.OtherNodes)
	return
}

func printClusterInfo(out io.Writer, name string, nodes []string) {
	ct.ChangeColor(ct.Green, false, ct.None, false)
	fmt.Fprint(out, name)
	ct.ResetColor()
	fmt.Fprint(out, " nodes list ")
	ct.ChangeColor(ct.Yellow, false, ct.None, false)
	fmt.Fprint(out, nodes)
	ct.ResetColor()
	fmt.Fprintln(out, "")
}
