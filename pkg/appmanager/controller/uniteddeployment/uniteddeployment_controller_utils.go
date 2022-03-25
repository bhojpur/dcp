package uniteddeployment

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

	unitv1alpha1 "github.com/bhojpur/dcp/pkg/appmanager/apis/apps/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const updateRetries = 5

type UnitedDeploymentPatches struct {
	Replicas int32
	Patch    string
}

func getPoolNameFrom(metaObj metav1.Object) (string, error) {
	name, exist := metaObj.GetLabels()[unitv1alpha1.PoolNameLabelKey]
	if !exist {
		return "", fmt.Errorf("fail to get pool name from label of pool %s/%s: no label %s found", metaObj.GetNamespace(), metaObj.GetName(), unitv1alpha1.PoolNameLabelKey)
	}

	if len(name) == 0 {
		return "", fmt.Errorf("fail to get pool name from label of pool %s/%s: label %s has an empty value", metaObj.GetNamespace(), metaObj.GetName(), unitv1alpha1.PoolNameLabelKey)
	}

	return name, nil
}

// NewUnitedDeploymentCondition creates a new UnitedDeployment condition.
func NewUnitedDeploymentCondition(condType unitv1alpha1.UnitedDeploymentConditionType, status corev1.ConditionStatus, reason, message string) *unitv1alpha1.UnitedDeploymentCondition {
	return &unitv1alpha1.UnitedDeploymentCondition{
		Type:               condType,
		Status:             status,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	}
}

// GetUnitedDeploymentCondition returns the condition with the provided type.
func GetUnitedDeploymentCondition(status unitv1alpha1.UnitedDeploymentStatus, condType unitv1alpha1.UnitedDeploymentConditionType) *unitv1alpha1.UnitedDeploymentCondition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == condType {
			return &c
		}
	}
	return nil
}

// SetUnitedDeploymentCondition updates the UnitedDeployment to include the provided condition. If the condition that
// we are about to add already exists and has the same status, reason and message then we are not going to update.
func SetUnitedDeploymentCondition(status *unitv1alpha1.UnitedDeploymentStatus, condition *unitv1alpha1.UnitedDeploymentCondition) {
	currentCond := GetUnitedDeploymentCondition(*status, condition.Type)
	if currentCond != nil && currentCond.Status == condition.Status && currentCond.Reason == condition.Reason {
		return
	}

	if currentCond != nil && currentCond.Status == condition.Status {
		condition.LastTransitionTime = currentCond.LastTransitionTime
	}
	newConditions := filterOutCondition(status.Conditions, condition.Type)
	status.Conditions = append(newConditions, *condition)
}

// RemoveUnitedDeploymentCondition removes the UnitedDeployment condition with the provided type.
func RemoveUnitedDeploymentCondition(status *unitv1alpha1.UnitedDeploymentStatus, condType unitv1alpha1.UnitedDeploymentConditionType) {
	status.Conditions = filterOutCondition(status.Conditions, condType)
}

func filterOutCondition(conditions []unitv1alpha1.UnitedDeploymentCondition, condType unitv1alpha1.UnitedDeploymentConditionType) []unitv1alpha1.UnitedDeploymentCondition {
	var newConditions []unitv1alpha1.UnitedDeploymentCondition
	for _, c := range conditions {
		if c.Type == condType {
			continue
		}
		newConditions = append(newConditions, c)
	}
	return newConditions
}

func GetNextPatches(ud *unitv1alpha1.UnitedDeployment) map[string]UnitedDeploymentPatches {
	next := make(map[string]UnitedDeploymentPatches)
	for _, pool := range ud.Spec.Topology.Pools {
		t := UnitedDeploymentPatches{}
		if pool.Replicas != nil {
			t.Replicas = *pool.Replicas
		}
		if pool.Patch != nil {
			t.Patch = string(pool.Patch.Raw)
		}
		next[pool.Name] = t
	}
	return next
}
