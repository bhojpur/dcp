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
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	bootstrapapi "k8s.io/cluster-bootstrap/token/api"
	"k8s.io/klog/v2"

	kubeadmapi "github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/phases/bootstraptoken/clusterinfo"
	"github.com/bhojpur/dcp/pkg/client/lock"
	kubeutil "github.com/bhojpur/dcp/pkg/client/util/kubernetes"
	strutil "github.com/bhojpur/dcp/pkg/client/util/strings"
	"github.com/bhojpur/dcp/pkg/engine/util"
	nodeservant "github.com/bhojpur/dcp/pkg/node-servant"
	"github.com/bhojpur/dcp/pkg/preflight"
)

const (
	// defaultEngineHealthCheckTimeout defines the default timeout for Bhojpur DCP engine health check phase
	defaultEngineHealthCheckTimeout = 2 * time.Minute

	latestEngineImage            = "bhojpur/dcpsvr:latest"
	latestControllerManagerImage = "bhojpur/controller-manager:latest"
	latestNodeServantImage       = "bhojpur/node-servant:latest"
	latestTunnelServerImage      = "bhojpur/tunnel-server:latest"
	latestTunnelAgentImage       = "bhojpur/tunnel-agent:latest"
	versionedAppManagerImage     = "bhojpur/app-manager:v0.4.0"
)

// ClusterConverter do the cluster convert job.
// During the conversion, the pre-check will be performed first, and
// the conversion will be performed only after the pre-check is passed.
type ClusterConverter struct {
	ConvertOptions
}

// NewConvertCmd generates a new convert command
func NewConvertCmd() *cobra.Command {
	co := NewConvertOptions()
	cmd := &cobra.Command{
		Use:   "convert -c CLOUDNODES",
		Short: "Converts the kubernetes cluster to a Bhojpur DCP cluster",
		Run: func(cmd *cobra.Command, _ []string) {
			if err := co.Complete(cmd.Flags()); err != nil {
				klog.Errorf("Fail to complete the convert option: %s", err)
				os.Exit(1)
			}
			if err := co.Validate(); err != nil {
				klog.Errorf("Fail to validate convert option: %s", err)
				os.Exit(1)
			}
			converter := NewClusterConverter(co)
			if err := converter.PreflightCheck(); err != nil {
				klog.Errorf("Fail to run pre-flight checks: %s", err)
				os.Exit(1)
			}
			if err := converter.RunConvert(); err != nil {
				klog.Errorf("Fail to convert the cluster: %s", err)
				os.Exit(1)
			}
		},
		Args: cobra.NoArgs,
	}

	setFlags(cmd)
	return cmd
}

// setFlags sets flags.
func setFlags(cmd *cobra.Command) {
	cmd.Flags().StringP(
		"cloud-nodes", "c", "",
		"The list of cloud nodes.(e.g. -c cloudnode1,cloudnode2)",
	)
	cmd.Flags().StringP(
		"autonomous-nodes", "a", "",
		"The list of nodes that will be marked as autonomous. If not set, all edge nodes will be marked as autonomous."+
			"(e.g. -a autonomousnode1,autonomousnode2)",
	)
	cmd.Flags().String(
		"tunnel-server-address", "",
		"The tunnel-server address.",
	)
	cmd.Flags().String(
		"kubeadm-conf-path", "",
		"The path to kubelet service conf that is used by kubelet component to join the cluster on the edge node.",
	)
	cmd.Flags().StringP(
		"provider", "p", "minikube",
		"The provider of the original Kubernetes cluster.",
	)
	cmd.Flags().Duration(
		"dcpsvr-healthcheck-timeout", defaultEngineHealthCheckTimeout,
		"The timeout for Bhojpur DCP engine health check.",
	)
	cmd.Flags().Duration(
		"wait-servant-job-timeout", kubeutil.DefaultWaitServantJobTimeout,
		"The timeout for servant-job run check.")
	cmd.Flags().String(
		"ignore-preflight-errors", "",
		"A list of checks whose errors will be shown as warnings. Example: 'NodeEdgeWorkerLabel,NodeAutonomy'.Value 'all' ignores errors from all checks.",
	)
	cmd.Flags().BoolP(
		"deploy-tunnel", "t", false,
		"If set, tunnel will be deployed.",
	)
	cmd.Flags().BoolP(
		"enable-app-manager", "e", false,
		"If set, appmanager will be deployed.",
	)
	cmd.Flags().String(
		"system-architecture", "amd64",
		"The system architecture of cloud nodes.",
	)

	cmd.Flags().String("dcpsvr-image", latestEngineImage, "The Bhojpur DCP server engine image.")
	cmd.Flags().String("controller-manager-image", latestControllerManagerImage, "The controller-manager image.")
	cmd.Flags().String("node-servant-image", latestNodeServantImage, "The node-servant image.")
	cmd.Flags().String("tunnel-server-image", latestTunnelServerImage, "The tunnel-server image.")
	cmd.Flags().String("tunnel-agent-image", latestTunnelAgentImage, "The tunnel-agent image.")
	cmd.Flags().String("app-manager-image", versionedAppManagerImage, "The app-manager image.")
}

