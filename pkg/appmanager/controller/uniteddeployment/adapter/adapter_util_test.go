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
	"testing"

	appsv1 "k8s.io/api/apps/v1"

	"k8s.io/apimachinery/pkg/runtime"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	unitv1alpha1 "github.com/bhojpur/dcp/pkg/appmanager/apis/apps/v1alpha1"
)

func TestGetCurrentPartitionForStrategyOnDelete(t *testing.T) {
	currentPods := buildPodList([]int{0, 1, 2}, []string{"v1", "v2", "v2"}, t)
	if partition := getCurrentPartition(currentPods, "v2"); *partition != 1 {
		t.Fatalf("expected partition 1, got %d", *partition)
	}

	currentPods = buildPodList([]int{0, 1, 2}, []string{"v1", "v1", "v2"}, t)
	if partition := getCurrentPartition(currentPods, "v2"); *partition != 2 {
		t.Fatalf("expected partition 2, got %d", *partition)
	}

	currentPods = buildPodList([]int{0, 1, 2, 3}, []string{"v2", "v1", "v2", "v2"}, t)
	if partition := getCurrentPartition(currentPods, "v2"); *partition != 1 {
		t.Fatalf("expected partition 1, got %d", *partition)
	}

	currentPods = buildPodList([]int{1, 2, 3}, []string{"v1", "v2", "v2"}, t)
	if partition := getCurrentPartition(currentPods, "v2"); *partition != 1 {
		t.Fatalf("expected partition 1, got %d", *partition)
	}

	currentPods = buildPodList([]int{0, 1, 3}, []string{"v2", "v1", "v2"}, t)
	if partition := getCurrentPartition(currentPods, "v2"); *partition != 1 {
		t.Fatalf("expected partition 1, got %d", *partition)
	}

	currentPods = buildPodList([]int{0, 1, 2}, []string{"v1", "v1", "v1"}, t)
	if partition := getCurrentPartition(currentPods, "v2"); *partition != 3 {
		t.Fatalf("expected partition 3, got %d", *partition)
	}

	currentPods = buildPodList([]int{0, 1, 2, 4}, []string{"v1", "", "v2", "v3"}, t)
	if partition := getCurrentPartition(currentPods, "v2"); *partition != 3 {
		t.Fatalf("expected partition 3, got %d", *partition)
	}
}

func buildPodList(ordinals []int, revisions []string, t *testing.T) []*corev1.Pod {
	if len(ordinals) != len(revisions) {
		t.Fatalf("ordinals count should equals to revision count")
	}
	pods := []*corev1.Pod{}
	for i, ordinal := range ordinals {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      fmt.Sprintf("pod-%d", ordinal),
			},
		}
		if revisions[i] != "" {
			pod.Labels = map[string]string{
				unitv1alpha1.ControllerRevisionHashLabelKey: revisions[i],
			}
		}
		pods = append(pods, pod)
	}

	return pods
}

func TestCreateNewPatchedObject(t *testing.T) {
	cases := []struct {
		Name         string
		PatchInfo    *runtime.RawExtension
		OldObj       *appsv1.Deployment
		EqualFuntion func(new *appsv1.Deployment) bool
	}{
		{
			Name:      "replace image",
			PatchInfo: &runtime.RawExtension{Raw: []byte(`{"spec":{"template":{"spec":{"containers":[{"image":"nginx:1.18.0","name":"nginx"}]}}}}`)},
			OldObj: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "nginx",
									Image: "nginx:1.19.0",
								},
							},
						},
					},
				},
			},
			EqualFuntion: func(new *appsv1.Deployment) bool {
				return new.Spec.Template.Spec.Containers[0].Image == "nginx:1.18.0"
			},
		},
		{
			Name:      "add other image",
			PatchInfo: &runtime.RawExtension{Raw: []byte(`{"spec":{"template":{"spec":{"containers":[{"image":"nginx:1.18.0","name":"nginx111"}]}}}}`)},
			OldObj: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "nginx",
									Image: "nginx:1.19.0",
								},
							},
						},
					},
				},
			},
			EqualFuntion: func(new *appsv1.Deployment) bool {
				if len(new.Spec.Template.Spec.Containers) != 2 {
					return false
				}
				containerMap := make(map[string]string)
				for _, container := range new.Spec.Template.Spec.Containers {
					containerMap[container.Name] = container.Image
				}
				image, ok := containerMap["nginx"]
				if !ok {
					return false
				}

				image1, ok := containerMap["nginx111"]
				if !ok {
					return false
				}
				return image == "nginx:1.19.0" && image1 == "nginx:1.18.0"
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			newObj := &appsv1.Deployment{}
			if err := CreateNewPatchedObject(c.PatchInfo, c.OldObj, newObj); err != nil {
				t.Fatalf("%s CreateNewPatchedObject error %v", c.Name, err)
			}
			if !c.EqualFuntion(newObj) {
				t.Fatalf("%s Not Expect equal funtion", c.Name)
			}
		})
	}

}
