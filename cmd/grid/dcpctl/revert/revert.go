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
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/pkg/client/constants"
	"github.com/bhojpur/dcp/pkg/client/lock"
	enutil "github.com/bhojpur/dcp/pkg/client/util/edgenode"
	kubeutil "github.com/bhojpur/dcp/pkg/client/util/kubernetes"
	nodeutil "github.com/bhojpur/dcp/pkg/controller/util/node"
	nodeservant "github.com/bhojpur/dcp/pkg/node-servant"
	"github.com/bhojpur/dcp/pkg/projectinfo"
	tunneldns "github.com/bhojpur/dcp/pkg/tunnel/trafficforward/dns"
)

// RevertOptions has the information required by the revert operation
type RevertOptions struct {
	clientSet             *kubernetes.Clientset
	waitServantJobTimeout time.Duration
	NodeServantImage      string
	PodMainfestPath       string
	KubeadmConfPath       string
	AppManagerClientSet   dynamic.Interface
}

// NewRevertOptions creates a new RevertOptions
func NewRevertOptions() *RevertOptions {
	return &RevertOptions{}
}

// NewRevertCmd generates a new revert command
func NewRevertCmd() *cobra.Command {
	ro := NewRevertOptions()
	cmd := &cobra.Command{
		Use:   "revert",
		Short: "Reverts the Bhojpur DCP cluster back to a Kubernetes cluster",
		Run: func(cmd *cobra.Command, _ []string) {
			if err := ro.Complete(cmd.Flags()); err != nil {
				klog.Errorf("fail to complete the revert option: %s", err)
				os.Exit(1)
			}
			if err := ro.RunRevert(); err != nil {
				klog.Errorf("fail to revert Bhojpur DCP to kubernetes: %s", err)
				os.Exit(1)
			}
		},
		Args: cobra.NoArgs,
	}

	cmd.Flags().String("node-servant-image",
		"bhojpur/node-servant:latest",
		"The node-servant image.")
	cmd.Flags().String("kubeadm-conf-path",
		"/etc/systemd/system/kubelet.service.d/10-kubeadm.conf",
		"The path to kubelet service conf that is used by kubelet component to join the cluster on the edge node.")
	cmd.Flags().Duration("wait-servant-job-timeout", kubeutil.DefaultWaitServantJobTimeout,
		"The timeout for servant-job run check.")
	return cmd
}

// Complete completes all the required options
func (ro *RevertOptions) Complete(flags *pflag.FlagSet) error {
	nsi, err := flags.GetString("node-servant-image")
	if err != nil {
		return err
	}
	ro.NodeServantImage = nsi

	kcp, err := flags.GetString("kubeadm-conf-path")
	if err != nil {
		return err
	}
	ro.KubeadmConfPath = kcp

	ro.PodMainfestPath = enutil.GetPodManifestPath()

	waitServantJobTimeout, err := flags.GetDuration("wait-servant-job-timeout")
	if err != nil {
		return err
	}
	ro.waitServantJobTimeout = waitServantJobTimeout

	ro.clientSet, err = kubeutil.GenClientSet(flags)
	if err != nil {
		return err
	}

	ro.AppManagerClientSet, err = kubeutil.GenDynamicClientSet(flags)
	if err != nil {
		return err
	}
	return nil
}