func NewClusterConverter(co *ConvertOptions) *ClusterConverter {
	return &ClusterConverter{
		*co,
	}
}

// preflightCheck executes preflight checks logic.
func (c *ClusterConverter) PreflightCheck() error {
	fmt.Println("[preflight] Running pre-flight checks")
	if err := preflight.RunConvertClusterChecks(c.ClientSet, c.IgnorePreflightErrors); err != nil {
		return err
	}

	fmt.Println("[preflight] Running node-servant-preflight-convert jobs to check on edge and cloud nodes. " +
		"Job running may take a long time, and job failure will affect the execution of the next stage")
	jobLst, err := c.generatePreflightJobs()
	if err != nil {
		return err
	}
	if err := preflight.RunNodeServantJobCheck(c.ClientSet, jobLst, c.WaitServantJobTimeout, kubeutil.CheckServantJobPeriod, c.IgnorePreflightErrors); err != nil {
		return err
	}

	return nil
}

// RunConvert performs the conversion
func (c *ClusterConverter) RunConvert() (err error) {
	if err = lock.AcquireLock(c.ClientSet); err != nil {
		return
	}
	defer func() {
		if releaseLockErr := lock.ReleaseLock(c.ClientSet); releaseLockErr != nil {
			klog.Error(releaseLockErr)
		}
	}()

	nodeLst, err := c.ClientSet.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return
	}
	edgeNodeNames := getEdgeNodeNames(nodeLst, c.CloudNodes)

	fmt.Println("[runConvert] Label all nodes with edgeworker label, annotate all nodes with autonomy annotation")
	for _, node := range nodeLst.Items {
		isEdge := strutil.IsInStringLst(edgeNodeNames, node.Name)
		isAuto := strutil.IsInStringLst(c.AutonomousNodes, node.Name)
		if _, err = kubeutil.AddEdgeWorkerLableAndAutonomyAnnotation(c.ClientSet, &node,
			strconv.FormatBool(isEdge), strconv.FormatBool(isAuto)); err != nil {
			return
		}
	}

	// deploy controller manager
	fmt.Println("[runConvert] Deploying controller-manager")
	if err = kubeutil.DeployControllerManager(c.ClientSet, c.ControllerManagerImage); err != nil {
		return
	}

	// deploy the tunnel if required
	if c.DeployTunnel {
		fmt.Println("[runConvert] Deploying tunnel-server and tunnel-agent")
		var certIP string
		if c.TunnelServerAddress != "" {
			certIP, _, _ = net.SplitHostPort(c.TunnelServerAddress)
		}
		if err = kubeutil.DeployTunnelServer(c.ClientSet,
			certIP,
			c.TunnelServerImage,
			c.SystemArchitecture); err != nil {
			return
		}
		// we will deploy tunnel-agent on every edge node
		if err = kubeutil.DeployTunnelAgent(c.ClientSet,
			c.TunnelServerAddress,
			c.TunnelAgentImage); err != nil {
			return
		}
	}

	// deploy the appmanager if required
	if c.EnableAppManager {
		fmt.Println("[runConvert] Deploying app-manager")
		if err = kubeutil.DeployAppManager(c.ClientSet,
			c.AppManagerImage,
			c.AppManagerClientSet,
			c.SystemArchitecture); err != nil {
			return
		}
	}

	fmt.Println("[runConvert] Running jobs for convert. Job running may take a long time, and job failure will not affect the execution of the next stage")

	// disable node-controller
	fmt.Println("[runConvert] Running disable-node-controller jobs to disable node-controller")
	var kcmNodeNames []string
	if kcmNodeNames, err = kubeutil.GetKubeControllerManagerHANodes(c.ClientSet); err != nil {
		return
	}

	if err = kubeutil.RunServantJobs(c.ClientSet, c.WaitServantJobTimeout, func(nodeName string) (*batchv1.Job, error) {
		ctx := map[string]string{
			"node_servant_image": c.NodeServantImage,
			"pod_manifest_path":  c.PodMainfestPath,
		}
		return kubeutil.RenderServantJob("disable", ctx, nodeName)
	}, kcmNodeNames, os.Stderr); err != nil {
		return
	}

	// deploy dcpsvr and reset the kubelet service on edge nodes.
	fmt.Println("[runConvert] Running node-servant-convert jobs to deploy the dcpsvr and reset the kubelet service on edge and cloud nodes")
	var joinToken string
	if joinToken, err = prepareEngineStart(c.ClientSet, c.KubeConfigPath); err != nil {
		return
	}

	convertCtx := map[string]string{
		"node_servant_image": c.NodeServantImage,
		"dcpsvr_image":       c.EngineImage,
		"joinToken":          joinToken,
		"kubeadm_conf_path":  c.KubeadmConfPath,
		"working_mode":       string(util.WorkingModeEdge),
	}
	if c.EngineHealthCheckTimeout != defaultEngineHealthCheckTimeout {
		convertCtx["dcpsvr_healthcheck_timeout"] = c.EngineHealthCheckTimeout.String()
	}
	if len(edgeNodeNames) != 0 {
		convertCtx["working_mode"] = string(util.WorkingModeEdge)
		if err = kubeutil.RunServantJobs(c.ClientSet, c.WaitServantJobTimeout, func(nodeName string) (*batchv1.Job, error) {
			return nodeservant.RenderNodeServantJob("convert", convertCtx, nodeName)
		}, edgeNodeNames, os.Stderr); err != nil {
			return
		}
	}

	// deploy dcpsvr and reset the kubelet service on cloud nodes
	convertCtx["working_mode"] = string(util.WorkingModeCloud)
	if err = kubeutil.RunServantJobs(c.ClientSet, c.WaitServantJobTimeout, func(nodeName string) (*batchv1.Job, error) {
		return nodeservant.RenderNodeServantJob("convert", convertCtx, nodeName)
	}, c.CloudNodes, os.Stderr); err != nil {
		return
	}

	fmt.Println("[runConvert] If any job fails, you can get job information through 'kubectl get jobs -n kube-system' to debug.\n" +
		"\tNote that before the next conversion, please delete all related jobs so as not to affect the conversion.")

	return

}

