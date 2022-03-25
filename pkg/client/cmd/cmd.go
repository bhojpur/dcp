package cmd

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
	goflag "flag"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/pkg/client/cmd/clusterinfo"
	"github.com/bhojpur/dcp/pkg/client/cmd/convert"
	"github.com/bhojpur/dcp/pkg/client/cmd/dcpinit"
	"github.com/bhojpur/dcp/pkg/client/cmd/join"
	"github.com/bhojpur/dcp/pkg/client/cmd/markautonomous"
	"github.com/bhojpur/dcp/pkg/client/cmd/reset"
	"github.com/bhojpur/dcp/pkg/client/cmd/revert"
	"github.com/bhojpur/dcp/pkg/projectinfo"
)

// NewClientCommand creates a new Client command
func NewClientCommand() *cobra.Command {
	version := fmt.Sprintf("%#v", projectinfo.Get())
	cmds := &cobra.Command{
		Use:     "dcpctl",
		Short:   "dcpctl controls the distributed cloud platform cluster",
		Version: version,
	}

	setVersion(cmds)
	// add kubeconfig to persistent flags
	cmds.PersistentFlags().String("kubeconfig", "", "The path to the kubeconfig file")
	cmds.AddCommand(convert.NewConvertCmd())
	cmds.AddCommand(revert.NewRevertCmd())
	cmds.AddCommand(markautonomous.NewMarkAutonomousCmd())
	cmds.AddCommand(clusterinfo.NewClusterInfoCmd())
	cmds.AddCommand(dcpinit.NewCmdInit())
	cmds.AddCommand(join.NewCmdJoin(os.Stdout, nil))
	cmds.AddCommand(reset.NewCmdReset(os.Stdin, os.Stdout, nil))

	klog.InitFlags(nil)
	// goflag.Parse()
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	return cmds
}

func setVersion(cmd *cobra.Command) {
	cmd.SetVersionTemplate(`{{with .Name}}{{printf "%s " .}}{{end}}{{printf "version: %s" .Version}}`)
}
