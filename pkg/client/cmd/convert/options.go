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
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/spf13/pflag"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	enutil "github.com/bhojpur/dcp/pkg/client/util/edgenode"
	kubeutil "github.com/bhojpur/dcp/pkg/client/util/kubernetes"
	strutil "github.com/bhojpur/dcp/pkg/client/util/strings"
)

// Provider signifies the provider type
type Provider string

const (
	// ProviderMinikube is used if the target kubernetes is run on minikube
	ProviderMinikube Provider = "minikube"
	// ProviderACK is used if the target kubernetes is run on ack
	ProviderACK     Provider = "ack"
	ProviderKubeadm Provider = "kubeadm"
	// ProviderKind is used if the target kubernetes is run on kind
	ProviderKind Provider = "kind"

	Amd64 string = "amd64"
	Arm64 string = "arm64"
	Arm   string = "arm"
)

// ConvertOptions has the information that required by convert operation
type ConvertOptions struct {
	CloudNodes []string
	// AutonomousNodes stores the names of edge nodes that are going to be marked as autonomous.
	// If empty, all edge nodes will be marked as autonomous.
	AutonomousNodes          []string
	TunnelServerAddress      string
	KubeConfigPath           string
	KubeadmConfPath          string
	Provider                 Provider
	EngineHealthCheckTimeout time.Duration
	WaitServantJobTimeout    time.Duration
	IgnorePreflightErrors    sets.String
	DeployTunnel             bool
	EnableAppManager         bool

	SystemArchitecture     string
	EngineImage            string
	ControllerManagerImage string
	NodeServantImage       string
	TunnelServerImage      string
	TunnelAgentImage       string
	AppManagerImage        string

	PodMainfestPath     string
	ClientSet           *kubernetes.Clientset
	AppManagerClientSet dynamic.Interface
}

// NewConvertOptions creates a new ConvertOptions
func NewConvertOptions() *ConvertOptions {
	return &ConvertOptions{
		CloudNodes:            []string{},
		AutonomousNodes:       []string{},
		IgnorePreflightErrors: sets.NewString(),
	}
}

// Complete completes all the required options
func (co *ConvertOptions) Complete(flags *pflag.FlagSet) error {
	cnStr, err := flags.GetString("cloud-nodes")
	if err != nil {
		return err
	}
	if cnStr != "" {
		co.CloudNodes = strings.Split(cnStr, ",")
	}

	anStr, err := flags.GetString("autonomous-nodes")
	if err != nil {
		return err
	}
	if anStr != "" {
		co.AutonomousNodes = strings.Split(anStr, ",")
	}

	ytsa, err := flags.GetString("tunnel-server-address")
	if err != nil {
		return err
	}
	co.TunnelServerAddress = ytsa

	kcp, err := flags.GetString("kubeadm-conf-path")
	if err != nil {
		return err
	}
	co.KubeadmConfPath = kcp

	pStr, err := flags.GetString("provider")
	if err != nil {
		return err
	}
	co.Provider = Provider(pStr)

	engineHealthCheckTimeout, err := flags.GetDuration("dcpsvr-healthcheck-timeout")
	if err != nil {
		return err
	}
	co.EngineHealthCheckTimeout = engineHealthCheckTimeout

	waitServantJobTimeout, err := flags.GetDuration("wait-servant-job-timeout")
	if err != nil {
		return err
	}
	co.WaitServantJobTimeout = waitServantJobTimeout

	ipStr, err := flags.GetString("ignore-preflight-errors")
	if err != nil {
		return err
	}
	if ipStr != "" {
		ipStr = strings.ToLower(ipStr)
		co.IgnorePreflightErrors = sets.NewString(strings.Split(ipStr, ",")...)
	}

	dt, err := flags.GetBool("deploy-tunnel")
	if err != nil {
		return err
	}
	co.DeployTunnel = dt

	eam, err := flags.GetBool("enable-app-manager")
	if err != nil {
		return err
	}
	co.EnableAppManager = eam

	sa, err := flags.GetString("system-architecture")
	if err != nil {
		return err
	}
	co.SystemArchitecture = sa

	yhi, err := flags.GetString("dcpsvr-image")
	if err != nil {
		return err
	}
	co.EngineImage = yhi

	ycmi, err := flags.GetString("controller-manager-image")
	if err != nil {
		return err
	}
	co.ControllerManagerImage = ycmi

	nsi, err := flags.GetString("node-servant-image")
	if err != nil {
		return err
	}
	co.NodeServantImage = nsi

	ytsi, err := flags.GetString("tunnel-server-image")
	if err != nil {
		return err
	}
	co.TunnelServerImage = ytsi

	ytai, err := flags.GetString("tunnel-agent-image")
	if err != nil {
		return err
	}
	co.TunnelAgentImage = ytai

	yami, err := flags.GetString("app-manager-image")
	if err != nil {
		return err
	}
	co.AppManagerImage = yami

	// prepare path of cluster kubeconfig file
	co.KubeConfigPath, err = kubeutil.PrepareKubeConfigPath(flags)
	if err != nil {
		return err
	}

	co.PodMainfestPath = enutil.GetPodManifestPath()

	// parse kubeconfig and generate the clientset
	co.ClientSet, err = kubeutil.GenClientSet(flags)
	if err != nil {
		return err
	}

	// parse kubeconfig and generate the appmanagerclientset
	co.AppManagerClientSet, err = kubeutil.GenDynamicClientSet(flags)
	if err != nil {
		return err
	}

	return nil
}

