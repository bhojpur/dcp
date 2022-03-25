package dcpingress

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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	appsv1alpha1 "github.com/bhojpur/dcp/pkg/appmanager/apis/apps/v1alpha1"
	"github.com/bhojpur/dcp/pkg/appmanager/util/gate"
	dcpapputil "github.com/bhojpur/dcp/pkg/appmanager/util/kubernetes"
	"github.com/bhojpur/dcp/pkg/appmanager/util/refmanager"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	controllerName         = "ingress-controller"
	ingressDeploymentLabel = "ingress.io/nodepool"
)

const updateRetries = 5

// IngressReconciler reconciles a DcpIngress object
type IngressReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
}

// Add creates a new DcpIngress Controller and adds it to the Manager with default RBAC.
// The Manager will set fields on the Controller and start it when the Manager is started.
func Add(mgr manager.Manager, ctx context.Context) error {
	if !gate.ResourceEnabled(&appsv1alpha1.DcpIngress{}) {
		return nil
	}
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
//func newReconciler(mgr manager.Manager, createSingletonPoolIngress bool) reconcile.Reconciler {
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &IngressReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		recorder: mgr.GetEventRecorderFor(controllerName),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	// Watch for changes to DcpIngress
	err = c.Watch(&source.Kind{Type: &appsv1alpha1.DcpIngress{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsv1alpha1.DcpIngress{},
	})
	if err != nil {
		return err
	}
	return nil
}

// +kubebuilder:rbac:groups=apps.bhojpur.net,resources=dcpingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.bhojpur.net,resources=dcpingresses/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=*
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=*
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=*
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=*

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *IngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	klog.V(4).Infof("Reconcile DcpIngress: %s", req.Name)
	if req.Name != appsv1alpha1.SingletonDcpIngressInstanceName {
		return ctrl.Result{}, nil
	}
	// Fetch the DcpIngress instance
	instance := &appsv1alpha1.DcpIngress{}
	err := r.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	// Add finalizer if not exist
	if !controllerutil.ContainsFinalizer(instance, appsv1alpha1.DcpIngressFinalizer) {
		controllerutil.AddFinalizer(instance, appsv1alpha1.DcpIngressFinalizer)
		if err := r.Update(context.TODO(), instance); err != nil {
			return ctrl.Result{}, err
		}
	}
	// Handle ingress controller resources cleanup
	if !instance.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.cleanupIngressResources(instance)
	}
	// Set the default version at current stage
	instance.Status.Version = appsv1alpha1.NginxIngressControllerVersion

	var desiredPoolNames, currentPoolNames []string
	desiredPoolNames = getDesiredPoolNames(instance)
	currentPoolNames = getCurrentPoolNames(instance)
	isIngressCRChanged := false
	addedPools, removedPools, unchangedPools := getPools(desiredPoolNames, currentPoolNames)
	if addedPools != nil {
		klog.V(4).Infof("added pool list is %s", addedPools)
		isIngressCRChanged = true
		ownerRef := prepareDeploymentOwnerReferences(instance)
		if currentPoolNames == nil {
			if err := dcpapputil.CreateNginxIngressCommonResource(r.Client); err != nil {
				return ctrl.Result{}, err
			}
		}
		for _, pool := range addedPools {
			replicas := instance.Spec.Replicas
			if err := dcpapputil.CreateNginxIngressSpecificResource(r.Client, pool, replicas, ownerRef); err != nil {
				return ctrl.Result{}, err
			}
			notReadyPool := appsv1alpha1.IngressNotReadyPool{Name: pool, Info: nil}
			instance.Status.Conditions.IngressNotReadyPools = append(instance.Status.Conditions.IngressNotReadyPools, notReadyPool)
			instance.Status.UnreadyNum += 1
		}
	}
	if removedPools != nil {
		klog.V(4).Infof("removed pool list is %s", removedPools)
		isIngressCRChanged = true
		for _, pool := range removedPools {
			if desiredPoolNames == nil {
				if err := dcpapputil.DeleteNginxIngressSpecificResource(r.Client, pool, true); err != nil {
					return ctrl.Result{}, err
				}
			} else {
				if err := dcpapputil.DeleteNginxIngressSpecificResource(r.Client, pool, false); err != nil {
					return ctrl.Result{}, err
				}
			}
			if desiredPoolNames != nil && !removePoolfromCondition(instance, pool) {
				klog.V(4).Infof("Pool/%s is not found from conditions!", pool)
			}
		}
		if desiredPoolNames == nil {
			if err := dcpapputil.DeleteNginxIngressCommonResource(r.Client); err != nil {
				return ctrl.Result{}, err
			}
			instance.Status.Conditions.IngressReadyPools = nil
			instance.Status.Conditions.IngressNotReadyPools = nil
			instance.Status.ReadyNum = 0
			instance.Status.UnreadyNum = 0
		}
	}
	if unchangedPools != nil {
		klog.V(4).Infof("unchanged pool list is %s", unchangedPools)
		desiredReplicas := instance.Spec.Replicas
		currentReplicas := instance.Status.Replicas
		if desiredReplicas != currentReplicas {
			klog.V(4).Infof("Per-Pool ingress controller replicas is changed!")
			isIngressCRChanged = true
			for _, pool := range unchangedPools {
				if err := dcpapputil.ScaleNginxIngressControllerDeploymment(r.Client, pool, desiredReplicas); err != nil {
					return ctrl.Result{}, err
				}
			}
		}
	}
	r.updateStatus(instance, isIngressCRChanged)
	return ctrl.Result{}, nil
}

