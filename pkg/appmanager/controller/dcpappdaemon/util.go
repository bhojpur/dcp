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
	unitv1alpha1 "github.com/bhojpur/dcp/pkg/appmanager/apis/apps/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/apis/core/v1/helper"
)

const updateRetries = 5

func IsTolerationsAllTaints(tolerations []corev1.Toleration, taints []corev1.Taint) bool {
	for i, _ := range taints {
		if !helper.TolerationsTolerateTaint(tolerations, &taints[i]) {
			return false
		}
	}
	return true
}

// NewAppDaemonCondition creates a new DcpAppDaemon condition.
func NewAppDaemonCondition(condType unitv1alpha1.DcpAppDaemonConditionType, status corev1.ConditionStatus, reason, message string) *unitv1alpha1.DcpAppDaemonCondition {
	return &unitv1alpha1.DcpAppDaemonCondition{
		Type:               condType,
		Status:             status,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	}
}

// GetAppDaemonCondition returns the condition with the provided type.
func GetAppDaemonCondition(status unitv1alpha1.DcpAppDaemonStatus, condType unitv1alpha1.DcpAppDaemonConditionType) *unitv1alpha1.DcpAppDaemonCondition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == condType {
			return &c
		}
	}
	return nil
}

// SetAppDaemonCondition updates the DcpAppDaemon to include the provided condition. If the condition that
// we are about to add already exists and has the same status, reason and message then we are not going to update.
func SetAppDaemonCondition(status *unitv1alpha1.DcpAppDaemonStatus, condition *unitv1alpha1.DcpAppDaemonCondition) {
	currentCond := GetAppDaemonCondition(*status, condition.Type)
	if currentCond != nil && currentCond.Status == condition.Status && currentCond.Reason == condition.Reason {
		return
	}

	if currentCond != nil && currentCond.Status == condition.Status {
		condition.LastTransitionTime = currentCond.LastTransitionTime
	}
	newConditions := filterOutCondition(status.Conditions, condition.Type)
	status.Conditions = append(newConditions, *condition)
}

func filterOutCondition(conditions []unitv1alpha1.DcpAppDaemonCondition, condType unitv1alpha1.DcpAppDaemonConditionType) []unitv1alpha1.DcpAppDaemonCondition {
	var newConditions []unitv1alpha1.DcpAppDaemonCondition
	for _, c := range conditions {
		if c.Type == condType {
			continue
		}
		newConditions = append(newConditions, c)
	}
	return newConditions
}
