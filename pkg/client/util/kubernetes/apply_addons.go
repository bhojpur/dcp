package kubernetes

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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/bhojpur/dcp/pkg/client/constants"
	"github.com/bhojpur/dcp/pkg/client/util/edgenode"
	"github.com/bhojpur/dcp/pkg/projectinfo"
)

func DeployControllerManager(client *kubernetes.Clientset, dcpControllerManagerImage string) error {
	if err := CreateServiceAccountFromYaml(client,
		SystemNamespace, constants.ControllerManagerServiceAccount); err != nil {
		return err
	}
	// create the clusterrole
	if err := CreateClusterRoleFromYaml(client,
		constants.ControllerManagerClusterRole); err != nil {
		return err
	}
	// bind the clusterrole
	if err := CreateClusterRoleBindingFromYaml(client,
		constants.ControllerManagerClusterRoleBinding); err != nil {
		return err
	}
	// create the controller-manager deployment
	if err := CreateDeployFromYaml(client,
		SystemNamespace,
		constants.ControllerManagerDeployment,
		map[string]string{
			"image":         dcpControllerManagerImage,
			"edgeNodeLabel": projectinfo.GetEdgeWorkerLabelKey()}); err != nil {
		return err
	}
	return nil
}

func DeployAppManager(
	client *kubernetes.Clientset,
	dcpappmanagerImage string,
	dcpAppManagerClient dynamic.Interface,
	systemArchitecture string) error {

	// 1.create the AppManagerCustomResourceDefinition
	// 1.1 nodepool
	if err := CreateCRDFromYaml(client, dcpAppManagerClient, "", []byte(constants.AppManagerNodePool)); err != nil {
		return err
	}

	// 1.2 uniteddeployment
	if err := CreateCRDFromYaml(client, dcpAppManagerClient, "", []byte(constants.AppManagerUnitedDeployment)); err != nil {
		return err
	}

	// 2. create the AppManagerRole
	if err := CreateRoleFromYaml(client, SystemNamespace,
		constants.AppManagerRole); err != nil {
		return err
	}

	// 3. create the ClusterRole
	if err := CreateClusterRoleFromYaml(client,
		constants.AppManagerClusterRole); err != nil {
		return err
	}

	// 4. create the RoleBinding
	if err := CreateRoleBindingFromYaml(client, SystemNamespace,
		constants.AppManagerRolebinding); err != nil {
		return err
	}

	// 5. create the ClusterRoleBinding
	if err := CreateClusterRoleBindingFromYaml(client,
		constants.AppManagerClusterRolebinding); err != nil {
		return err
	}

	// 6. create the Secret
	if err := CreateSecretFromYaml(client, SystemNamespace,
		constants.AppManagerSecret); err != nil {
		return err
	}

	// 7. create the Service
	if err := CreateServiceFromYaml(client,
		constants.AppManagerService); err != nil {
		return err
	}

	// 8. create the Deployment
	if err := CreateDeployFromYaml(client,
		SystemNamespace,
		constants.AppManagerDeployment,
		map[string]string{
			"image":           dcpappmanagerImage,
			"arch":            systemArchitecture,
			"edgeWorkerLabel": projectinfo.GetEdgeWorkerLabelKey()}); err != nil {
		return err
	}

	// 9. create the AppManagerMutatingWebhookConfiguration
	if err := CreateMutatingWebhookConfigurationFromYaml(client,
		constants.AppManagerMutatingWebhookConfiguration); err != nil {
		return err
	}

	// 10. create the dcpAppManagerValidatingWebhookConfiguration
	if err := CreateValidatingWebhookConfigurationFromYaml(client,
		constants.AppManagerValidatingWebhookConfiguration); err != nil {
		return err
	}

	return nil
}

func DeployTunnelServer(
	client *kubernetes.Clientset,
	certIP string,
	dcptunnelServerImage string,
	systemArchitecture string) error {
	// 1. create the ClusterRole
	if err := CreateClusterRoleFromYaml(client,
		constants.TunnelServerClusterRole); err != nil {
		return err
	}

	// 2. create the ServiceAccount
	if err := CreateServiceAccountFromYaml(client, SystemNamespace,
		constants.TunnelServerServiceAccount); err != nil {
		return err
	}

	// 3. create the ClusterRoleBinding
	if err := CreateClusterRoleBindingFromYaml(client,
		constants.TunnelServerClusterRolebinding); err != nil {
		return err
	}

	// 4. create the Service
	if err := CreateServiceFromYaml(client,
		constants.TunnelServerService); err != nil {
		return err
	}

	// 5. create the internal Service(type=ClusterIP)
	if err := CreateServiceFromYaml(client,
		constants.TunnelServerInternalService); err != nil {
		return err
	}

	// 6. create the Configmap
	if err := CreateConfigMapFromYaml(client,
		SystemNamespace,
		constants.TunnelServerConfigMap); err != nil {
		return err
	}

	// 7. create the Deployment
	if err := CreateDeployFromYaml(client,
		SystemNamespace,
		constants.TunnelServerDeployment,
		map[string]string{
			"image":           dcptunnelServerImage,
			"arch":            systemArchitecture,
			"certIP":          certIP,
			"edgeWorkerLabel": projectinfo.GetEdgeWorkerLabelKey()}); err != nil {
		return err
	}

	return nil
}

func DeployTunnelAgent(
	client *kubernetes.Clientset,
	tunnelServerAddress string,
	dcptunnelAgentImage string) error {
	// 1. Deploy the tunnel-agent DaemonSet
	if err := CreateDaemonSetFromYaml(client,
		constants.TunnelAgentDaemonSet,
		map[string]string{
			"image":               dcptunnelAgentImage,
			"edgeWorkerLabel":     projectinfo.GetEdgeWorkerLabelKey(),
			"tunnelServerAddress": tunnelServerAddress}); err != nil {
		return err
	}
	return nil
}

// DeployEngineSetting deploy clusterrole, clusterrolebinding for Bhojpur DCP server engine static pod.
func DeployEngineSetting(client *kubernetes.Clientset) error {
	// 1. create the ClusterRole
	if err := CreateClusterRoleFromYaml(client, edgenode.EngineClusterRole); err != nil {
		return err
	}

	// 2. create the ClusterRoleBinding
	if err := CreateClusterRoleBindingFromYaml(client, edgenode.EngineClusterRoleBinding); err != nil {
		return err
	}

	// 3. create the Configmap
	if err := CreateConfigMapFromYaml(client,
		SystemNamespace,
		edgenode.EngineConfigMap); err != nil {
		return err
	}

	return nil
}

// DeleteEngineSetting rm settings for Bhojpur DCP server engine pod
func DeleteEngineSetting(client *kubernetes.Clientset) error {

	// 1. delete the ClusterRoleBinding
	if err := client.RbacV1().ClusterRoleBindings().
		Delete(context.Background(), edgenode.EngineComponentName,
			metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("fail to delete the clusterrolebinding/%s: %s",
			edgenode.EngineComponentName, err)
	}

	// 2. delete the ClusterRole
	if err := client.RbacV1().ClusterRoles().
		Delete(context.Background(), edgenode.EngineComponentName,
			metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("fail to delete the clusterrole/%s: %s",
			edgenode.EngineComponentName, err)
	}

	// 3. remove the ConfigMap
	if err := client.CoreV1().ConfigMaps(edgenode.EngineNamespace).
		Delete(context.Background(), edgenode.EngineCmName,
			metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("fail to delete the configmap/%s: %s",
			edgenode.EngineCmName, err)
	}

	return nil
}