func getPools(desired, current []string) (added, removed, unchanged []string) {
	swap := false
	for i := 0; i < 2; i++ {
		for _, s1 := range desired {
			found := false
			for _, s2 := range current {
				if s1 == s2 {
					found = true
					if !swap {
						unchanged = append(unchanged, s1)
					}
					break
				}
			}
			if !found {
				if !swap {
					added = append(added, s1)
				} else {
					removed = append(removed, s1)
				}
			}
		}
		if i == 0 {
			swap = true
			desired, current = current, desired
		}
	}
	return added, removed, unchanged
}

func getDesiredPoolNames(ying *appsv1alpha1.DcpIngress) []string {
	var desiredPoolNames []string
	for _, pool := range ying.Spec.Pools {
		desiredPoolNames = append(desiredPoolNames, pool.Name)
	}
	return desiredPoolNames
}

func getCurrentPoolNames(ying *appsv1alpha1.DcpIngress) []string {
	var currentPoolNames []string
	currentPoolNames = ying.Status.Conditions.IngressReadyPools
	for _, pool := range ying.Status.Conditions.IngressNotReadyPools {
		currentPoolNames = append(currentPoolNames, pool.Name)
	}
	return currentPoolNames
}

func removePoolfromCondition(ying *appsv1alpha1.DcpIngress, poolname string) bool {
	for i, pool := range ying.Status.Conditions.IngressReadyPools {
		if pool == poolname {
			length := len(ying.Status.Conditions.IngressReadyPools)
			if i == length-1 {
				ying.Status.Conditions.IngressReadyPools = ying.Status.Conditions.IngressReadyPools[:i]
			} else {
				ying.Status.Conditions.IngressReadyPools = append(ying.Status.Conditions.IngressReadyPools[:i],
					ying.Status.Conditions.IngressReadyPools[i+1:]...)
			}
			if ying.Status.ReadyNum >= 1 {
				ying.Status.ReadyNum -= 1
			}
			return true
		}
	}
	for i, pool := range ying.Status.Conditions.IngressNotReadyPools {
		if pool.Name == poolname {
			length := len(ying.Status.Conditions.IngressNotReadyPools)
			if i == length-1 {
				ying.Status.Conditions.IngressNotReadyPools = ying.Status.Conditions.IngressNotReadyPools[:i]
			} else {
				ying.Status.Conditions.IngressNotReadyPools = append(ying.Status.Conditions.IngressNotReadyPools[:i],
					ying.Status.Conditions.IngressNotReadyPools[i+1:]...)
			}
			if ying.Status.UnreadyNum >= 1 {
				ying.Status.UnreadyNum -= 1
			}
			return true
		}
	}
	return false
}