// RunRevert reverts the target Bhojpur DCP cluster back to a standard Kubernetes cluster
func (ro *RevertOptions) RunRevert() (err error) {
	if err = lock.AcquireLock(ro.clientSet); err != nil {
		return
	}
	defer func() {
		if deleteLockErr := lock.DeleteLock(ro.clientSet); deleteLockErr != nil {
			klog.Error(deleteLockErr)
		}
	}()
	klog.V(4).Info("successfully acquire the lock")

	// 1. check the server version
	if err = kubeutil.ValidateServerVersion(ro.clientSet); err != nil {
		return
	}
	klog.V(4).Info("the server version is valid")

	// 1.1. get kube-controller-manager HA nodes
	kcmNodeNames, err := kubeutil.GetKubeControllerManagerHANodes(ro.clientSet)
	if err != nil {
		return
	}

	// 1.2. check the state of all nodes
	nodeLst, err := ro.clientSet.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return
	}

	var nodeNames []string
	for _, node := range nodeLst.Items {
		if !isNodeReady(&node.Status) {
			return fmt.Errorf("cannot do the revert, the status of worker or kube-controller-manager node: %s is not 'Ready'", node.Name)
		}
		nodeNames = append(nodeNames, node.GetName())
	}
	klog.V(4).Info("the status of all nodes are satisfied")

	// 2. remove the Bhojpur DCP controller manager
	if err = ro.clientSet.AppsV1().Deployments(constants.AppManagerNamespace).
		Delete(context.Background(), constants.ControllerManager, metav1.DeleteOptions{
			PropagationPolicy: &kubeutil.PropagationPolicy,
		}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("fail to remove Bhojpur DCP controller manager: %s", err)
	}
	klog.Info("Bhojpur DCP controller manager is removed")

	// 2.1 remove the serviceaccount for controller-manager
	if err = ro.clientSet.CoreV1().ServiceAccounts(constants.AppManagerNamespace).
		Delete(context.Background(), constants.ControllerManager, metav1.DeleteOptions{
			PropagationPolicy: &kubeutil.PropagationPolicy,
		}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("fail to remove serviceaccount for Bhojpur DCP controller manager: %s", err)
	}
	klog.Info("serviceaccount for Bhojpur DCP controller manager is removed")

	// 2.2 remove the clusterrole for controller-manager
	if err = ro.clientSet.RbacV1().ClusterRoles().
		Delete(context.Background(), constants.ControllerManager, metav1.DeleteOptions{
			PropagationPolicy: &kubeutil.PropagationPolicy,
		}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("fail to remove clusterrole for Bhojpur DCP controller manager: %s", err)
	}
	klog.Info("clusterrole for Bhojpur DCP controller manager is removed")

	// 2.3 remove the clusterrolebinding for controller-manager
	if err = ro.clientSet.RbacV1().ClusterRoleBindings().
		Delete(context.Background(), constants.ControllerManager, metav1.DeleteOptions{
			PropagationPolicy: &kubeutil.PropagationPolicy,
		}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("fail to remove clusterrolebinding for Bhojpur DCP controller manager: %s", err)
	}
	klog.Info("clusterrolebinding for Bhojpur DCP controller manager is removed")

	// 3. remove the tunnel agent
	if err = removeTunnelAgent(ro.clientSet); err != nil {
		return fmt.Errorf("fail to remove the Bhojpur DCP tunnel agent: %s", err)
	}

	// 4. remove the tunnel server
	if err = removeTunnelServer(ro.clientSet); err != nil {
		return fmt.Errorf("fail to remove the Bhojpur DCP tunnel server: %s", err)
	}

	// 5. remove the app manager
	if err = removeAppManager(ro.clientSet, ro.AppManagerClientSet); err != nil {
		return fmt.Errorf("fail to remove the Bhojpur DCP app manager: %s", err)
	}
	klog.Info("Bhojpur DCP app manager is removed")

	// 6. enable node-controller
	if err = kubeutil.RunServantJobs(ro.clientSet, ro.waitServantJobTimeout,
		func(nodeName string) (*batchv1.Job, error) {
			ctx := map[string]string{
				"node_servant_image": ro.NodeServantImage,
				"pod_manifest_path":  ro.PodMainfestPath,
			}
			return kubeutil.RenderServantJob("enable", ctx, nodeName)
		},
		kcmNodeNames, os.Stderr); err != nil {
		return fmt.Errorf("fail to run EnableNodeControllerJobs: %s", err)
	}
	klog.Info("complete enabling node-controller")

	// 7. remove dcpsvr and revert kubelet service on edge nodes
	if err = kubeutil.RunServantJobs(ro.clientSet, ro.waitServantJobTimeout, func(nodeName string) (*batchv1.Job, error) {
		ctx := map[string]string{
			"node_servant_image": ro.NodeServantImage,
			"kubeadm_conf_path":  ro.KubeadmConfPath,
		}
		return nodeservant.RenderNodeServantJob("revert", ctx, nodeName)
	}, nodeNames, os.Stderr); err != nil {
		klog.Errorf("fail to revert node: %s", err)
		return
	}
	klog.Info("complete removing dcpsvr and resetting kubelet service")

	// 8. remove server engine k8s config, roleBinding role
	err = kubeutil.DeleteEngineSetting(ro.clientSet)
	if err != nil {
		return fmt.Errorf("fail to delete server engine setting: %s", err)
	}
	klog.Info("delete server engine clusterrole and clusterrolebinding")

	// 9. remove label and annotation of nodes
	nodeLst, err = ro.clientSet.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, node := range nodeLst.Items {
		_, foundAutonomy := node.Annotations[constants.AnnotationAutonomy]
		if foundAutonomy {
			delete(node.Annotations, constants.AnnotationAutonomy)
		}
		delete(node.Labels, projectinfo.GetEdgeWorkerLabelKey())
		if _, err = ro.clientSet.CoreV1().Nodes().Update(context.Background(), &node, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("fail to remove label or annotation for node: %s: %s", node.GetName(), err)
		}
		klog.Infof("label %s is removed from node %s", projectinfo.GetEdgeWorkerLabelKey(), node.GetName())
		if foundAutonomy {
			klog.Infof("annotation %s is removed from node %s", constants.AnnotationAutonomy, node.GetName())
		}
	}

	return
}

