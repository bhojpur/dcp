package validating

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
	"strconv"
	"strings"
	"testing"

	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	unitv1alpha1 "github.com/bhojpur/dcp/pkg/appmanager/apis/apps/v1alpha1"
)

func TestValidateUnitedDeployment(t *testing.T) {
	validLabels := map[string]string{"a": "b"}
	validPodTemplate := v1.PodTemplate{
		WorkloadTemplate: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: validLabels,
			},
			Spec: v1.PodSpec{
				RestartPolicy: v1.RestartPolicyAlways,
				DNSPolicy:     v1.DNSClusterFirst,
				Containers:    []v1.Container{{Name: "abc", Image: "image", ImagePullPolicy: "IfNotPresent"}},
			},
		},
	}

	var val int32 = 10
	replicas1 := intstr.FromInt(1)
	replicas2 := intstr.FromString("90%")
	replicas3 := intstr.FromString("71%")
	replicas4 := intstr.FromString("29%")
	successCases := []appsv1alpha1.UnitedDeployment{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
			Spec: appsv1alpha1.UnitedDeploymentSpec{
				Replicas: &val,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
					StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: validLabels,
						},
						Spec: apps.StatefulSetSpec{
							WorkloadTemplate: validPodTemplate.WorkloadTemplate,
						},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
			Spec: appsv1alpha1.UnitedDeploymentSpec{
				Replicas: &val,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
					StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: validLabels,
						},
						Spec: apps.StatefulSetSpec{
							WorkloadTemplate: validPodTemplate.WorkloadTemplate,
						},
					},
				},
				Topology: appsv1alpha1.Topology{
					Pools: []appsv1alpha1.Pool{
						{
							Name: "pool",
						},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
			Spec: appsv1alpha1.UnitedDeploymentSpec{
				Replicas: &val,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
					StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: validLabels,
						},
						Spec: apps.StatefulSetSpec{
							WorkloadTemplate: validPodTemplate.WorkloadTemplate,
						},
					},
				},
				Topology: appsv1alpha1.Topology{
					Pools: []appsv1alpha1.Pool{
						{
							Name:     "pool1",
							Replicas: &replicas1,
						},
						{
							Name: "pool2",
						},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
			Spec: appsv1alpha1.UnitedDeploymentSpec{
				Replicas: &val,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
					StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: validLabels,
						},
						Spec: apps.StatefulSetSpec{
							WorkloadTemplate: validPodTemplate.WorkloadTemplate,
						},
					},
				},
				Topology: appsv1alpha1.Topology{
					Pools: []appsv1alpha1.Pool{
						{
							Name:     "pool1",
							Replicas: &replicas1,
						},
						{
							Name:     "pool2",
							Replicas: &replicas2,
						},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
			Spec: appsv1alpha1.UnitedDeploymentSpec{
				Replicas: &val,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
					StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: validLabels,
						},
						Spec: apps.StatefulSetSpec{
							WorkloadTemplate: validPodTemplate.WorkloadTemplate,
						},
					},
				},
				Topology: appsv1alpha1.Topology{
					Pools: []appsv1alpha1.Pool{
						{
							Name:     "pool1",
							Replicas: &replicas3,
						},
						{
							Name:     "pool2",
							Replicas: &replicas4,
						},
					},
				},
			},
		},
	}

	for i, successCase := range successCases {
		t.Run("success case "+strconv.Itoa(i), func(t *testing.T) {
			setTestDefault(&successCase)
			if errs := validateUnitedDeployment(&successCase); len(errs) != 0 {
				t.Errorf("expected success: %v", errs)
			}
		})
	}

	errorCases := map[string]appsv1alpha1.UnitedDeployment{
		"no pod template label": {
			ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
			Spec: appsv1alpha1.UnitedDeploymentSpec{
				Replicas: &val,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
					StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
						Spec: apps.StatefulSetSpec{
							WorkloadTemplate: v1.PodTemplateSpec{
								ObjectMeta: metav1.ObjectMeta{},
								Spec: v1.PodSpec{
									RestartPolicy: v1.RestartPolicyAlways,
									DNSPolicy:     v1.DNSClusterFirst,
									Containers:    []v1.Container{{Name: "abc", Image: "image", ImagePullPolicy: "IfNotPresent"}},
								},
							},
						},
					},
				},
			},
		},
		"no pool template": {
			ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
			Spec: appsv1alpha1.UnitedDeploymentSpec{
				Replicas: &val,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				WorkloadTemplate: appsv1alpha1.WorkloadTemplate{},
			},
		},
		"no pool name": {
			ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
			Spec: appsv1alpha1.UnitedDeploymentSpec{
				Replicas: &val,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
					StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: validLabels,
						},
						Spec: apps.StatefulSetSpec{
							WorkloadTemplate: validPodTemplate.WorkloadTemplate,
						},
					},
				},
				Topology: appsv1alpha1.Topology{
					Pools: []appsv1alpha1.Pool{
						{},
					},
				},
			},
		},
		"invalid pool nodeSelectorTerm": {
			ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
			Spec: appsv1alpha1.UnitedDeploymentSpec{
				Replicas: &val,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
					StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: validLabels,
						},
						Spec: apps.StatefulSetSpec{
							WorkloadTemplate: validPodTemplate.WorkloadTemplate,
						},
					},
				},
				Topology: appsv1alpha1.Topology{
					Pools: []appsv1alpha1.Pool{
						{
							Name: "pool",
							NodeSelectorTerm: corev1.NodeSelectorTerm{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{
										Key:      "key",
										Operator: corev1.NodeSelectorOpExists,
										Values:   []string{"unexpected"},
									},
								},
							},
						},
					},
				},
			},
		},
		"pool replicas is not enough": {
			ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
			Spec: appsv1alpha1.UnitedDeploymentSpec{
				Replicas: &val,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
					StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: validLabels,
						},
						Spec: apps.StatefulSetSpec{
							WorkloadTemplate: validPodTemplate.WorkloadTemplate,
						},
					},
				},
				Topology: appsv1alpha1.Topology{
					Pools: []appsv1alpha1.Pool{
						{
							Name:     "pool1",
							Replicas: &replicas1,
						},
					},
				},
			},
		},
		"pool replicas is too small": {
			ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
			Spec: appsv1alpha1.UnitedDeploymentSpec{
				Replicas: &val,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
					StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: validLabels,
						},
						Spec: apps.StatefulSetSpec{
							WorkloadTemplate: validPodTemplate.WorkloadTemplate,
						},
					},
				},
				Topology: appsv1alpha1.Topology{
					Pools: []appsv1alpha1.Pool{
						{
							Name:     "pool1",
							Replicas: &replicas1,
						},
						{
							Name:     "pool2",
							Replicas: &replicas3,
						},
					},
				},
			},
		},
		"pool replicas is too much": {
			ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
			Spec: appsv1alpha1.UnitedDeploymentSpec{
				Replicas: &val,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
					StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: validLabels,
						},
						Spec: apps.StatefulSetSpec{
							WorkloadTemplate: validPodTemplate.WorkloadTemplate,
						},
					},
				},
				Topology: appsv1alpha1.Topology{
					Pools: []appsv1alpha1.Pool{
						{
							Name:     "pool1",
							Replicas: &replicas3,
						},
						{
							Name:     "pool2",
							Replicas: &replicas2,
						},
					},
				},
			},
		},
		"partition not exist": {
			ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
			Spec: appsv1alpha1.UnitedDeploymentSpec{
				Replicas: &val,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
					StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: validLabels,
						},
						Spec: apps.StatefulSetSpec{
							WorkloadTemplate: validPodTemplate.WorkloadTemplate,
						},
					},
				},
				UpdateStrategy: appsv1alpha1.UnitedDeploymentUpdateStrategy{
					StatefulSetUpdateStrategy: &appsv1alpha1.StatefulSetUpdateStrategy{
						Partitions: map[string]int32{
							"notExist": 1,
						},
					},
				},
				Topology: appsv1alpha1.Topology{
					Pools: []appsv1alpha1.Pool{
						{
							Name:     "pool1",
							Replicas: &replicas3,
						},
						{
							Name:     "pool2",
							Replicas: &replicas2,
						},
					},
				},
			},
		},
		"duplicated templates": {
			ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
			Spec: appsv1alpha1.UnitedDeploymentSpec{
				Replicas: &val,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
					StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: validLabels,
						},
						Spec: apps.StatefulSetSpec{
							WorkloadTemplate: validPodTemplate.WorkloadTemplate,
						},
					},
					DeploymentTemplate: &appsv1alpha1.DeploymentTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: validLabels,
						},
						Spec: apps.DeploymentSpec{
							WorkloadTemplate: validPodTemplate.WorkloadTemplate,
						},
					},
				},
			},
		},
	}

	for k, v := range errorCases {
		t.Run(k, func(t *testing.T) {
			setTestDefault(&v)
			errs := validateUnitedDeployment(&v)
			if len(errs) == 0 {
				t.Errorf("expected failure for %s", k)
			}

			for i := range errs {
				field := errs[i].Field
				if !strings.HasPrefix(field, "spec.template") &&
					field != "spec.selector" &&
					field != "spec.topology.pools" &&
					field != "spec.topology.pools[0]" &&
					field != "spec.topology.pools[0].name" &&
					field != "spec.updateStrategy.partitions" &&
					field != "spec.topology.pools[0].nodeSelectorTerm.matchExpressions[0].values" {
					t.Errorf("%s: missing prefix for: %v", k, errs[i])
				}
			}
		})
	}
}

