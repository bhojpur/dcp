package markautonomous

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
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/pkg/client/constants"
	"github.com/bhojpur/dcp/pkg/client/lock"
	kubeutil "github.com/bhojpur/dcp/pkg/client/util/kubernetes"
	"github.com/bhojpur/dcp/pkg/projectinfo"
)

// MarkAutonomousOptions has the information that required by convert operation
type MarkAutonomousOptions struct {
	*kubernetes.Clientset
	AutonomousNodes  []string
	MarkAllEdgeNodes bool
}

// NewMarkAutonomousOptions creates a new MarkAutonomousOptions
func NewMarkAutonomousOptions() *MarkAutonomousOptions {
	return &MarkAutonomousOptions{}
}

// NewMarkAutonomousCmd generates a new markautonomous command
func NewMarkAutonomousCmd() *cobra.Command {
	co := NewMarkAutonomousOptions()
	cmd := &cobra.Command{
		Use:   "markautonomous -a AUTONOMOUSNODES",
		Short: "mark the nodes as autonomous",
		Run: func(cmd *cobra.Command, _ []string) {
			if err := co.Complete(cmd.Flags()); err != nil {
				klog.Fatalf("fail to complete the markautonomous option: %s", err)
			}
			if err := co.RunMarkAutonomous(); err != nil {
				klog.Fatalf("fail to make nodes autonomous: %s", err)
			}
		},
		Args: cobra.NoArgs,
	}

	cmd.Flags().StringP("autonomous-nodes", "a", "",
		"The list of nodes that will be marked as autonomous. If not set, all edge nodes will be marked as autonomous."+
			"(e.g. -a autonomousnode1,autonomousnode2)")

	return cmd
}

// Complete completes all the required options
func (mao *MarkAutonomousOptions) Complete(flags *pflag.FlagSet) error {
	anStr, err := flags.GetString("autonomous-nodes")
	if err != nil {
		return err
	}
	if anStr == "" {
		mao.AutonomousNodes = []string{}
	} else {
		mao.AutonomousNodes = strings.Split(anStr, ",")
	}

	// set mark-all-edge-node to false, as user has specified autonomous nodes
	if len(mao.AutonomousNodes) == 0 {
		mao.MarkAllEdgeNodes = true
	}

	mao.Clientset, err = kubeutil.GenClientSet(flags)
	if err != nil {
		return err
	}

	return nil
}

// RunMarkAutonomous annotates specified edge nodes as autonomous
func (mao *MarkAutonomousOptions) RunMarkAutonomous() (err error) {
	if err = lock.AcquireLock(mao.Clientset); err != nil {
		return
	}
	defer func() {
		err = lock.ReleaseLock(mao.Clientset)
	}()
	var (
		autonomousNodes []*v1.Node
		edgeNodeList    *v1.NodeList
	)
	if mao.MarkAllEdgeNodes {
		// make all edge nodes autonomous
		labelSelector := fmt.Sprintf("%s=true", projectinfo.GetEdgeWorkerLabelKey())
		edgeNodeList, err = mao.CoreV1().Nodes().
			List(context.Background(), metav1.ListOptions{LabelSelector: labelSelector})
		if err != nil {
			return
		}
		if len(edgeNodeList.Items) == 0 {
			klog.Warning("there is no edge nodes, please label the edge node first")
			return
		}
		for i := range edgeNodeList.Items {
			autonomousNodes = append(autonomousNodes, &edgeNodeList.Items[i])
		}
	} else {
		// make only the specified edge nodes autonomous
		for _, nodeName := range mao.AutonomousNodes {
			var node *v1.Node
			node, err = mao.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
			if err != nil {
				return
			}
			if node.Labels[projectinfo.GetEdgeWorkerLabelKey()] == "false" {
				err = fmt.Errorf("can't make cloud node(%s) autonomous",
					node.GetName())
				return
			}
			autonomousNodes = append(autonomousNodes, node)
		}
	}

	for _, anode := range autonomousNodes {
		klog.Infof("mark %s as autonomous", anode.GetName())
		if _, err = kubeutil.AnnotateNode(mao.Clientset,
			anode, constants.AnnotationAutonomy, "true"); err != nil {
			return
		}
	}

	return
}