func removeTunnelServer(client *kubernetes.Clientset) error {
	// 1. remove the DaemonSet
	if err := client.AppsV1().
		Deployments(constants.TunnelNamespace).
		Delete(context.Background(), constants.TunnelServerComponentName,
			metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("fail to delete the daemonset/%s: %s",
			constants.TunnelServerComponentName, err)
	}
	klog.V(4).Infof("deployment/%s is deleted", constants.TunnelServerComponentName)

	// 2.1 remove the Service
	if err := client.CoreV1().Services(constants.TunnelNamespace).
		Delete(context.Background(), constants.TunnelServerSvcName,
			metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("fail to delete the service/%s: %s",
			constants.TunnelServerSvcName, err)
	}
	klog.V(4).Infof("service/%s is deleted", constants.TunnelServerSvcName)

	// 2.2 remove the internal Service(type=ClusterIP)
	if err := client.CoreV1().Services(constants.TunnelNamespace).
		Delete(context.Background(), constants.TunnelServerInternalSvcName,
			metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("fail to delete the service/%s: %s",
			constants.TunnelServerInternalSvcName, err)
	}
	klog.V(4).Infof("service/%s is deleted", constants.TunnelServerInternalSvcName)

	// 3. remove the ClusterRoleBinding
	if err := client.RbacV1().ClusterRoleBindings().
		Delete(context.Background(), constants.TunnelServerComponentName,
			metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("fail to delete the clusterrolebinding/%s: %s",
			constants.TunnelServerComponentName, err)
	}
	klog.V(4).Infof("clusterrolebinding/%s is deleted", constants.TunnelServerComponentName)

	// 4. remove the SerivceAccount
	if err := client.CoreV1().ServiceAccounts(constants.TunnelNamespace).
		Delete(context.Background(), constants.TunnelServerComponentName,
			metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("fail to delete the serviceaccount/%s: %s",
			constants.TunnelServerComponentName, err)
	}
	klog.V(4).Infof("serviceaccount/%s is deleted", constants.TunnelServerComponentName)

	// 5. remove the ClusterRole
	if err := client.RbacV1().ClusterRoles().
		Delete(context.Background(), constants.TunnelServerComponentName,
			metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("fail to delete the clusterrole/%s: %s",
			constants.TunnelServerComponentName, err)
	}
	klog.V(4).Infof("clusterrole/%s is deleted", constants.TunnelServerComponentName)

	// 6. remove the tunnel-server-cfg
	if err := client.CoreV1().ConfigMaps(constants.TunnelNamespace).
		Delete(context.Background(), constants.TunnelServerCmName,
			metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("fail to delete the configmap/%s: %s",
			constants.TunnelServerCmName, err)
	}

	// 7. remove the dns record configmap
	dcptunnelDnsRecordConfigMapName := tunneldns.GetTunnelDNSRecordConfigMapName()
	if err := client.CoreV1().ConfigMaps(constants.TunnelNamespace).
		Delete(context.Background(), dcptunnelDnsRecordConfigMapName,
			metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("fail to delete configmap/%s: %s",
			dcptunnelDnsRecordConfigMapName, err)
	}

	return nil
}

