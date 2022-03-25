package workloadcontroller

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
	"errors"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/bhojpur/dcp/pkg/appmanager/apis/apps/v1alpha1"
	"github.com/bhojpur/dcp/pkg/appmanager/util/refmanager"
)

const updateRetries = 5

type DeploymentControllor struct {
	client.Client
	Scheme *runtime.Scheme
}

func (d *DeploymentControllor) GetTemplateType() v1alpha1.TemplateType {
	return v1alpha1.DeploymentTemplateType
}

func (d *DeploymentControllor) DeleteWorkload(yda *v1alpha1.DcpAppDaemon, load *Workload) error {
	klog.Infof("DcpAppDaemon[%s/%s] prepare delete Deployment[%s/%s]", yda.GetNamespace(),
		yda.GetName(), load.Namespace, load.Name)

	set := load.Spec.Ref.(runtime.Object)
	cliSet, ok := set.(client.Object)
	if !ok {
		return errors.New("fail to convert runtime.Object to client.Object")
	}
	return d.Delete(context.TODO(), cliSet, client.PropagationPolicy(metav1.DeletePropagationBackground))
}

// ApplyTemplate updates the object to the latest revision, depending on the DcpAppDaemon.
func (a *DeploymentControllor) applyTemplate(scheme *runtime.Scheme, yad *v1alpha1.DcpAppDaemon, nodepool v1alpha1.NodePool, revision string, set *appsv1.Deployment) error {

	if set.Labels == nil {
		set.Labels = map[string]string{}
	}
	for k, v := range yad.Spec.WorkloadTemplate.DeploymentTemplate.Labels {
		set.Labels[k] = v
	}
	for k, v := range yad.Spec.Selector.MatchLabels {
		set.Labels[k] = v
	}
	set.Labels[v1alpha1.ControllerRevisionHashLabelKey] = revision
	set.Labels[v1alpha1.PoolNameLabelKey] = nodepool.GetName()

	if set.Annotations == nil {
		set.Annotations = map[string]string{}
	}
	for k, v := range yad.Spec.WorkloadTemplate.DeploymentTemplate.Annotations {
		set.Annotations[k] = v
	}
	set.Annotations[v1alpha1.AnnotationRefNodePool] = nodepool.GetName()

	set.Namespace = yad.GetNamespace()
	set.GenerateName = getWorkloadPrefix(yad.GetName(), nodepool.GetName())

	set.Spec = *yad.Spec.WorkloadTemplate.DeploymentTemplate.Spec.DeepCopy()
	set.Spec.Selector.MatchLabels[v1alpha1.PoolNameLabelKey] = nodepool.GetName()

	// set RequiredDuringSchedulingIgnoredDuringExecution nil
	if set.Spec.Template.Spec.Affinity != nil && set.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
		set.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = nil
	}

	if set.Spec.Template.Labels == nil {
		set.Spec.Template.Labels = map[string]string{}
	}
	set.Spec.Template.Labels[v1alpha1.PoolNameLabelKey] = nodepool.GetName()
	set.Spec.Template.Labels[v1alpha1.ControllerRevisionHashLabelKey] = revision

	// use nodeSelector
	set.Spec.Template.Spec.NodeSelector = CreateNodeSelectorByNodepoolName(nodepool.GetName())

	// toleration
	set.Spec.Template.Spec.Tolerations = TaintsToTolerations(nodepool.Spec.Taints)

	if err := controllerutil.SetControllerReference(yad, set, scheme); err != nil {
		return err
	}
	return nil
}

func (d *DeploymentControllor) ObjectKey(load *Workload) client.ObjectKey {
	return types.NamespacedName{
		Namespace: load.Namespace,
		Name:      load.Name,
	}
}

func (d *DeploymentControllor) UpdateWorkload(load *Workload, yad *v1alpha1.DcpAppDaemon, nodepool v1alpha1.NodePool, revision string) error {
	klog.Infof("DcpAppDaemon[%s/%s] prepare update Deployment[%s/%s]", yad.GetNamespace(),
		yad.GetName(), load.Namespace, load.Name)

	deploy := &appsv1.Deployment{}
	var updateError error
	for i := 0; i < updateRetries; i++ {
		getError := d.Client.Get(context.TODO(), d.ObjectKey(load), deploy)
		if getError != nil {
			return getError
		}

		if err := d.applyTemplate(d.Scheme, yad, nodepool, revision, deploy); err != nil {
			return err
		}
		updateError = d.Client.Update(context.TODO(), deploy)
		if updateError == nil {
			break
		}
	}

	return updateError
}

func (d *DeploymentControllor) CreateWorkload(yad *v1alpha1.DcpAppDaemon, nodepool v1alpha1.NodePool, revision string) error {
	klog.Infof("DcpAppDaemon[%s/%s] prepare create new deployment by nodepool %s ", yad.GetNamespace(), yad.GetName(), nodepool.GetName())

	deploy := appsv1.Deployment{}
	if err := d.applyTemplate(d.Scheme, yad, nodepool, revision, &deploy); err != nil {
		klog.Errorf("DcpAppDaemon[%s/%s] faild to apply template, when create deployment: %v", yad.GetNamespace(),
			yad.GetName(), err)
		return err
	}
	return d.Client.Create(context.TODO(), &deploy)
}

func (d *DeploymentControllor) GetAllWorkloads(set *v1alpha1.DcpAppDaemon) ([]*Workload, error) {
	allDeployments := appsv1.DeploymentList{}
	// DcpAppDaemon Deployment, OwnerRef
	selector, err := metav1.LabelSelectorAsSelector(set.Spec.Selector)
	if err != nil {
		return nil, err
	}
	// List all Deployment to include those that don't match the selector anymore but
	// have a ControllerRef pointing to this controller.
	if err := d.Client.List(context.TODO(), &allDeployments, &client.ListOptions{LabelSelector: selector}); err != nil {
		return nil, err
	}

	manager, err := refmanager.New(d.Client, set.Spec.Selector, set, d.Scheme)
	if err != nil {
		return nil, err
	}

	selected := make([]metav1.Object, 0, len(allDeployments.Items))
	for i := 0; i < len(allDeployments.Items); i++ {
		t := allDeployments.Items[i]
		selected = append(selected, &t)
	}

	objs, err := manager.ClaimOwnedObjects(selected)
	if err != nil {
		return nil, err
	}

	workloads := make([]*Workload, 0, len(objs))
	for i, o := range objs {
		deploy := o.(*appsv1.Deployment)
		spec := deploy.Spec
		w := &Workload{
			Name:      o.GetName(),
			Namespace: o.GetNamespace(),
			Kind:      deploy.Kind,
			Spec: WorkloadSpec{
				Ref:          objs[i],
				NodeSelector: spec.Template.Spec.NodeSelector,
				Toleration:   spec.Template.Spec.Tolerations,
			},
		}
		workloads = append(workloads, w)
	}
	return workloads, nil
}

var _ WorkloadControllor = &DeploymentControllor{}
