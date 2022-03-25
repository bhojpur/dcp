package refmanager

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
	"testing"

	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

var (
	val int32 = 2
)

func Test(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	UID := "uid"
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
			UID:       types.UID(UID),
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &val,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "foo",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "foo",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginxImage",
						},
					},
				},
			},
		},
	}

	pods := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod",
				Namespace: "default",
				Labels: map[string]string{
					"app": "foo",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "nginx",
						Image: "nginx",
					},
				},
			},
		},
	}

	var ownerRefs []metav1.OwnerReference
	scheme := runtime.NewScheme()
	scheme.AddKnownTypeWithName(schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"}, &appsv1.StatefulSet{})
	m, err := New(nil, sts.Spec.Selector, sts, scheme)
	g.Expect(err).Should(gomega.BeNil())

	mts := make([]metav1.Object, 1)
	for i, pod := range pods {
		mts[i] = &pod
	}
	ps, err := m.ClaimOwnedObjects(mts)
	g.Expect(err).Should(gomega.BeNil())
	g.Expect(len(ps)).Should(gomega.BeEquivalentTo(1))

	// remove pod label
	pod := pods[0]
	pod.Labels["app"] = "foo2"

	mts = make([]metav1.Object, 1)
	mts[0] = &pod
	ps, err = m.ClaimOwnedObjects(mts)
	g.Expect(err).Should(gomega.BeNil())
	g.Expect(len(ps)).Should(gomega.BeEquivalentTo(0))

	// remove pod label
	pod.OwnerReferences = ownerRefs

	mts = make([]metav1.Object, 1)
	mts[0] = &pod
	ps, err = m.ClaimOwnedObjects(mts)
	g.Expect(err).Should(gomega.BeNil())
	g.Expect(len(ps)).Should(gomega.BeEquivalentTo(0))
}
