package join

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
	"io"
	"os"
	"strings"

	"github.com/lithammer/dedent"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/sets"
	clientset "k8s.io/client-go/kubernetes"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/pkg/client/cmd/join/joindata"
	dcpphase "github.com/bhojpur/dcp/pkg/client/cmd/join/phases"
	dcpconstants "github.com/bhojpur/dcp/pkg/client/constants"
	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/cmd/options"
	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/cmd/phases/workflow"
	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/constants"
	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/discovery/token"
	kubeconfigutil "github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/util/kubeconfig"
	clientutil "github.com/bhojpur/dcp/pkg/client/util/kubernetes"
)

var (
	joinWorkerNodeDoneMsg = dedent.Dedent(`
		This node has joined the cluster:
		* Certificate signing request was sent to apiserver and a response was received.
		* The Kubelet was informed of the new secure connection details.

		Run 'kubectl get nodes' on the control-plane to see this node join the cluster.

		`)
)

type joinOptions struct {
	token                    string
	nodeType                 string
	nodeName                 string
	criSocket                string
	organizations            string
	pauseImage               string
	engineImage              string
	caCertHashes             []string
	unsafeSkipCAVerification bool
	ignorePreflightErrors    []string
	nodeLabels               string
	kubernetesResourceServer string
}

// newJoinOptions returns a struct ready for being used for creating cmd join flags.
func newJoinOptions() *joinOptions {
	return &joinOptions{
		nodeType:                 dcpconstants.EdgeNode,
		criSocket:                constants.DefaultDockerCRISocket,
		pauseImage:               dcpconstants.PauseImagePath,
		engineImage:              fmt.Sprintf("%s/%s:%s", dcpconstants.DefaultDcpImageRegistry, dcpconstants.Engine, dcpconstants.DefaultDcpVersion),
		caCertHashes:             make([]string, 0),
		unsafeSkipCAVerification: false,
		ignorePreflightErrors:    make([]string, 0),
		kubernetesResourceServer: dcpconstants.DefaultKubernetesResourceServer,
	}
}

// NewCmdJoin returns "dcpctl join" command.
func NewCmdJoin(out io.Writer, joinOptions *joinOptions) *cobra.Command {
	if joinOptions == nil {
		joinOptions = newJoinOptions()
	}
	joinRunner := workflow.NewRunner()

	cmd := &cobra.Command{
		Use:   "join [api-server-endpoint]",
		Short: "Run this on any machine you wish to join an existing cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := joinRunner.Run(args); err != nil {
				return err
			}
			fmt.Fprint(out, joinWorkerNodeDoneMsg)
			return nil
		},
	}

	addJoinConfigFlags(cmd.Flags(), joinOptions)

	joinRunner.AppendPhase(dcpphase.NewPreparePhase())
	joinRunner.AppendPhase(dcpphase.NewPreflightPhase())
	joinRunner.AppendPhase(dcpphase.NewEdgeNodePhase())
	joinRunner.AppendPhase(dcpphase.NewPostcheckPhase())
	joinRunner.SetDataInitializer(func(cmd *cobra.Command, args []string) (workflow.RunData, error) {
		return newJoinData(cmd, args, joinOptions, out)
	})
	joinRunner.BindToCommand(cmd)
	return cmd
}

