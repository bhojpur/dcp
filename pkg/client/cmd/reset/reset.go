package reset

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
	"io"
	"strings"

	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"

	dcpphases "github.com/bhojpur/dcp/pkg/client/cmd/reset/phases"
	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/cmd/options"
	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/cmd/phases/workflow"
	utilruntime "github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/util/runtime"
)

var (
	iptablesCleanupInstructions = dedent.Dedent(`
		The reset process does not reset or clean up iptables rules or IPVS tables.
		If you wish to reset iptables, you must do so manually by using the "iptables" command.

		If your cluster was setup to utilize IPVS, run ipvsadm --clear (or similar)
		to reset your system's IPVS tables.

		The reset process does not clean your kubeconfig files and you must remove them manually.
		Please, check the contents of the $HOME/.kube/config file.
	`)

	cniCleanupInstructions = dedent.Dedent(`
		The reset process does not clean CNI configuration. To do so, you must remove /etc/cni/net.d
	`)
)

// resetOptions defines all the options exposed via flags by kubeadm reset.
type resetOptions struct {
	criSocketPath         string
	forceReset            bool
	ignorePreflightErrors []string
}

// resetData defines all the runtime information used when running the kubeadm reset workflow;
// this data is shared across all the phases that are included in the workflow.
type resetData struct {
	criSocketPath         string
	forceReset            bool
	ignorePreflightErrors sets.String
	inputReader           io.Reader
	outputWriter          io.Writer
	dirsToClean           []string
}

// newResetOptions returns a struct ready for being used for creating cmd join flags.
func newResetOptions() *resetOptions {
	return &resetOptions{
		forceReset: false,
	}
}

// newResetData returns a new resetData struct to be used for the execution of the kubeadm reset workflow.
func newResetData(cmd *cobra.Command, options *resetOptions, in io.Reader, out io.Writer) (*resetData, error) {
	var criSocketPath string
	var err error
	if options.criSocketPath == "" {
		criSocketPath, err = utilruntime.DetectCRISocket()
		if err != nil {
			return nil, err
		}
		klog.V(1).Infof("[reset] Detected and using CRI socket: %s", criSocketPath)
	} else {
		criSocketPath = options.criSocketPath
		klog.V(1).Infof("[reset] Using specified CRI socket: %s", criSocketPath)
	}

	var ignoreErrors sets.String
	for _, item := range options.ignorePreflightErrors {
		ignoreErrors.Insert(strings.ToLower(item))
	}

	return &resetData{
		criSocketPath:         criSocketPath,
		forceReset:            options.forceReset,
		ignorePreflightErrors: ignoreErrors,
		inputReader:           in,
		outputWriter:          out,
	}, nil
}

// AddResetFlags adds reset flags
func AddResetFlags(flagSet *flag.FlagSet, resetOptions *resetOptions) {
	flagSet.BoolVarP(
		&resetOptions.forceReset, options.ForceReset, "f", false,
		"Reset the node without prompting for confirmation.",
	)
	flagSet.StringVar(
		&resetOptions.criSocketPath, options.NodeCRISocket, resetOptions.criSocketPath,
		"Path to the CRI socket to connect",
	)
	flagSet.StringSliceVar(
		&resetOptions.ignorePreflightErrors, options.IgnorePreflightErrors, resetOptions.ignorePreflightErrors,
		"A list of checks whose errors will be shown as warnings. Example: 'IsPrivilegedUser,Swap'. Value 'all' ignores errors from all checks.",
	)
}

// NewCmdReset returns the "dcpctl reset" command
func NewCmdReset(in io.Reader, out io.Writer, resetOptions *resetOptions) *cobra.Command {
	if resetOptions == nil {
		resetOptions = newResetOptions()
	}
	resetRunner := workflow.NewRunner()

	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Performs a best effort revert of changes made to this host by 'dcpctl join'",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := resetRunner.InitData(args)
			if err != nil {
				return err
			}

			err = resetRunner.Run(args)
			if err != nil {
				return err
			}

			// Then clean contents from the stateful kubelet, etcd and cni directories
			data := c.(*resetData)
			cleanDirs(data)

			// output help text instructing user how to remove cni folders
			klog.Info(cniCleanupInstructions)
			// Output help text instructing user how to remove iptables rules
			klog.Info(iptablesCleanupInstructions)
			return nil
		},
	}

	AddResetFlags(cmd.Flags(), resetOptions)

	// initialize the workflow runner with the list of phases
	resetRunner.AppendPhase(dcpphases.NewPreflightPhase())
	resetRunner.AppendPhase(dcpphases.NewCleanupNodePhase())
	resetRunner.AppendPhase(dcpphases.NewCleanFilePhase())

	// sets the data builder function, that will be used by the runner
	// both when running the entire workflow or single phases
	resetRunner.SetDataInitializer(func(cmd *cobra.Command, args []string) (workflow.RunData, error) {
		return newResetData(cmd, resetOptions, in, out)
	})

	// binds the Runner to kubeadm init command by altering
	// command help, adding --skip-phases flag and by adding phases subcommands
	resetRunner.BindToCommand(cmd)

	return cmd
}

func cleanDirs(data *resetData) {
	klog.Infof("[reset] Deleting contents of stateful directories: %v\n", data.dirsToClean)
	for _, dir := range data.dirsToClean {
		klog.V(1).Infof("[reset] Deleting contents of %s", dir)
		if err := dcpphases.CleanDir(dir); err != nil {
			klog.Warningf("[reset] Failed to delete contents of %q directory: %v", dir, err)
		}
	}
}

// ForceReset returns the forceReset flag.
func (r *resetData) ForceReset() bool {
	return r.forceReset
}

// InputReader returns the io.reader used to read messages.
func (r *resetData) InputReader() io.Reader {
	return r.inputReader
}

// IgnorePreflightErrors returns the list of preflight errors to ignore.
func (r *resetData) IgnorePreflightErrors() sets.String {
	return r.ignorePreflightErrors
}

// AddDirsToClean add a list of dirs to the list of dirs that will be removed.
func (r *resetData) AddDirsToClean(dirs ...string) {
	r.dirsToClean = append(r.dirsToClean, dirs...)
}

// CRISocketPath returns the criSocketPath.
func (r *resetData) CRISocketPath() string {
	return r.criSocketPath
}