func (r *IngressReconciler) updateStatus(ying *appsv1alpha1.DcpIngress, ingressCRChanged bool) error {
	ying.Status.Replicas = ying.Spec.Replicas
	if !ingressCRChanged {
		deployments, err := r.getAllDeployments(ying)
		if err != nil {
			klog.V(4).Infof("Get all the ingress controller deployments err: %v", err)
			return err
		}
		ying.Status.Conditions.IngressReadyPools = nil
		ying.Status.Conditions.IngressNotReadyPools = nil
		ying.Status.ReadyNum = 0
		for _, dply := range deployments {
			pool := dply.ObjectMeta.GetLabels()[ingressDeploymentLabel]
			if dply.Status.ReadyReplicas == ying.Spec.Replicas {
				klog.V(4).Infof("Ingress on pool %s is ready!", pool)
				ying.Status.ReadyNum += 1
				ying.Status.Conditions.IngressReadyPools = append(ying.Status.Conditions.IngressReadyPools, pool)
			} else {
				klog.V(4).Infof("Ingress on pool %s is NOT ready!", pool)
				condition := getUnreadyDeploymentCondition(dply)
				if condition == nil {
					klog.V(4).Infof("Get deployment/%s conditions nil!", dply.GetName())
				} else {
					notReadyPool := appsv1alpha1.IngressNotReadyPool{Name: pool, Info: condition}
					ying.Status.Conditions.IngressNotReadyPools = append(ying.Status.Conditions.IngressNotReadyPools, notReadyPool)
				}
			}
		}
		ying.Status.UnreadyNum = int32(len(ying.Spec.Pools)) - ying.Status.ReadyNum
	}
	var updateErr error
	for i, obj := 0, ying; i < updateRetries; i++ {
		updateErr = r.Status().Update(context.TODO(), obj)
		if updateErr == nil {
			klog.V(4).Infof("%s status is updated!", obj.Name)
			return nil
		}
	}
	klog.Errorf("Fail to update DcpIngress %s status: %v", ying.Name, updateErr)
	return updateErr
}

func (r *IngressReconciler) cleanupIngressResources(instance *appsv1alpha1.DcpIngress) (ctrl.Result, error) {
	pools := getDesiredPoolNames(instance)
	if pools != nil {
		for _, pool := range pools {
			if err := dcpapputil.DeleteNginxIngressSpecificResource(r.Client, pool, true); err != nil {
				return ctrl.Result{}, err
			}
		}
		if err := dcpapputil.DeleteNginxIngressCommonResource(r.Client); err != nil {
			return ctrl.Result{}, err
		}
	}
	if controllerutil.ContainsFinalizer(instance, appsv1alpha1.DcpIngressFinalizer) {
		controllerutil.RemoveFinalizer(instance, appsv1alpha1.DcpIngressFinalizer)
		if err := r.Update(context.TODO(), instance); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func prepareDeploymentOwnerReferences(instance *appsv1alpha1.DcpIngress) *metav1.OwnerReference {
	isController := true
	isBlockOwnerDeletion := true
	ownerRef := metav1.OwnerReference{
		//TODO: optimize the APIVersion/Kind with instance values
		APIVersion:         "apps.bhojpur.net/v1alpha1",
		Kind:               "DcpIngress",
		Name:               instance.Name,
		UID:                instance.UID,
		Controller:         &isController,
		BlockOwnerDeletion: &isBlockOwnerDeletion,
	}
	return &ownerRef
}

// getAllDeployments returns all of deployments owned by DcpIngress
func (r *IngressReconciler) getAllDeployments(ying *appsv1alpha1.DcpIngress) ([]*appsv1.Deployment, error) {
	labelSelector := metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      ingressDeploymentLabel,
				Operator: metav1.LabelSelectorOpExists,
			},
		},
	}
	selector, err := metav1.LabelSelectorAsSelector(&labelSelector)
	if err != nil {
		return nil, err
	}

	dplyList := &appsv1.DeploymentList{}
	err = r.Client.List(context.TODO(), dplyList, &client.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, err
	}

	manager, err := refmanager.New(r.Client, &labelSelector, ying, r.Scheme)
	if err != nil {
		return nil, err
	}

	selected := make([]metav1.Object, len(dplyList.Items))
	for i, dply := range dplyList.Items {
		selected[i] = dply.DeepCopy()
	}
	claimed, err := manager.ClaimOwnedObjects(selected)
	if err != nil {
		return nil, err
	}

	claimedDplys := make([]*appsv1.Deployment, len(claimed))
	for i, dply := range claimed {
		claimedDplys[i] = dply.(*appsv1.Deployment)
	}
	return claimedDplys, nil
}

func getUnreadyDeploymentCondition(dply *appsv1.Deployment) (info *appsv1alpha1.IngressNotReadyConditionInfo) {
	len := len(dply.Status.Conditions)
	if len == 0 {
		return nil
	}
	var conditionInfo appsv1alpha1.IngressNotReadyConditionInfo
	condition := dply.Status.Conditions[len-1]
	if condition.Type == appsv1.DeploymentReplicaFailure {
		conditionInfo.Type = appsv1alpha1.IngressFailure
	} else {
		conditionInfo.Type = appsv1alpha1.IngressPending
	}
	conditionInfo.LastTransitionTime = condition.LastTransitionTime
	conditionInfo.Message = condition.Message
	conditionInfo.Reason = condition.Reason
	return &conditionInfo
}
