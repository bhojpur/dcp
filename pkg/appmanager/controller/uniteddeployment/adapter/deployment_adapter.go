package adapter

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

	"k8s.io/klog"

	alpha1 "github.com/bhojpur/dcp/pkg/appmanager/apis/apps/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type DeploymentAdapter struct {
	client.Client

	Scheme *runtime.Scheme
}

var _ Adapter = &DeploymentAdapter{}

// NewResourceObject creates a empty Deployment object.
func (a *DeploymentAdapter) NewResourceObject() runtime.Object {
	return &appsv1.Deployment{}
}

// NewResourceListObject creates a empty DeploymentList object.
func (a *DeploymentAdapter) NewResourceListObject() runtime.Object {
	return &appsv1.DeploymentList{}
}

// GetStatusObservedGeneration returns the observed generation of the pool.
func (a *DeploymentAdapter) GetStatusObservedGeneration(obj metav1.Object) int64 {
	return obj.(*appsv1.Deployment).Status.ObservedGeneration
}

// GetDetails returns the replicas detail the pool needs.
func (a *DeploymentAdapter) GetDetails(obj metav1.Object) (ReplicasInfo, error) {
	set := obj.(*appsv1.Deployment)

	var specReplicas int32
	if set.Spec.Replicas != nil {
		specReplicas = *set.Spec.Replicas
	}
	replicasInfo := ReplicasInfo{
		Replicas:      specReplicas,
		ReadyReplicas: set.Status.ReadyReplicas,
	}
	return replicasInfo, nil
}

// GetPoolFailure returns the failure information of the pool.
// Deployment has no condition.
func (a *DeploymentAdapter) GetPoolFailure() *string {
	return nil
}

// ApplyPoolTemplate updates the pool to the latest revision, depending on the DeploymentTemplate.
func (a *DeploymentAdapter) ApplyPoolTemplate(ud *alpha1.UnitedDeployment, poolName, revision string,
	replicas int32, obj runtime.Object) error {
	set := obj.(*appsv1.Deployment)

	var poolConfig *alpha1.Pool
	for i, pool := range ud.Spec.Topology.Pools {
		if pool.Name == poolName {
			poolConfig = &(ud.Spec.Topology.Pools[i])
			break
		}
	}
	if poolConfig == nil {
		return fmt.Errorf("fail to find pool config %s", poolName)
	}

	set.Namespace = ud.Namespace

	if set.Labels == nil {
		set.Labels = map[string]string{}
	}
	for k, v := range ud.Spec.WorkloadTemplate.DeploymentTemplate.Labels {
		set.Labels[k] = v
	}
	for k, v := range ud.Spec.Selector.MatchLabels {
		set.Labels[k] = v
	}
	set.Labels[alpha1.ControllerRevisionHashLabelKey] = revision
	// record the pool name as a label
	set.Labels[alpha1.PoolNameLabelKey] = poolName

	if set.Annotations == nil {
		set.Annotations = map[string]string{}
	}
	for k, v := range ud.Spec.WorkloadTemplate.DeploymentTemplate.Annotations {
		set.Annotations[k] = v
	}

	set.GenerateName = getPoolPrefix(ud.Name, poolName)

	selectors := ud.Spec.Selector.DeepCopy()
	selectors.MatchLabels[alpha1.PoolNameLabelKey] = poolName

	if err := controllerutil.SetControllerReference(ud, set, a.Scheme); err != nil {
		return err
	}

	set.Spec.Selector = selectors
	set.Spec.Replicas = &replicas

	set.Spec.Strategy = *ud.Spec.WorkloadTemplate.DeploymentTemplate.Spec.Strategy.DeepCopy()
	set.Spec.Template = *ud.Spec.WorkloadTemplate.DeploymentTemplate.Spec.Template.DeepCopy()
	if set.Spec.Template.Labels == nil {
		set.Spec.Template.Labels = map[string]string{}
	}
	set.Spec.Template.Labels[alpha1.PoolNameLabelKey] = poolName
	set.Spec.Template.Labels[alpha1.ControllerRevisionHashLabelKey] = revision

	set.Spec.RevisionHistoryLimit = ud.Spec.RevisionHistoryLimit
	set.Spec.MinReadySeconds = ud.Spec.WorkloadTemplate.DeploymentTemplate.Spec.MinReadySeconds
	set.Spec.Paused = ud.Spec.WorkloadTemplate.DeploymentTemplate.Spec.Paused
	set.Spec.ProgressDeadlineSeconds = ud.Spec.WorkloadTemplate.DeploymentTemplate.Spec.ProgressDeadlineSeconds

	attachNodeAffinityAndTolerations(&set.Spec.Template.Spec, poolConfig)

	if !PoolHasPatch(poolConfig, set) {
		klog.Infof("Deployment[%s/%s-] has no patches, do not need strategicmerge", set.Namespace,
			set.GenerateName)
		return nil
	}

	patched := &appsv1.Deployment{}
	if err := CreateNewPatchedObject(poolConfig.Patch, set, patched); err != nil {
		klog.Errorf("Deployment[%s/%s-] strategic merge by patch %s error %v", set.Namespace,
			set.GenerateName, string(poolConfig.Patch.Raw), err)
		return err
	}
	patched.DeepCopyInto(set)

	klog.Infof("Deployment [%s/%s-] has patches configure successfully:%v", set.Namespace,
		set.GenerateName, string(poolConfig.Patch.Raw))
	return nil
}

// PostUpdate does some works after pool updated. Deployment will implement this method to clean stuck pods.
func (a *DeploymentAdapter) PostUpdate(ud *alpha1.UnitedDeployment, obj runtime.Object, revision string) error {
	// Do nothing,
	return nil
}

// IsExpected checks the pool is the expected revision or not.
// The revision label can tell the current pool revision.
func (a *DeploymentAdapter) IsExpected(obj metav1.Object, revision string) bool {
	return obj.GetLabels()[alpha1.ControllerRevisionHashLabelKey] != revision
}