func removeAppManager(client *kubernetes.Clientset, dcpAppManagerClientSet dynamic.Interface) error {
	// 1. remove the Deployment
	if err := client.AppsV1().
		Deployments(constants.AppManagerNamespace).
		Delete(context.Background(), constants.AppManager,
			metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("fail to delete the deployment/%s: %s",
			constants.AppManager, err)
	}
	klog.Info("deployment for Bhojpur DCP app manager is removed")
	klog.V(4).Infof("deployment/%s is deleted", constants.AppManager)

	// 2. remove the Role
	if err := client.RbacV1().Roles(constants.AppManagerNamespace).
		Delete(context.Background(), "app-leader-election-role",
			metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("fail to delete the role/%s: %s",
			"app-leader-election-role", err)
	}
	klog.Info("Role for Bhojpur DCP app manager is removed")

	// 3. remove the ClusterRole
	if err := client.RbacV1().ClusterRoles().
		Delete(context.Background(), "app-manager-role",
			metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("fail to delete the clusterrole/%s: %s",
			"app-manager-role", err)
	}
	klog.Info("ClusterRole for Bhojpur DCP app manager is removed")

	// 4. remove the ClusterRoleBinding
	if err := client.RbacV1().ClusterRoleBindings().
		Delete(context.Background(), "app-manager-rolebinding",
			metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("fail to delete the clusterrolebinding/%s: %s",
			"app-manager-rolebinding", err)
	}
	klog.Info("ClusterRoleBinding for Bhojpur DCP app manager is removed")
	klog.V(4).Infof("clusterrolebinding/%s is deleted", "app-manager-rolebinding")

	// 5. remove the RoleBinding
	if err := client.RbacV1().RoleBindings(constants.AppManagerNamespace).
		Delete(context.Background(), "app-leader-election-rolebinding",
			metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("fail to delete the rolebinding/%s: %s",
			"app-leader-election-rolebinding", err)
	}
	klog.Info("RoleBinding for Bhojpur DCP app manager is removed")
	klog.V(4).Infof("clusterrolebinding/%s is deleted", "app-leader-election-rolebinding")

	// 6 remove the Secret
	if err := client.CoreV1().Secrets(constants.AppManagerNamespace).
		Delete(context.Background(), "app-webhook-certs",
			metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("fail to delete the secret/%s: %s",
			"app-webhook-certs", err)
	}
	klog.Info("secret for Bhojpur DCP app manager is removed")
	klog.V(4).Infof("secret/%s is deleted", "app-webhook-certs")

	// 7 remove Service
	if err := client.CoreV1().Services(constants.AppManagerNamespace).
		Delete(context.Background(), "app-webhook-service",
			metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("fail to delete the service/%s: %s",
			"app-webhook-service", err)
	}
	klog.Info("Service for Bhojpur DCP app manager is removed")
	klog.V(4).Infof("service/%s is deleted", "app-webhook-service")

	// 8. remove the MutatingWebhookConfiguration
	if err := client.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().
		Delete(context.Background(), "app-mutating-webhook-configuration",
			metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("fail to delete the MutatingWebhookConfiguration/%s: %s",
			"app-mutating-webhook-configuration", err)
	}
	klog.Info("MutatingWebhookConfiguration for Bhojpur DCP app manager is removed")
	klog.V(4).Infof("MutatingWebhookConfiguration/%s is deleted", "app-mutating-webhook-configuration")

	// 9. remove the ValidatingWebhookConfiguration
	if err := client.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().
		Delete(context.Background(), "app-validating-webhook-configuration",
			metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("fail to delete the ValidatingWebhookConfiguration/%s: %s",
			"app-validating-webhook-configuration", err)
	}
	klog.Info("ValidatingWebhookConfiguration for Bhojpur DCP app manager is removed")
	klog.V(4).Infof("ValidatingWebhookConfiguration/%s is deleted", "app-validating-webhook-configuration")

	// 10. remove nodepoolcrd
	if err := kubeutil.DeleteCRDResource(client, dcpAppManagerClientSet,
		"NodePool", "nodepools.apps.bhojpur.net", []byte(constants.AppManagerNodePool)); err != nil {
		return fmt.Errorf("fail to delete the NodePoolCRD/%s: %s",
			"nodepoolcrd", err)
	}
	klog.Info("crd for Bhojpur DCP app manager is removed")
	klog.V(4).Infof("NodePoolCRD/%s is deleted", "NodePool")

	// 11. remove UnitedDeploymentcrd
	if err := kubeutil.DeleteCRDResource(client, dcpAppManagerClientSet,
		"UnitedDeployment", "uniteddeployments.apps.bhojpur.net", []byte(constants.AppManagerUnitedDeployment)); err != nil {
		return fmt.Errorf("fail to delete the UnitedDeploymentCRD/%s: %s",
			"UnitedDeployment", err)
	}
	klog.Info("UnitedDeploymentcrd for Bhojpur DCP app manager is removed")
	klog.V(4).Infof("UnitedDeploymentCRD/%s is deleted", "UnitedDeployment")
	return nil
}

func removeTunnelAgent(client *kubernetes.Clientset) error {
	// 1. remove the DaemonSet
	if err := client.AppsV1().
		DaemonSets(constants.TunnelNamespace).
		Delete(context.Background(), constants.TunnelAgentComponentName,
			metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("fail to delete the daemonset/%s: %s",
			constants.TunnelAgentComponentName, err)
	}
	klog.V(4).Infof("daemonset/%s is deleted", constants.TunnelAgentComponentName)

	// 2. remove the ClusterRoleBinding
	if err := client.RbacV1().ClusterRoleBindings().
		Delete(context.Background(), constants.TunnelAgentComponentName,
			metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("fail to delete the clusterrolebinding/%s: %s",
			constants.TunnelAgentComponentName, err)
	}
	klog.V(4).Infof("clusterrolebinding/%s is deleted", constants.TunnelAgentComponentName)

	// 3. remove the ClusterRole
	if err := client.RbacV1().ClusterRoles().
		Delete(context.Background(), constants.TunnelAgentComponentName,
			metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("fail to delete the clusterrole/%s: %s",
			constants.TunnelAgentComponentName, err)
	}
	klog.V(4).Infof("clusterrole/%s is deleted", constants.TunnelAgentComponentName)
	return nil
}

func isNodeReady(status *v1.NodeStatus) bool {
	_, condition := nodeutil.GetNodeCondition(status, v1.NodeReady)
	return condition != nil && condition.Status == v1.ConditionTrue
}
