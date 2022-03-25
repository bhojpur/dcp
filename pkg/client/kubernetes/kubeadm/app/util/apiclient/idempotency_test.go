package apiclient

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
	"testing"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
)

const configMapName = "configmap"

func TestPatchNodeNonErrorCases(t *testing.T) {
	testcases := []struct {
		name       string
		lookupName string
		node       v1.Node
		success    bool
	}{
		{
			name:       "simple update",
			lookupName: "testnode",
			node: v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "testnode",
					Labels: map[string]string{v1.LabelHostname: ""},
				},
			},
			success: true,
		},
		{
			name:       "node does not exist",
			lookupName: "whale",
			success:    false,
		},
		{
			name:       "node not labelled yet",
			lookupName: "robin",
			node: v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "robin",
				},
			},
			success: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			client := fake.NewSimpleClientset()
			_, err := client.CoreV1().Nodes().Create(context.TODO(), &tc.node, metav1.CreateOptions{})
			if err != nil {
				t.Fatalf("failed to create node to fake client: %v", err)
			}
			conditionFunction := PatchNodeOnce(client, tc.lookupName, func(node *v1.Node) {
				node.Annotations = map[string]string{
					"updatedBy": "test",
				}
			})
			success, err := conditionFunction()
			if err != nil {
				t.Fatalf("did not expect error: %v", err)
			}
			if success != tc.success {
				t.Fatalf("expected %v got %v", tc.success, success)
			}
		})
	}
}

func TestCreateOrMutateConfigMap(t *testing.T) {
	client := fake.NewSimpleClientset()
	err := CreateOrMutateConfigMap(client, &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string]string{
			"key": "some-value",
		},
	}, func(cm *v1.ConfigMap) error {
		t.Fatal("mutate should not have been called, since the ConfigMap should have been created instead of mutated")
		return nil
	})
	if err != nil {
		t.Fatalf("error creating ConfigMap: %v", err)
	}
	_, err = client.CoreV1().ConfigMaps(metav1.NamespaceSystem).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("error retrieving ConfigMap: %v", err)
	}
}

func createClientAndConfigMap(t *testing.T) *fake.Clientset {
	client := fake.NewSimpleClientset()
	_, err := client.CoreV1().ConfigMaps(metav1.NamespaceSystem).Create(context.TODO(), &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string]string{
			"key": "some-value",
		},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error creating ConfigMap: %v", err)
	}
	return client
}

func TestMutateConfigMap(t *testing.T) {
	client := createClientAndConfigMap(t)

	err := MutateConfigMap(client, metav1.ObjectMeta{
		Name:      configMapName,
		Namespace: metav1.NamespaceSystem,
	}, func(cm *v1.ConfigMap) error {
		cm.Data["key"] = "some-other-value"
		return nil
	})
	if err != nil {
		t.Fatalf("error mutating regular ConfigMap: %v", err)
	}

	cm, _ := client.CoreV1().ConfigMaps(metav1.NamespaceSystem).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if cm.Data["key"] != "some-other-value" {
		t.Fatalf("ConfigMap mutation was invalid, has: %q", cm.Data["key"])
	}
}

func TestMutateConfigMapWithConflict(t *testing.T) {
	client := createClientAndConfigMap(t)

	// Mimic that the first 5 updates of the ConfigMap returns a conflict, whereas the sixth update
	// succeeds
	conflict := 5
	client.PrependReactor("update", "configmaps", func(action core.Action) (bool, runtime.Object, error) {
		update := action.(core.UpdateAction)
		if conflict > 0 {
			conflict--
			return true, update.GetObject(), apierrors.NewConflict(action.GetResource().GroupResource(), configMapName, errors.New("conflict"))
		}
		return false, update.GetObject(), nil
	})

	err := MutateConfigMap(client, metav1.ObjectMeta{
		Name:      configMapName,
		Namespace: metav1.NamespaceSystem,
	}, func(cm *v1.ConfigMap) error {
		cm.Data["key"] = "some-other-value"
		return nil
	})
	if err != nil {
		t.Fatalf("error mutating conflicting ConfigMap: %v", err)
	}

	cm, _ := client.CoreV1().ConfigMaps(metav1.NamespaceSystem).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if cm.Data["key"] != "some-other-value" {
		t.Fatalf("ConfigMap mutation with conflict was invalid, has: %q", cm.Data["key"])
	}
}