// Validate makes sure provided values for ConvertOptions are valid
func (co *ConvertOptions) Validate() error {
	if err := ValidateKubeConfig(co.KubeConfigPath); err != nil {
		return err
	}

	if err := kubeutil.ValidateServerVersion(co.ClientSet); err != nil {
		return err
	}

	nodeLst, err := co.ClientSet.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	if err := ValidateCloudNodes(co.CloudNodes, nodeLst); err != nil {
		return err
	}

	edgeNodeNames := getEdgeNodeNames(nodeLst, co.CloudNodes)
	if err := ValidateNodeAutonomy(edgeNodeNames, co.AutonomousNodes); err != nil {
		return err
	}
	// If empty, mark all edge nodes as autonomous
	if len(co.AutonomousNodes) == 0 {
		co.AutonomousNodes = make([]string, len(edgeNodeNames))
		copy(co.AutonomousNodes, edgeNodeNames)
	}

	if err := ValidateTunnelServerAddress(co.TunnelServerAddress); err != nil {
		return err
	}
	if err := ValidateProvider(co.Provider); err != nil {
		return err
	}
	if err := ValidateEngineHealthCheckTimeout(co.EngineHealthCheckTimeout); err != nil {
		return err
	}
	if err := ValidateWaitServantJobTimeout(co.WaitServantJobTimeout); err != nil {
		return err
	}
	if err := ValidateIgnorePreflightErrors(co.IgnorePreflightErrors); err != nil {
		return err
	}
	if err := ValidateSystemArchitecture(co.SystemArchitecture); err != nil {
		return err
	}

	return nil
}

func ValidateCloudNodes(cloudNodeNames []string, nodeLst *v1.NodeList) error {
	if cloudNodeNames == nil || len(cloudNodeNames) == 0 {
		return fmt.Errorf("invalid --cloud-nodes: cannot be empty, please specify the cloud nodes")
	}

	var notExistNodeNames []string
	nodeNameSet := make(map[string]struct{})
	for _, node := range nodeLst.Items {
		nodeNameSet[node.GetName()] = struct{}{}
	}
	for _, name := range cloudNodeNames {
		if _, ok := nodeNameSet[name]; !ok {
			notExistNodeNames = append(notExistNodeNames, name)
		}
	}
	if len(notExistNodeNames) != 0 {
		return fmt.Errorf("invalid --cloud-nodes: the nodes %v are not kubernetes node, can't be converted to cloud node", notExistNodeNames)
	}
	return nil
}

func ValidateNodeAutonomy(edgeNodeNames []string, autonomousNodeNames []string) error {
	var invaildation []string
	for _, name := range autonomousNodeNames {
		if !strutil.IsInStringLst(edgeNodeNames, name) {
			invaildation = append(invaildation, name)
		}
	}
	if len(invaildation) != 0 {
		return fmt.Errorf("invalid --autonomous-nodes: can't make unedge nodes %v autonomous", invaildation)
	}
	return nil
}

func ValidateProvider(provider Provider) error {
	if provider != ProviderMinikube && provider != ProviderACK &&
		provider != ProviderKubeadm && provider != ProviderKind {
		return fmt.Errorf("invalid --provider: %s, valid providers are: minikube, ack, kubeadm, kind",
			provider)
	}
	return nil
}

func ValidateTunnelServerAddress(address string) error {
	if address != "" {
		if _, _, err := net.SplitHostPort(address); err != nil {
			return fmt.Errorf("invalid --tunnel-server-address: %s", err)
		}
	}
	return nil
}

func ValidateSystemArchitecture(arch string) error {
	if arch != Amd64 && arch != Arm64 && arch != Arm {
		return fmt.Errorf("invalid --system-architecture: %s, valid arch are: amd64, arm64, arm", arch)
	}
	return nil
}

func ValidateIgnorePreflightErrors(ignoreErrors sets.String) error {
	if ignoreErrors.Has("all") && ignoreErrors.Len() > 1 {
		return fmt.Errorf("invalid --ignore-preflight-errors: please don't specify individual checks if 'all' is used in option 'ignorePreflightErrors'")
	}
	return nil
}

func ValidateKubeConfig(kbCfgPath string) error {
	if _, err := enutil.FileExists(kbCfgPath); err != nil {
		return fmt.Errorf("invalid kubeconfig path: %v", err)
	}
	return nil

}

func ValidateEngineHealthCheckTimeout(t time.Duration) error {
	if t <= 0 {
		return fmt.Errorf("invalid --dcpsvr-healthcheck-timeout: time must be a valid number(greater than 0)")
	}
	return nil
}

func ValidateWaitServantJobTimeout(t time.Duration) error {
	if t <= 0 {
		return fmt.Errorf("invalid --wait-servant-job-timeout: time must be a valid number(greater than 0)")
	}
	return nil
}