func prepareEngineStart(cliSet *kubernetes.Clientset, kcfg string) (string, error) {
	// prepare kube-public/cluster-info configmap before convert
	if err := prepareClusterInfoConfigMap(cliSet, kcfg); err != nil {
		return "", err
	}

	// prepare global settings(like RBAC, configmap) for Bhojpur DCP server engine
	if err := kubeutil.DeployEngineSetting(cliSet); err != nil {
		return "", err
	}

	// prepare join-token for Bhojpur DCP server engine
	joinToken, err := kubeutil.GetOrCreateJoinTokenString(cliSet)
	if err != nil || joinToken == "" {
		return "", fmt.Errorf("fail to get join token: %v", err)
	}
	return joinToken, nil

}

// generatePreflightJobs generate preflight-check job for each node
func (c *ClusterConverter) generatePreflightJobs() ([]*batchv1.Job, error) {
	jobLst := make([]*batchv1.Job, 0)
	preflightCtx := map[string]string{
		"node_servant_image":      c.NodeServantImage,
		"kubeadm_conf_path":       c.KubeadmConfPath,
		"ignore_preflight_errors": strings.Join(c.IgnorePreflightErrors.List(), ","),
	}

	nodeLst, err := c.ClientSet.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, node := range nodeLst.Items {
		job, err := nodeservant.RenderNodeServantJob("preflight-convert", preflightCtx, node.Name)
		if err != nil {
			return nil, fmt.Errorf("fail to get job for node %s: %s", node.Name, err)
		}
		jobLst = append(jobLst, job)
	}

	return jobLst, nil
}

// prepareClusterInfoConfigMap will create cluster-info configmap in kube-public namespace if it does not exist
func prepareClusterInfoConfigMap(client *kubernetes.Clientset, file string) error {
	info, err := client.CoreV1().ConfigMaps(metav1.NamespacePublic).Get(context.Background(), bootstrapapi.ConfigMapClusterInfo, metav1.GetOptions{})
	if err != nil && apierrors.IsNotFound(err) {
		// Create the cluster-info ConfigMap with the associated RBAC rules
		if err := kubeadmapi.CreateBootstrapConfigMapIfNotExists(client, file); err != nil {
			return fmt.Errorf("error creating bootstrap ConfigMap, %v", err)
		}
		if err := kubeadmapi.CreateClusterInfoRBACRules(client); err != nil {
			return fmt.Errorf("error creating clusterinfo RBAC rules, %v", err)
		}
	} else if err != nil || info == nil {
		return fmt.Errorf("fail to get configmap, %v", err)
	} else {
		klog.V(4).Infof("%s/%s configmap already exists, skip to prepare it", info.Namespace, info.Name)
	}

	return nil
}

func getEdgeNodeNames(nodeLst *v1.NodeList, cloudNodeNames []string) (edgeNodeNames []string) {
	for _, node := range nodeLst.Items {
		if !strutil.IsInStringLst(cloudNodeNames, node.GetName()) {
			edgeNodeNames = append(edgeNodeNames, node.GetName())
		}
	}
	return
}
