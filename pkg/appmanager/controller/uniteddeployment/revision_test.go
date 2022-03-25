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

/*
import (
	"testing"

	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1alpha1 "github.com/bhojpur/dcp/pkg/appmanager/apis/apps/v1alpha1"
)

func TestRevisionManage(t *testing.T) {
	g, requests, stopMgr, mgrStopped := setUp(t)
	defer func() {
		clean(g, c)
		close(stopMgr)
		mgrStopped.Wait()
	}()

	instance := &appsv1alpha1.UnitedDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: appsv1alpha1.UnitedDeploymentSpec{
			Replicas: &one,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "foo",
				},
			},
			WorkloadTemplate: appsv1alpha1.PoolTemplate{
				StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"name": "foo",
						},
					},
					Spec: appsv1.StatefulSetSpec{
						WorkloadTemplate: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"name": "foo",
								},
							},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "container-a",
										Image: "nginx:1.0",
									},
								},
							},
						},
					},
				},
			},
			Topology: appsv1alpha1.Topology{
				Pools: []appsv1alpha1.Pool{
					{
						Name: "pool-a",
						NodeSelectorTerm: corev1.NodeSelectorTerm{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      "node-name",
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{"nodeA"},
								},
							},
						},
					},
				},
			},
			RevisionHistoryLimit: &two,
		},
	}

	// Create the UnitedDeployment object and expect the Reconcile and Deployment to be created
	err := c.Create(context.TODO(), instance)
	// The instance object may not be a valid object because it might be missing some required fields.
	// Please modify the instance object by adding required fields and then remove the following if statement.
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), instance)
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))

	revisionList := &appsv1.ControllerRevisionList{}
	g.Expect(c.List(context.TODO(), revisionList, &client.ListOptions{})).Should(gomega.BeNil())
	g.Expect(len(revisionList.Items)).Should(gomega.BeEquivalentTo(1))

	g.Expect(c.Get(context.TODO(), client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, instance)).Should(gomega.BeNil())
	instance.Spec.WorkloadTemplate.StatefulSetTemplate.Labels["version"] = "v2"
	g.Expect(c.Update(context.TODO(), instance)).Should(gomega.BeNil())
	waitReconcilerProcessFinished(g, requests, 0)

	revisionList = &appsv1.ControllerRevisionList{}
	g.Expect(c.List(context.TODO(), revisionList, &client.ListOptions{})).Should(gomega.BeNil())
	g.Expect(len(revisionList.Items)).Should(gomega.BeEquivalentTo(2))

	g.Expect(c.Get(context.TODO(), client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, instance)).Should(gomega.BeNil())
	instance.Spec.WorkloadTemplate.StatefulSetTemplate.Spec.WorkloadTemplate.Spec.Containers[0].Image = "nginx:1.1"
	g.Expect(c.Update(context.TODO(), instance)).Should(gomega.BeNil())
	waitReconcilerProcessFinished(g, requests, 0)

	revisionList = &appsv1.ControllerRevisionList{}
	g.Expect(c.List(context.TODO(), revisionList, &client.ListOptions{})).Should(gomega.BeNil())
	g.Expect(len(revisionList.Items)).Should(gomega.BeEquivalentTo(2))
}

*/