// addJoinConfigFlags adds join flags bound to the config to the specified flagset
func addJoinConfigFlags(flagSet *flag.FlagSet, joinOptions *joinOptions) {
	flagSet.StringVar(
		&joinOptions.token, options.TokenStr, "",
		"Use this token for both discovery-token and tls-bootstrap-token when those values are not provided.",
	)
	flagSet.StringVar(
		&joinOptions.nodeType, options.NodeType, joinOptions.nodeType,
		"Sets the node is edge or cloud",
	)
	flagSet.StringVar(
		&joinOptions.nodeName, options.NodeName, joinOptions.nodeName,
		`Specify the node name. if not specified, hostname will be used.`,
	)
	flagSet.StringVar(
		&joinOptions.criSocket, options.NodeCRISocket, joinOptions.criSocket,
		"Path to the CRI socket to connect",
	)
	flagSet.StringVar(
		&joinOptions.organizations, options.Organizations, joinOptions.organizations,
		"Organizations that will be added into hub's client certificate",
	)
	flagSet.StringVar(
		&joinOptions.pauseImage, options.PauseImage, joinOptions.pauseImage,
		"Sets the image version of pause container",
	)
	flagSet.StringVar(
		&joinOptions.engineImage, options.EngineImage, joinOptions.engineImage,
		"Sets the image version of Bhojpur DCP server engine component",
	)
	flagSet.StringSliceVar(
		&joinOptions.caCertHashes, options.TokenDiscoveryCAHash, joinOptions.caCertHashes,
		"For token-based discovery, validate that the root CA public key matches this hash (format: \"<type>:<value>\").",
	)
	flagSet.BoolVar(
		&joinOptions.unsafeSkipCAVerification, options.TokenDiscoverySkipCAHash, false,
		"For token-based discovery, allow joining without --discovery-token-ca-cert-hash pinning.",
	)
	flagSet.StringSliceVar(
		&joinOptions.ignorePreflightErrors, options.IgnorePreflightErrors, joinOptions.ignorePreflightErrors,
		"A list of checks whose errors will be shown as warnings. Example: 'IsPrivilegedUser,Swap'. Value 'all' ignores errors from all checks.",
	)
	flagSet.StringVar(
		&joinOptions.nodeLabels, options.NodeLabels, joinOptions.nodeLabels,
		"Sets the labels for joining node",
	)
	flagSet.StringVar(
		&joinOptions.kubernetesResourceServer, options.KubernetesResourceServer, joinOptions.kubernetesResourceServer,
		"Sets the address for downloading k8s node resources",
	)
}

type joinData struct {
	joinNodeData             *joindata.NodeRegistration
	apiServerEndpoint        string
	token                    string
	tlsBootstrapCfg          *clientcmdapi.Config
	clientSet                *clientset.Clientset
	ignorePreflightErrors    sets.String
	organizations            string
	pauseImage               string
	engineImage              string
	kubernetesVersion        string
	caCertHashes             sets.String
	nodeLabels               map[string]string
	kubernetesResourceServer string
}

