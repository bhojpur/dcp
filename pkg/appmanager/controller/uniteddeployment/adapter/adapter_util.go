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
	"encoding/json"
	"fmt"

	appsv1alpha1 "github.com/bhojpur/dcp/pkg/appmanager/apis/apps/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/klog"
)

func getPoolPrefix(controllerName, poolName string) string {
	prefix := fmt.Sprintf("%s-%s-", controllerName, poolName)
	if len(validation.NameIsDNSSubdomain(prefix, true)) != 0 {
		prefix = fmt.Sprintf("%s-", controllerName)
	}
	return prefix
}

func attachNodeAffinityAndTolerations(podSpec *corev1.PodSpec, pool *appsv1alpha1.Pool) {
	attachNodeAffinity(podSpec, pool)
	attachTolerations(podSpec, pool)
}

func attachNodeAffinity(podSpec *corev1.PodSpec, pool *appsv1alpha1.Pool) {
	if podSpec.Affinity == nil {
		podSpec.Affinity = &corev1.Affinity{}
	}

	if podSpec.Affinity.NodeAffinity == nil {
		podSpec.Affinity.NodeAffinity = &corev1.NodeAffinity{}
	}

	if podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &corev1.NodeSelector{}
	}

	if podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms == nil {
		podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = []corev1.NodeSelectorTerm{}
	}

	if len(podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) == 0 {
		podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = append(podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms, corev1.NodeSelectorTerm{})
	}

	for _, matchExpression := range pool.NodeSelectorTerm.MatchExpressions {
		for i, term := range podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
			term.MatchExpressions = append(term.MatchExpressions, matchExpression)
			podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[i] = term
		}
	}
}

func attachTolerations(podSpec *corev1.PodSpec, poolConfig *appsv1alpha1.Pool) {

	if poolConfig.Tolerations == nil {
		return
	}

	if podSpec.Tolerations == nil {
		podSpec.Tolerations = []corev1.Toleration{}
	}

	for _, toleration := range poolConfig.Tolerations {
		podSpec.Tolerations = append(podSpec.Tolerations, toleration)
	}

	return
}

func getRevision(objMeta metav1.Object) string {
	if objMeta.GetLabels() == nil {
		return ""
	}
	return objMeta.GetLabels()[appsv1alpha1.ControllerRevisionHashLabelKey]
}

// getCurrentPartition calculates current partition by counting the pods not having the updated revision
func getCurrentPartition(pods []*corev1.Pod, revision string) *int32 {
	var partition int32
	for _, pod := range pods {
		if getRevision(&pod.ObjectMeta) != revision {
			partition++
		}
	}

	return &partition
}

func StrategicMergeByPatches(oldobj interface{}, patch *runtime.RawExtension, newPatched interface{}) error {
	patchMap := make(map[string]interface{})
	if err := json.Unmarshal(patch.Raw, &patchMap); err != nil {
		klog.Errorf("Unmarshal pool patch error %v, patch Raw %v", err, string(patch.Raw))
		return err
	}

	originalObjMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(oldobj)
	if err != nil {
		klog.Errorf("ToUnstructured error %v", err)
		return err
	}

	patchedObjMap, err := strategicpatch.StrategicMergeMapPatch(originalObjMap, patchMap, newPatched)
	if err != nil {
		klog.Errorf("StartegicMergeMapPatch error %v", err)
		return err
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(patchedObjMap, newPatched); err != nil {
		klog.Errorf("FromUnstructured error %v", err)
		return err
	}
	return nil
}

func PoolHasPatch(poolConfig *appsv1alpha1.Pool, set metav1.Object) bool {
	if poolConfig.Patch == nil {
		// If No Patches, Must Set patches annotation to ""
		if anno := set.GetAnnotations(); anno != nil {
			anno[appsv1alpha1.AnnotationPatchKey] = ""
		}
		return false
	}
	return true
}

func CreateNewPatchedObject(patchInfo *runtime.RawExtension, set metav1.Object, newPatched metav1.Object) error {

	if err := StrategicMergeByPatches(set, patchInfo, newPatched); err != nil {
		return err
	}

	if anno := newPatched.GetAnnotations(); anno == nil {
		newPatched.SetAnnotations(map[string]string{
			appsv1alpha1.AnnotationPatchKey: string(patchInfo.Raw),
		})
	} else {
		anno[appsv1alpha1.AnnotationPatchKey] = string(patchInfo.Raw)
	}
	return nil
}