type UpdateCase struct {
	Old appsv1alpha1.UnitedDeployment
	New appsv1alpha1.UnitedDeployment
}

func TestValidateUnitedDeploymentUpdate(t *testing.T) {
	validLabels := map[string]string{"a": "b"}
	validPodTemplate := v1.PodTemplate{
		WorkloadTemplate: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: validLabels,
			},
			Spec: v1.PodSpec{
				RestartPolicy: v1.RestartPolicyAlways,
				DNSPolicy:     v1.DNSClusterFirst,
				Containers:    []v1.Container{{Name: "abc", Image: "image", ImagePullPolicy: "IfNotPresent"}},
			},
		},
	}

	var val int32 = 10
	successCases := []UpdateCase{
		{
			Old: appsv1alpha1.UnitedDeployment{
				ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault, ResourceVersion: "1"},
				Spec: appsv1alpha1.UnitedDeploymentSpec{
					Replicas: &val,
					Selector: &metav1.LabelSelector{MatchLabels: validLabels},
					WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
						StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: validLabels,
							},
							Spec: apps.StatefulSetSpec{
								WorkloadTemplate: validPodTemplate.WorkloadTemplate,
							},
						},
					},
					Topology: appsv1alpha1.Topology{
						Pools: []appsv1alpha1.Pool{
							{
								Name: "pool-a",
								NodeSelectorTerm: v1.NodeSelectorTerm{
									MatchExpressions: []v1.NodeSelectorRequirement{
										{
											Key:      "domain",
											Operator: v1.NodeSelectorOpIn,
											Values:   []string{"a"},
										},
									},
								},
							},
						},
					},
				},
			},
			New: appsv1alpha1.UnitedDeployment{
				ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault, ResourceVersion: "1"},
				Spec: appsv1alpha1.UnitedDeploymentSpec{
					Replicas: &val,
					Selector: &metav1.LabelSelector{MatchLabels: validLabels},
					WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
						StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: validLabels,
							},
							Spec: apps.StatefulSetSpec{
								WorkloadTemplate: validPodTemplate.WorkloadTemplate,
							},
						},
					},
					Topology: appsv1alpha1.Topology{
						Pools: []appsv1alpha1.Pool{
							{
								Name: "pool-a",
								NodeSelectorTerm: v1.NodeSelectorTerm{
									MatchExpressions: []v1.NodeSelectorRequirement{
										{
											Key:      "domain",
											Operator: v1.NodeSelectorOpIn,
											Values:   []string{"a"},
										},
									},
								},
							},
							{
								Name: "pool-b",
								NodeSelectorTerm: v1.NodeSelectorTerm{
									MatchExpressions: []v1.NodeSelectorRequirement{
										{
											Key:      "domain",
											Operator: v1.NodeSelectorOpIn,
											Values:   []string{"b"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Old: appsv1alpha1.UnitedDeployment{
				ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault, ResourceVersion: "1"},
				Spec: appsv1alpha1.UnitedDeploymentSpec{
					Replicas: &val,
					Selector: &metav1.LabelSelector{MatchLabels: validLabels},
					WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
						StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: validLabels,
							},
							Spec: apps.StatefulSetSpec{
								WorkloadTemplate: validPodTemplate.WorkloadTemplate,
							},
						},
					},
					Topology: appsv1alpha1.Topology{
						Pools: []appsv1alpha1.Pool{
							{
								Name: "pool-a",
								NodeSelectorTerm: v1.NodeSelectorTerm{
									MatchExpressions: []v1.NodeSelectorRequirement{
										{
											Key:      "domain",
											Operator: v1.NodeSelectorOpIn,
											Values:   []string{"a"},
										},
									},
								},
							},
							{
								Name: "pool-b",
								NodeSelectorTerm: v1.NodeSelectorTerm{
									MatchExpressions: []v1.NodeSelectorRequirement{
										{
											Key:      "domain",
											Operator: v1.NodeSelectorOpIn,
											Values:   []string{"b"},
										},
									},
								},
							},
						},
					},
				},
			},
			New: appsv1alpha1.UnitedDeployment{
				ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault, ResourceVersion: "1"},
				Spec: appsv1alpha1.UnitedDeploymentSpec{
					Replicas: &val,
					Selector: &metav1.LabelSelector{MatchLabels: validLabels},
					WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
						StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: validLabels,
							},
							Spec: apps.StatefulSetSpec{
								WorkloadTemplate: validPodTemplate.WorkloadTemplate,
							},
						},
					},
					Topology: appsv1alpha1.Topology{
						Pools: []appsv1alpha1.Pool{
							{
								Name: "pool-a",
								NodeSelectorTerm: v1.NodeSelectorTerm{
									MatchExpressions: []v1.NodeSelectorRequirement{
										{
											Key:      "domain",
											Operator: v1.NodeSelectorOpIn,
											Values:   []string{"a"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for i, successCase := range successCases {
		t.Run("success case "+strconv.Itoa(i), func(t *testing.T) {
			setTestDefault(&successCase.Old)
			setTestDefault(&successCase.New)
			if errs := ValidateUnitedDeploymentUpdate(&successCase.Old, &successCase.New); len(errs) != 0 {
				t.Errorf("expected success: %v", errs)
			}
		})
	}

	errorCases := map[string]UpdateCase{
		"pool nodeSelector changed": {
			Old: appsv1alpha1.UnitedDeployment{
				ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault, ResourceVersion: "1"},
				Spec: appsv1alpha1.UnitedDeploymentSpec{
					Replicas: &val,
					Selector: &metav1.LabelSelector{MatchLabels: validLabels},
					WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
						StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: validLabels,
							},
							Spec: apps.StatefulSetSpec{
								WorkloadTemplate: validPodTemplate.WorkloadTemplate,
							},
						},
					},
					Topology: appsv1alpha1.Topology{
						Pools: []appsv1alpha1.Pool{
							{
								Name: "pool-a",
								NodeSelectorTerm: v1.NodeSelectorTerm{
									MatchExpressions: []v1.NodeSelectorRequirement{
										{
											Key:      "domain",
											Operator: v1.NodeSelectorOpIn,
											Values:   []string{"a", "b"},
										},
									},
								},
							},
						},
					},
				},
			},
			New: appsv1alpha1.UnitedDeployment{
				ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault, ResourceVersion: "1"},
				Spec: appsv1alpha1.UnitedDeploymentSpec{
					Replicas: &val,
					Selector: &metav1.LabelSelector{MatchLabels: validLabels},
					WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
						StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: validLabels,
							},
							Spec: apps.StatefulSetSpec{
								WorkloadTemplate: validPodTemplate.WorkloadTemplate,
							},
						},
					},
					Topology: appsv1alpha1.Topology{
						Pools: []appsv1alpha1.Pool{
							{
								Name: "pool-a",
								NodeSelectorTerm: v1.NodeSelectorTerm{
									MatchExpressions: []v1.NodeSelectorRequirement{
										{
											Key:      "domain",
											Operator: v1.NodeSelectorOpIn,
											Values:   []string{"a"},
										},
									},
								},
							},
							{
								Name: "pool-b",
								NodeSelectorTerm: v1.NodeSelectorTerm{
									MatchExpressions: []v1.NodeSelectorRequirement{
										{
											Key:      "domain",
											Operator: v1.NodeSelectorOpIn,
											Values:   []string{"b"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for k, v := range errorCases {
		t.Run(k, func(t *testing.T) {
			setTestDefault(&v.Old)
			setTestDefault(&v.New)
			errs := ValidateUnitedDeploymentUpdate(&v.Old, &v.New)
			if len(errs) == 0 {
				t.Errorf("expected failure for %s", k)
			}

			for i := range errs {
				field := errs[i].Field
				if !strings.HasPrefix(field, "spec.template.") &&
					field != "spec.selector" &&
					field != "spec.topology.pool" &&
					field != "spec.topology.pool.name" &&
					field != "spec.updateStrategy.partitions" &&
					field != "spec.topology.pools[0].nodeSelectorTerm" {
					t.Errorf("%s: missing prefix for: %v", k, errs[i])
				}
			}
		})
	}
}

func setTestDefault(obj *appsv1alpha1.UnitedDeployment) {
	if obj.Spec.Replicas == nil {
		obj.Spec.Replicas = new(int32)
		*obj.Spec.Replicas = 1
	}
	if obj.Spec.RevisionHistoryLimit == nil {
		obj.Spec.RevisionHistoryLimit = new(int32)
		*obj.Spec.RevisionHistoryLimit = 10
	}
}


*/