// newJoinData returns a new joinData struct to be used for the execution of the kubeadm join workflow.
// This func takes care of validating joinOptions passed to the command, and then it converts
// options into the internal JoinData type that is used as input all the phases in the kubeadm join workflow
func newJoinData(cmd *cobra.Command, args []string, opt *joinOptions, out io.Writer) (*joinData, error) {
	// if an APIServerEndpoint from which to retrieve cluster information was not provided, unset the Discovery.BootstrapToken object
	var apiServerEndpoint string
	if len(args) == 0 {
		return nil, errors.New("apiServer endpoint is empty")
	} else {
		if len(args) > 1 {
			klog.Warningf("[preflight] WARNING: More than one API server endpoint supplied on command line %v. Using the first one.", args)
		}
		apiServerEndpoint = args[0]
	}

	if len(opt.token) == 0 {
		return nil, errors.New("join token is empty, so unable to bootstrap worker node.")
	}

	if opt.nodeType != dcpconstants.EdgeNode && opt.nodeType != dcpconstants.CloudNode {
		return nil, errors.Errorf("node type(%s) is invalid, only \"edge and cloud\" are supported", opt.nodeType)
	}

	if opt.unsafeSkipCAVerification && len(opt.caCertHashes) != 0 {
		return nil, errors.Errorf("when --discovery-token-ca-cert-hash is specified, --discovery-token-unsafe-skip-ca-verification should be false.")
	} else if len(opt.caCertHashes) == 0 && !opt.unsafeSkipCAVerification {
		return nil, errors.Errorf("when --discovery-token-ca-cert-hash is not specified, --discovery-token-unsafe-skip-ca-verification should be true")
	}

	ignoreErrors := sets.String{}
	for i := range opt.ignorePreflightErrors {
		ignoreErrors.Insert(opt.ignorePreflightErrors[i])
	}

	// Either use the config file if specified, or convert public kubeadm API to the internal JoinConfiguration
	// and validates JoinConfiguration
	name := opt.nodeName
	if name == "" {
		klog.V(1).Infoln("[preflight] found NodeName empty; using OS hostname as NodeName")
		hostname, err := os.Hostname()
		if err != nil {
			return nil, err
		}
		name = hostname
	}

	data := &joinData{
		apiServerEndpoint:     apiServerEndpoint,
		token:                 opt.token,
		tlsBootstrapCfg:       nil,
		ignorePreflightErrors: ignoreErrors,
		pauseImage:            opt.pauseImage,
		engineImage:           opt.engineImage,
		caCertHashes:          sets.NewString(opt.caCertHashes...),
		organizations:         opt.organizations,
		nodeLabels:            make(map[string]string),
		joinNodeData: &joindata.NodeRegistration{
			Name:          name,
			WorkingMode:   opt.nodeType,
			CRISocket:     opt.criSocket,
			Organizations: opt.organizations,
		},
		kubernetesResourceServer: opt.kubernetesResourceServer,
	}

	// parse node labels
	if len(opt.nodeLabels) != 0 {
		parts := strings.Split(opt.nodeLabels, ",")
		for i := range parts {
			kv := strings.Split(parts[i], "=")
			if len(kv) != 2 {
				klog.Warningf("node labels(%s) format is invalid, expect k1=v1,k2=v2", parts[i])
				continue
			}
			data.nodeLabels[kv[0]] = kv[1]
		}
	}

	// get tls bootstrap config
	cfg, err := token.RetrieveBootstrapConfig(data)
	if err != nil {
		klog.Errorf("failed to retrieve bootstrap config, %v", err)
		return nil, err
	}
	data.tlsBootstrapCfg = cfg

	// get kubernetes version
	client, err := kubeconfigutil.ToClientSet(cfg)
	if err != nil {
		klog.Errorf("failed to create bootstrap client, %v", err)
		return nil, err
	}
	data.clientSet = client

	k8sVersion, err := clientutil.GetKubernetesVersionFromCluster(client)
	if err != nil {
		klog.Errorf("failed to get kubernetes version, %v", err)
		return nil, err
	}
	data.kubernetesVersion = k8sVersion
	klog.Infof("node join data info: %#+v", *data)

	return data, nil
}

// ServerAddr returns the public address of kube-apiserver.
func (j *joinData) ServerAddr() string {
	return j.apiServerEndpoint
}

// JoinToken returns bootstrap token for joining node
func (j *joinData) JoinToken() string {
	return j.token
}

// PauseImage returns the pause image.
func (j *joinData) PauseImage() string {
	return j.pauseImage
}

// EngineImage returns the Bhojpur DCP server engine image.
func (j *joinData) EngineImage() string {
	return j.engineImage
}

// KubernetesVersion returns the kubernetes version.
func (j *joinData) KubernetesVersion() string {
	return j.kubernetesVersion
}

// TLSBootstrapCfg returns the cluster-info (kubeconfig).
func (j *joinData) TLSBootstrapCfg() *clientcmdapi.Config {
	return j.tlsBootstrapCfg
}

// BootstrapClient returns the kube clientset.
func (j *joinData) BootstrapClient() *clientset.Clientset {
	return j.clientSet
}

func (j *joinData) NodeRegistration() *joindata.NodeRegistration {
	return j.joinNodeData
}

// IgnorePreflightErrors returns the list of preflight errors to ignore.
func (j *joinData) IgnorePreflightErrors() sets.String {
	return j.ignorePreflightErrors
}

func (j *joinData) CaCertHashes() sets.String {
	return j.caCertHashes
}

func (j *joinData) NodeLabels() map[string]string {
	return j.nodeLabels
}

func (j *joinData) KubernetesResourceServer() string {
	return j.kubernetesResourceServer
}
