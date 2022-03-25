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
	"github.com/bhojpur/dcp/pkg/appmanager/constant"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateNginxIngressCommonResource(client client.Client) error {
	// 1. Create Namespace
	if err := CreateNamespaceFromYaml(client, constant.NginxIngressControllerNamespace); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 2. Create ClusterRole
	if err := CreateClusterRoleFromYaml(client, constant.NginxIngressControllerClusterRole); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	if err := CreateClusterRoleFromYaml(client, constant.NginxIngressAdmissionWebhookClusterRole); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 3. Create ClusterRoleBinding
	if err := CreateClusterRoleBindingFromYaml(client,
		constant.NginxIngressControllerClusterRoleBinding); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	if err := CreateClusterRoleBindingFromYaml(client,
		constant.NginxIngressAdmissionWebhookClusterRoleBinding); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 4. Create Role
	if err := CreateRoleFromYaml(client,
		constant.NginxIngressControllerRole); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	if err := CreateRoleFromYaml(client,
		constant.NginxIngressAdmissionWebhookRole); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 5. Create RoleBinding
	if err := CreateRoleBindingFromYaml(client,
		constant.NginxIngressControllerRoleBinding); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	if err := CreateRoleBindingFromYaml(client,
		constant.NginxIngressAdmissionWebhookRoleBinding); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 6. Create ServiceAccount
	if err := CreateServiceAccountFromYaml(client,
		constant.NginxIngressControllerServiceAccount); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	if err := CreateServiceAccountFromYaml(client,
		constant.NginxIngressAdmissionWebhookServiceAccount); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 7. Create Configmap
	if err := CreateConfigMapFromYaml(client,
		constant.NginxIngressControllerConfigMap); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	return nil
}

func DeleteNginxIngressCommonResource(client client.Client) error {
	// 1. Delete Configmap
	if err := DeleteConfigMapFromYaml(client,
		constant.NginxIngressControllerConfigMap); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 2. Delete ServiceAccount
	if err := DeleteServiceAccountFromYaml(client,
		constant.NginxIngressControllerServiceAccount); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	if err := DeleteServiceAccountFromYaml(client,
		constant.NginxIngressAdmissionWebhookServiceAccount); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 3. Delete RoleBinding
	if err := DeleteRoleBindingFromYaml(client,
		constant.NginxIngressControllerRoleBinding); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	if err := DeleteRoleBindingFromYaml(client,
		constant.NginxIngressAdmissionWebhookRoleBinding); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 4. Delete Role
	if err := DeleteRoleFromYaml(client,
		constant.NginxIngressControllerRole); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	if err := DeleteRoleFromYaml(client,
		constant.NginxIngressAdmissionWebhookRole); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 5. Delete ClusterRoleBinding
	if err := DeleteClusterRoleBindingFromYaml(client,
		constant.NginxIngressControllerClusterRoleBinding); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	if err := DeleteClusterRoleBindingFromYaml(client,
		constant.NginxIngressAdmissionWebhookClusterRoleBinding); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 6. Delete ClusterRole
	if err := DeleteClusterRoleFromYaml(client, constant.NginxIngressControllerClusterRole); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	if err := DeleteClusterRoleFromYaml(client, constant.NginxIngressAdmissionWebhookClusterRole); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 7. Delete Namespace
	if err := DeleteNamespaceFromYaml(client, constant.NginxIngressControllerNamespace); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	return nil
}

func CreateNginxIngressSpecificResource(client client.Client, poolname string, replicas int32, ownerRef *metav1.OwnerReference) error {
	// 1. Create Deployment
	if err := CreateDeployFromYaml(client,
		constant.NginxIngressControllerNodePoolDeployment,
		replicas,
		ownerRef,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	if err := CreateDeployFromYaml(client,
		constant.NginxIngressAdmissionWebhookDeployment,
		1,
		nil,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 2. Create Service
	if err := CreateServiceFromYaml(client,
		constant.NginxIngressControllerService,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	if err := CreateServiceFromYaml(client,
		constant.NginxIngressAdmissionWebhookService,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 3. Create ValidatingWebhookConfiguration
	if err := CreateValidatingWebhookConfigurationFromYaml(client,
		constant.NginxIngressValidatingWebhookConfiguration,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 4. Create Job
	if err := CreateJobFromYaml(client,
		constant.NginxIngressAdmissionWebhookJob,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 5. Create Job Patch
	if err := CreateJobFromYaml(client,
		constant.NginxIngressAdmissionWebhookJobPatch,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	return nil
}

func DeleteNginxIngressSpecificResource(client client.Client, poolname string, cleanup bool) error {
	// 1. Delete Deployment
	if err := DeleteDeployFromYaml(client,
		constant.NginxIngressControllerNodePoolDeployment,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	if err := DeleteDeployFromYaml(client,
		constant.NginxIngressAdmissionWebhookDeployment,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 2. Delete Service
	if err := DeleteServiceFromYaml(client,
		constant.NginxIngressControllerService,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	if err := DeleteServiceFromYaml(client,
		constant.NginxIngressAdmissionWebhookService,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 3. Delete ValidatingWebhookConfiguration
	if err := DeleteValidatingWebhookConfigurationFromYaml(client,
		constant.NginxIngressValidatingWebhookConfiguration,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 4. Delete Job
	if err := DeleteJobFromYaml(client,
		constant.NginxIngressAdmissionWebhookJob,
		cleanup,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 5. Delete Job Patch
	if err := DeleteJobFromYaml(client,
		constant.NginxIngressAdmissionWebhookJobPatch,
		cleanup,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	return nil
}

func ScaleNginxIngressControllerDeploymment(client client.Client, poolname string, replicas int32) error {
	if err := UpdateDeployFromYaml(client,
		constant.NginxIngressControllerNodePoolDeployment,
		&replicas,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	return nil
}
