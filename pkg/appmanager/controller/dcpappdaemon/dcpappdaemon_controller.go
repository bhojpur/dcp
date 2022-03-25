package dcpappdaemon

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
	"flag"
	"fmt"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	unitv1alpha1 "github.com/bhojpur/dcp/pkg/appmanager/apis/apps/v1alpha1"
	"github.com/bhojpur/dcp/pkg/appmanager/controller/dcpappdaemon/workloadcontroller"
	"github.com/bhojpur/dcp/pkg/appmanager/util"
	"github.com/bhojpur/dcp/pkg/appmanager/util/gate"
)

var (
	concurrentReconciles = 3
)

const (
	controllerName            = "appdaemon-controller"
	slowStartInitialBatchSize = 1

	eventTypeRevisionProvision  = "RevisionProvision"
	eventTypeTemplateController = "TemplateController"

	eventTypeWorkloadsCreated = "CreateWorkload"
	eventTypeWorkloadsUpdated = "UpdateWorkload"
	eventTypeWorkloadsDeleted = "DeleteWorkload"
)

func init() {
	flag.IntVar(&concurrentReconciles, "appdaemon-workers", concurrentReconciles, "Max concurrent workers for DcpAppDaemon controller.")
}

// Add creates a new DcpAppDaemon Controller and adds it to the Manager with default RBAC.
// The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, _ context.Context) error {
	if !gate.ResourceEnabled(&unitv1alpha1.DcpAppDaemon{}) {
		return nil
	}
	return add(mgr, newReconciler(mgr))
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: r, MaxConcurrentReconciles: concurrentReconciles})
	if err != nil {
		return err
	}

	// Watch for changes to DcpAppDaemon
	err = c.Watch(&source.Kind{Type: &unitv1alpha1.DcpAppDaemon{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to NodePool
	err = c.Watch(&source.Kind{Type: &unitv1alpha1.NodePool{}}, &EnqueueAppDaemonForNodePool{client: mgr.GetClient()})
	if err != nil {
		return err
	}
	return nil
}

var _ reconcile.Reconciler = &ReconcileAppDaemon{}

// ReconcileAppDaemon reconciles a DcpAppDaemon object
type ReconcileAppDaemon struct {
	client.Client
	scheme *runtime.Scheme

	recorder record.EventRecorder
	controls map[unitv1alpha1.TemplateType]workloadcontroller.WorkloadControllor
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileAppDaemon{
		Client: mgr.GetClient(),
		scheme: mgr.GetScheme(),

		recorder: mgr.GetEventRecorderFor(controllerName),
		controls: map[unitv1alpha1.TemplateType]workloadcontroller.WorkloadControllor{
			//			unitv1alpha1.StatefulSetTemplateType: &StatefulSetControllor{Client: mgr.GetClient(), scheme: mgr.GetScheme()},
			unitv1alpha1.DeploymentTemplateType: &workloadcontroller.DeploymentControllor{Client: mgr.GetClient(), Scheme: mgr.GetScheme()},
		},
	}
}

// +kubebuilder:rbac:groups=apps.bhojpur.net,resources=dcpappdaemons,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.bhojpur.net,resources=dcpappdaemons/status,verbs=get;update;patch

// Reconcile reads that state of the cluster for a DcpAppDaemon object and makes changes based on the state read
// and what is in the DcpAppDaemon.Spec
func (r *ReconcileAppDaemon) Reconcile(_ context.Context, request reconcile.Request) (reconcile.Result, error) {
	klog.V(4).Infof("Reconcile DcpAppDaemon %s/%s", request.Namespace, request.Name)
	// Fetch the DcpAppDaemon instance
	instance := &unitv1alpha1.DcpAppDaemon{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if instance.DeletionTimestamp != nil {
		return reconcile.Result{}, nil
	}

	oldStatus := instance.Status.DeepCopy()

	currentRevision, updatedRevision, collisionCount, err := r.constructAppDaemonRevisions(instance)
	if err != nil {
		klog.Errorf("Fail to construct controller revision of DcpAppDaemon %s/%s: %s", instance.Namespace, instance.Name, err)
		r.recorder.Event(instance.DeepCopy(), corev1.EventTypeWarning, fmt.Sprintf("Failed%s", eventTypeRevisionProvision), err.Error())
		return reconcile.Result{}, err
	}

	expectedRevision := currentRevision
	if updatedRevision != nil {
		expectedRevision = updatedRevision
	}

	klog.Infof("DcpAppDaemon [%s/%s] get expectRevision %v collisionCount %v", instance.GetNamespace(), instance.GetName(),
		expectedRevision.Name, collisionCount)

	control, templateType, err := r.getTemplateControls(instance)
	if err != nil {
		r.recorder.Event(instance.DeepCopy(), corev1.EventTypeWarning, fmt.Sprintf("Failed%s", eventTypeTemplateController), err.Error())
		return reconcile.Result{}, err
	}

	currentNPToWorkload, err := r.getNodePoolToWorkLoad(instance, control)
	if err != nil {
		klog.Errorf("DcpAppDaemon[%s/%s] Fail to get nodePoolWorkload, error: %s", instance.Namespace, instance.Name, err)
		return reconcile.Result{}, nil
	}

	allNameToNodePools, err := r.getNameToNodePools(instance)
	if err != nil {
		klog.Errorf("DcpAppDaemon[%s/%s] Fail to get nameToNodePools, error: %s", instance.Namespace, instance.Name, err)
		return reconcile.Result{}, nil
	}

	newStatus, err := r.manageWorkloads(instance, currentNPToWorkload, allNameToNodePools, expectedRevision.Name, templateType)
	if err != nil {
		return reconcile.Result{}, nil
	}

	return r.updateStatus(instance, newStatus, oldStatus, currentRevision, collisionCount, templateType)
}

func (r *ReconcileAppDaemon) updateStatus(instance *unitv1alpha1.DcpAppDaemon, newStatus, oldStatus *unitv1alpha1.DcpAppDaemonStatus,
	currentRevision *appsv1.ControllerRevision, collisionCount int32, templateType unitv1alpha1.TemplateType) (reconcile.Result, error) {

	newStatus = r.calculateStatus(instance, newStatus, currentRevision, collisionCount, templateType)
	_, err := r.updateAppDaemon(instance, oldStatus, newStatus)

	return reconcile.Result{}, err
}

func (r *ReconcileAppDaemon) updateAppDaemon(yad *unitv1alpha1.DcpAppDaemon, oldStatus, newStatus *unitv1alpha1.DcpAppDaemonStatus) (*unitv1alpha1.DcpAppDaemon, error) {
	if oldStatus.CurrentRevision == newStatus.CurrentRevision &&
		*oldStatus.CollisionCount == *newStatus.CollisionCount &&
		oldStatus.TemplateType == newStatus.TemplateType &&
		yad.Generation == newStatus.ObservedGeneration &&
		reflect.DeepEqual(oldStatus.NodePools, newStatus.NodePools) &&
		reflect.DeepEqual(oldStatus.Conditions, newStatus.Conditions) {
		klog.Infof("DcpAppDaemon[%s/%s] oldStatus==newStatus, no need to update status", yad.GetNamespace(), yad.GetName())
		return yad, nil
	}

	newStatus.ObservedGeneration = yad.Generation

	var getErr, updateErr error
	for i, obj := 0, yad; ; i++ {
		klog.V(4).Infof(fmt.Sprintf("DcpAppDaemon[%s/%s] The %d th time updating status for %v[%s/%s], ",
			yad.GetNamespace(), yad.GetName(), i, obj.Kind, obj.Namespace, obj.Name) +
			fmt.Sprintf("sequence No: %v->%v", obj.Status.ObservedGeneration, newStatus.ObservedGeneration))

		obj.Status = *newStatus

		updateErr = r.Client.Status().Update(context.TODO(), obj)
		if updateErr == nil {
			return obj, nil
		}
		if i >= updateRetries {
			break
		}
		tmpObj := &unitv1alpha1.DcpAppDaemon{}
		if getErr = r.Client.Get(context.TODO(), client.ObjectKey{Namespace: obj.Namespace, Name: obj.Name}, tmpObj); getErr != nil {
			return nil, getErr
		}
		obj = tmpObj
	}

	klog.Errorf("fail to update DcpAppDaemon %s/%s status: %s", yad.Namespace, yad.Name, updateErr)
	return nil, updateErr
}

func (r *ReconcileAppDaemon) calculateStatus(instance *unitv1alpha1.DcpAppDaemon, newStatus *unitv1alpha1.DcpAppDaemonStatus,
	currentRevision *appsv1.ControllerRevision, collisionCount int32, templateType unitv1alpha1.TemplateType) *unitv1alpha1.DcpAppDaemonStatus {

	newStatus.CollisionCount = &collisionCount

	if newStatus.CurrentRevision == "" {
		// init with current revision
		newStatus.CurrentRevision = currentRevision.Name
	}

	newStatus.TemplateType = templateType

	return newStatus
}

func (r *ReconcileAppDaemon) manageWorkloads(instance *unitv1alpha1.DcpAppDaemon, currentNodepoolToWorkload map[string]*workloadcontroller.Workload,
	allNameToNodePools map[string]unitv1alpha1.NodePool, expectedRevision string, templateType unitv1alpha1.TemplateType) (newStatus *unitv1alpha1.DcpAppDaemonStatus, updateErr error) {

	newStatus = instance.Status.DeepCopy()

	nps := make([]string, 0, len(allNameToNodePools))
	for np, _ := range allNameToNodePools {
		nps = append(nps, np)
	}
	newStatus.NodePools = nps

	needDeleted, needUpdate, needCreate := r.classifyWorkloads(instance, currentNodepoolToWorkload, allNameToNodePools, expectedRevision)
	provision, err := r.manageWorkloadsProvision(instance, allNameToNodePools, expectedRevision, templateType, needDeleted, needCreate)
	if err != nil {
		SetAppDaemonCondition(newStatus, NewAppDaemonCondition(unitv1alpha1.WorkLoadProvisioned, corev1.ConditionFalse, "Error", err.Error()))
		return newStatus, fmt.Errorf("fail to manage workload provision: %v", err)
	}

	if provision {
		SetAppDaemonCondition(newStatus, NewAppDaemonCondition(unitv1alpha1.WorkLoadProvisioned, corev1.ConditionTrue, "", ""))
	}

	if len(needUpdate) > 0 {
		_, updateErr = util.SlowStartBatch(len(needUpdate), slowStartInitialBatchSize, func(index int) error {
			u := needUpdate[index]
			updateWorkloadErr := r.controls[templateType].UpdateWorkload(u, instance, allNameToNodePools[u.GetNodePoolName()], expectedRevision)
			if updateWorkloadErr != nil {
				r.recorder.Event(instance.DeepCopy(), corev1.EventTypeWarning, fmt.Sprintf("Failed %s", eventTypeWorkloadsUpdated),
					fmt.Sprintf("Error updating workload type(%s) %s when updating: %s", templateType, u.Name, updateWorkloadErr))
				klog.Errorf("DcpAppDaemon[%s/%s] update workload[%s/%s/%s] error %v", instance.GetNamespace(), instance.GetName(),
					templateType, u.Namespace, u.Name, err)
			}
			return updateWorkloadErr
		})
	}

	if updateErr == nil {
		SetAppDaemonCondition(newStatus, NewAppDaemonCondition(unitv1alpha1.WorkLoadUpdated, corev1.ConditionTrue, "", ""))
	} else {
		SetAppDaemonCondition(newStatus, NewAppDaemonCondition(unitv1alpha1.WorkLoadUpdated, corev1.ConditionFalse, "Error", updateErr.Error()))
	}

	return newStatus, updateErr
}

func (r *ReconcileAppDaemon) manageWorkloadsProvision(instance *unitv1alpha1.DcpAppDaemon,
	allNameToNodePools map[string]unitv1alpha1.NodePool, expectedRevision string, templateType unitv1alpha1.TemplateType,
	needDeleted []*workloadcontroller.Workload, needCreate []string) (bool, error) {
	// Create

	var errs []error
	if len(needCreate) > 0 {
		// do not consider deletion
		var createdNum int
		var createdErr error
		createdNum, createdErr = util.SlowStartBatch(len(needCreate), slowStartInitialBatchSize, func(idx int) error {
			nodepoolName := needCreate[idx]
			err := r.controls[templateType].CreateWorkload(instance, allNameToNodePools[nodepoolName], expectedRevision)
			//err := r.poolControls[workloadType].CreatePool(ud, poolName, revision, replicas)
			if err != nil {
				klog.Errorf("DcpAppDaemon[%s/%s] templatetype %s create workload by nodepool %s error: %s",
					instance.GetNamespace(), instance.GetName(), templateType, nodepoolName, err.Error())
				if !errors.IsTimeout(err) {
					return fmt.Errorf("DcpAppDaemon[%s/%s] templatetype %s create workload by nodepool %s error: %s",
						instance.GetNamespace(), instance.GetName(), templateType, nodepoolName, err.Error())
				}
			}
			klog.Infof("DcpAppDaemon[%s/%s] templatetype %s create workload by nodepool %s success",
				instance.GetNamespace(), instance.GetName(), templateType, nodepoolName)
			return nil
		})
		if createdErr == nil {
			r.recorder.Eventf(instance.DeepCopy(), corev1.EventTypeNormal, fmt.Sprintf("Successful %s", eventTypeWorkloadsCreated), "Create %d Workload type(%s)", createdNum, templateType)
		} else {
			errs = append(errs, createdErr)
		}
	}

	// manage deleting
	if len(needDeleted) > 0 {
		var deleteErrs []error
		// need deleted
		for _, d := range needDeleted {
			if err := r.controls[templateType].DeleteWorkload(instance, d); err != nil {
				deleteErrs = append(deleteErrs, fmt.Errorf("DcpAppDaemon[%s/%s] delete workload[%s/%s/%s] error %v",
					instance.GetNamespace(), instance.GetName(), templateType, d.Namespace, d.Name, err))
			}
		}
		if len(deleteErrs) > 0 {
			errs = append(errs, deleteErrs...)
		} else {
			r.recorder.Eventf(instance.DeepCopy(), corev1.EventTypeNormal, fmt.Sprintf("Successful %s", eventTypeWorkloadsDeleted), "Delete %d Workload type(%s)", len(needDeleted), templateType)
		}
	}

	return len(needCreate) > 0 || len(needDeleted) > 0, utilerrors.NewAggregate(errs)
}

func (r *ReconcileAppDaemon) classifyWorkloads(instance *unitv1alpha1.DcpAppDaemon, currentNodepoolToWorkload map[string]*workloadcontroller.Workload,
	allNameToNodePools map[string]unitv1alpha1.NodePool, expectedRevision string) (needDeleted, needUpdate []*workloadcontroller.Workload,
	needCreate []string) {

	for npName, load := range currentNodepoolToWorkload {
		if np, ok := allNameToNodePools[npName]; ok {
			match := true
			// judge workload NodeSelector
			if !reflect.DeepEqual(load.GetNodeSelector(), workloadcontroller.CreateNodeSelectorByNodepoolName(npName)) {
				match = false
			}
			// judge workload whether toleration all taints
			match = IsTolerationsAllTaints(load.GetToleration(), np.Spec.Taints)

			// judge revision
			if load.GetRevision() != expectedRevision {
				match = false
			}

			if !match {
				klog.V(4).Infof("DcpAppDaemon[%s/%s] need update [%s/%s/%s]", instance.GetNamespace(),
					instance.GetName(), load.GetKind(), load.Namespace, load.Name)
				needUpdate = append(needUpdate, load)
			}
		} else {
			needDeleted = append(needDeleted, load)
			klog.V(4).Infof("DcpAppDaemon[%s/%s] need delete [%s/%s/%s]", instance.GetNamespace(),
				instance.GetName(), load.GetKind(), load.Namespace, load.Name)
		}
	}

	for vnp, _ := range allNameToNodePools {
		if _, ok := currentNodepoolToWorkload[vnp]; !ok {
			needCreate = append(needCreate, vnp)
			klog.V(4).Infof("DcpAppDaemon[%s/%s] need create new workload by nodepool %s", instance.GetNamespace(),
				instance.GetName(), vnp)
		}
	}

	return
}

func (r *ReconcileAppDaemon) getNameToNodePools(instance *unitv1alpha1.DcpAppDaemon) (map[string]unitv1alpha1.NodePool, error) {
	klog.V(4).Infof("DcpAppDaemon [%s/%s] prepare to get associated nodepools",
		instance.Namespace, instance.Name)

	nodepoolSelector, err := metav1.LabelSelectorAsSelector(instance.Spec.NodePoolSelector)
	if err != nil {
		return nil, err
	}

	nodepools := unitv1alpha1.NodePoolList{}
	if err := r.Client.List(context.TODO(), &nodepools, &client.ListOptions{LabelSelector: nodepoolSelector}); err != nil {
		klog.Errorf("DcpAppDaemon [%s/%s] Fail to get NodePoolList", instance.GetNamespace(),
			instance.GetName())
		return nil, nil
	}

	indexs := make(map[string]unitv1alpha1.NodePool)
	for i, v := range nodepools.Items {
		indexs[v.GetName()] = v
		klog.V(4).Infof("DcpAppDaemon [%s/%s] get %d's associated nodepools %s",
			instance.Namespace, instance.Name, i, v.Name)

	}

	return indexs, nil
}

func (r *ReconcileAppDaemon) getTemplateControls(instance *unitv1alpha1.DcpAppDaemon) (workloadcontroller.WorkloadControllor,
	unitv1alpha1.TemplateType, error) {
	switch {
	case instance.Spec.WorkloadTemplate.StatefulSetTemplate != nil:
		return r.controls[unitv1alpha1.StatefulSetTemplateType], unitv1alpha1.StatefulSetTemplateType, nil
	case instance.Spec.WorkloadTemplate.DeploymentTemplate != nil:
		return r.controls[unitv1alpha1.DeploymentTemplateType], unitv1alpha1.DeploymentTemplateType, nil
	default:
		klog.Errorf("The appropriate WorkloadTemplate was not found")
		return nil, "", fmt.Errorf("The appropriate WorkloadTemplate was not found, Now Support(%s/%s)",
			unitv1alpha1.StatefulSetTemplateType, unitv1alpha1.DeploymentTemplateType)
	}
}

func (r *ReconcileAppDaemon) getNodePoolToWorkLoad(instance *unitv1alpha1.DcpAppDaemon, c workloadcontroller.WorkloadControllor) (map[string]*workloadcontroller.Workload, error) {
	klog.V(4).Infof("DcpAppDaemon [%s/%s/%s] prepare to get all workload", c.GetTemplateType(), instance.Namespace, instance.Name)

	nodePoolsToWorkloads := make(map[string]*workloadcontroller.Workload)
	workloads, err := c.GetAllWorkloads(instance)
	if err != nil {
		klog.Errorf("Get all workloads for DcpAppDaemon[%s/%s] error %v", instance.GetNamespace(),
			instance.GetName(), err)
		return nil, err
	}
	// workload NodePool
	for i, w := range workloads {
		if w.GetNodePoolName() != "" {
			nodePoolsToWorkloads[w.GetNodePoolName()] = workloads[i]
			klog.V(4).Infof("DcpAppDaemon [%s/%s] get %d's workload[%s/%s/%s]",
				instance.Namespace, instance.Name, i, c.GetTemplateType(), w.Namespace, w.Name)
		} else {
			klog.Warningf("DcpAppDaemon [%s/%s] %d's workload[%s/%s/%s] has no nodepool annotation",
				instance.Namespace, instance.Name, i, c.GetTemplateType(), w.Namespace, w.Name)
		}
	}
	klog.V(4).Infof("DcpAppDaemon [%s/%s] get %d %s workloads",
		instance.Namespace, instance.Name, len(nodePoolsToWorkloads), c.GetTemplateType())
	return nodePoolsToWorkloads, nil
}
