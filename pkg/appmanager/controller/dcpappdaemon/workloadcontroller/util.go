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
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/validation"

	"github.com/bhojpur/dcp/pkg/appmanager/apis/apps/v1alpha1"
)

func getWorkloadPrefix(controllerName, nodepoolName string) string {
	prefix := fmt.Sprintf("%s-%s-", controllerName, nodepoolName)
	if len(validation.NameIsDNSSubdomain(prefix, true)) != 0 {
		prefix = fmt.Sprintf("%s-", controllerName)
	}
	return prefix
}

func CreateNodeSelectorByNodepoolName(nodepool string) map[string]string {
	return map[string]string{
		v1alpha1.LabelCurrentNodePool: nodepool,
	}
}

func TaintsToTolerations(taints []corev1.Taint) []corev1.Toleration {
	tolerations := []corev1.Toleration{}
	for _, taint := range taints {
		toleation := corev1.Toleration{
			Key:      taint.Key,
			Operator: corev1.TolerationOpExists,
			Effect:   taint.Effect,
		}
		tolerations = append(tolerations, toleation)
	}
	return tolerations
}
