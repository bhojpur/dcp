package nodepool

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
	"reflect"
	"sort"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	failed  = "\u2717"
	succeed = "\u2713"
)

func TestConciliateLables(t *testing.T) {
	tests := []struct {
		name      string
		node      *corev1.Node
		oldLabels map[string]string
		newLabels map[string]string
		expect    map[string]string
	}{
		{
			"remove lable",
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					Labels: map[string]string{
						"foo": "bar",
						"buz": "qux",
					},
				},
			},
			map[string]string{
				"foo": "bar",
				"buz": "qux",
			},
			map[string]string{"foo": "bar"},
			map[string]string{"foo": "bar"},
		},
		{
			"add labels",
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test-node",
					Labels: map[string]string{"foo": "bar"},
				},
			},
			map[string]string{"foo": "bar"},
			map[string]string{
				"foo": "bar",
				"buz": "qux",
			},
			map[string]string{
				"foo": "bar",
				"buz": "qux",
			},
		},
		{
			"remove and add labels",
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					Labels: map[string]string{
						"foo": "bar",
						"buz": "qux",
					},
				},
			},
			map[string]string{
				"foo": "bar",
				"buz": "qux",
			},
			map[string]string{
				"foo":  "bar",
				"quux": "quuz",
			},
			map[string]string{
				"foo":  "bar",
				"quux": "quuz",
			},
		},
		{
			"with existing node labels",
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					Labels: map[string]string{
						"foo":    "bar",
						"buz":    "qux",
						"grault": "corge",
					},
				},
			},
			map[string]string{
				"foo": "bar",
				"buz": "qux",
			},
			map[string]string{
				"foo":  "bar",
				"quux": "quuz",
			},
			map[string]string{
				"foo":    "bar",
				"quux":   "quuz",
				"grault": "corge",
			},
		},
	}
	for _, tt := range tests {
		st := tt
		tf := func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", st.name)
			{
				conciliateLabels(st.node, st.oldLabels, st.newLabels)
				get := st.node.Labels
				if !reflect.DeepEqual(get, st.expect) {
					t.Fatalf("\t%s\texpect %v, but get %v", failed, st.expect, get)
				}
				t.Logf("\t%s\texpect %v, get %v", succeed, st.expect, get)

			}
		}
		t.Run(st.name, tf)
	}
}

func TestConciliateAnnotations(t *testing.T) {
	tests := []struct {
		name     string
		node     *corev1.Node
		oldAnnos map[string]string
		newAnnos map[string]string
		expect   map[string]string
	}{
		{
			"remove an annotation",
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					Annotations: map[string]string{
						"foo": "bar",
						"buz": "qux",
					},
				},
			},
			map[string]string{
				"foo": "bar",
				"buz": "qux",
			},
			map[string]string{"foo": "bar"},
			map[string]string{"foo": "bar"},
		},
		{
			"add an annotation",
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-node",
					Annotations: map[string]string{"foo": "bar"},
				},
			},
			map[string]string{"foo": "bar"},
			map[string]string{
				"foo": "bar",
				"buz": "qux",
			},
			map[string]string{
				"foo": "bar",
				"buz": "qux",
			},
		},
		{
			"remove and add an annotation",
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					Annotations: map[string]string{
						"foo": "bar",
						"buz": "qux",
					},
				},
			},
			map[string]string{
				"foo": "bar",
				"buz": "qux",
			},
			map[string]string{
				"foo":  "bar",
				"quux": "quuz",
			},
			map[string]string{
				"foo":  "bar",
				"quux": "quuz",
			},
		},
		{
			"with existing node annotations",
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					Annotations: map[string]string{
						"foo":    "bar",
						"buz":    "qux",
						"grault": "corge",
					},
				},
			},
			map[string]string{
				"foo": "bar",
				"buz": "qux",
			},
			map[string]string{
				"foo":  "bar",
				"quux": "quuz",
			},
			map[string]string{
				"foo":    "bar",
				"quux":   "quuz",
				"grault": "corge",
			},
		},
	}
	for _, tt := range tests {
		st := tt
		tf := func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", st.name)
			{
				conciliateAnnotations(st.node, st.oldAnnos, st.newAnnos)
				get := st.node.Annotations
				if !reflect.DeepEqual(get, st.expect) {
					t.Fatalf("\t%s\texpect %v, but get %v", failed, st.expect, get)
				}
				t.Logf("\t%s\texpect %v, get %v", succeed, st.expect, get)

			}
		}
		t.Run(st.name, tf)
	}
}

func TestConciliateTaints(t *testing.T) {
	tests := []struct {
		name      string
		node      corev1.Node
		preTaints []corev1.Taint
		newTaints []corev1.Taint
		expect    []corev1.Taint
	}{
		{
			"remove the taint",
			corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "test-node"},
				Spec: corev1.NodeSpec{
					Taints: []corev1.Taint{
						{
							Key:    "foo",
							Effect: corev1.TaintEffectNoSchedule,
						},
					},
				},
			},
			[]corev1.Taint{
				{
					Key:    "foo",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			[]corev1.Taint{},
			[]corev1.Taint{},
		},
		{
			"add a new taint",
			corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "test-node"},
				Spec: corev1.NodeSpec{
					Taints: []corev1.Taint{
						{
							Key:    "foo",
							Effect: corev1.TaintEffectNoSchedule,
						},
					},
				},
			},
			[]corev1.Taint{
				{
					Key:    "foo",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			[]corev1.Taint{
				{
					Key:    "foo",
					Effect: corev1.TaintEffectNoSchedule,
				},
				{
					Key:    "bar",
					Effect: corev1.TaintEffectNoExecute,
				},
			},
			[]corev1.Taint{
				{
					Key:    "foo",
					Effect: corev1.TaintEffectNoSchedule,
				},
				{
					Key:    "bar",
					Effect: corev1.TaintEffectNoExecute,
				},
			},
		},
		{
			"update a taint",
			corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "test-node"},
				Spec: corev1.NodeSpec{
					Taints: []corev1.Taint{
						{
							Key:    "foo",
							Effect: corev1.TaintEffectNoSchedule,
						},
					},
				},
			},
			[]corev1.Taint{
				{
					Key:    "foo",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			[]corev1.Taint{
				{
					Key:    "foo",
					Value:  "bar",
					Effect: corev1.TaintEffectNoExecute,
				},
			},
			[]corev1.Taint{
				{
					Key:    "foo",
					Value:  "bar",
					Effect: corev1.TaintEffectNoExecute,
				},
			},
		},
		{
			"with existing node taints",
			corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "test-node"},
				Spec: corev1.NodeSpec{
					Taints: []corev1.Taint{
						{
							Key:    "foo",
							Effect: corev1.TaintEffectNoSchedule,
						},
						{
							Key:    "qux",
							Effect: corev1.TaintEffectNoSchedule,
						},
					},
				},
			},
			[]corev1.Taint{
				{
					Key:    "foo",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			[]corev1.Taint{
				{
					Key:    "foo",
					Value:  "bar",
					Effect: corev1.TaintEffectNoExecute,
				},
			},
			[]corev1.Taint{
				{
					Key:    "foo",
					Value:  "bar",
					Effect: corev1.TaintEffectNoExecute,
				},
				{
					Key:    "qux",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
		},
	}
	for _, tt := range tests {
		st := tt
		tf := func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", st.name)
			{
				conciliateTaints(&st.node, st.preTaints, st.newTaints)
				get := st.node.Spec.Taints
				if !taintSliceEqual(get, st.expect) {
					t.Fatalf("\t%s\texpect %v, but get %v", failed, st.expect, get)
				}
				t.Logf("\t%s\texpect %v, get %v", succeed, st.expect, get)

			}
		}
		t.Run(st.name, tf)
	}
}

func taintSliceEqual(s1, s2 []corev1.Taint) bool {
	sort.Slice(s1, func(i, j int) bool {
		if s1[i].Key < s1[j].Key {
			return true
		}
		if s1[i].Key > s1[j].Key {
			return false
		}
		// if s1[i].Key == s1[j].Key, compare the Effect
		return s1[i].Effect < s1[j].Effect
	})
	sort.Slice(s2, func(i, j int) bool {
		if s2[i].Key < s2[j].Key {
			return true
		}
		if s2[i].Key > s2[j].Key {
			return false
		}
		return s2[i].Effect < s2[j].Effect
	})
	return reflect.DeepEqual(s1, s2)
}

func TestContainTaint(t *testing.T) {
	tmpTime := metav1.Now()
	tests := []struct {
		name   string
		taint  corev1.Taint
		taints []corev1.Taint
		expect bool
	}{
		{
			"containt the taint",
			corev1.Taint{
				Key:    "foo",
				Effect: corev1.TaintEffectNoSchedule,
			},
			[]corev1.Taint{
				{
					Key:       "foo",
					Value:     "bar",
					Effect:    corev1.TaintEffectNoSchedule,
					TimeAdded: &tmpTime,
				},
			},
			true,
		},
		{
			"not containt the taint",
			corev1.Taint{
				Key:    "foo",
				Effect: corev1.TaintEffectNoSchedule,
			},
			[]corev1.Taint{
				{
					Key:       "foo",
					Value:     "bar",
					Effect:    corev1.TaintEffectNoExecute,
					TimeAdded: &tmpTime,
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		st := tt
		tf := func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", st.name)
			{
				_, get := containTaint(st.taint, st.taints)
				if get != st.expect {
					t.Fatalf("\t%s\texpect %v, but get %v", failed, st.expect, get)
				}
				t.Logf("\t%s\texpect %v, get %v", succeed, st.expect, get)

			}
		}
		t.Run(st.name, tf)
	}
}

func TestIsNodeReady(t *testing.T) {
	tests := []struct {
		name   string
		node   corev1.Node
		expect bool
	}{
		{
			"node is ready",
			corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "test-node"},
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{
							Type:   corev1.NodeReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			true,
		},
		{
			"node is not ready",
			corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "test-node"},
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		st := tt
		tf := func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", st.name)
			{
				get := isNodeReady(st.node)
				if get != st.expect {
					t.Fatalf("\t%s\texpect %v, but get %v", failed, st.expect, get)
				}
				t.Logf("\t%s\texpect %v, get %v", succeed, st.expect, get)

			}
		}
		t.Run(st.name, tf)
	}
}

func TestMergeMap(t *testing.T) {
	tests := []struct {
		name   string
		map1   map[string]string
		map2   map[string]string
		expect map[string]string
	}{
		{
			"add new key/val",
			map[string]string{"foo": "bar"},
			map[string]string{"buz": "qux"},
			map[string]string{
				"foo": "bar",
				"buz": "qux",
			},
		},
		{
			"replace exist key/val",
			map[string]string{
				"foo": "bar",
				"buz": "qux",
			},
			map[string]string{"buz": "quux"},
			map[string]string{
				"foo": "bar",
				"buz": "quux",
			},
		},
	}
	for _, tt := range tests {
		st := tt
		tf := func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", st.name)
			{
				get := mergeMap(st.map1, st.map2)
				if !reflect.DeepEqual(get, st.expect) {
					t.Fatalf("\t%s\texpect %v, but get %v", failed, st.expect, get)
				}
				t.Logf("\t%s\texpect %v, get %v", succeed, st.expect, get)

			}
		}
		t.Run(st.name, tf)
	}
}

func TestRemoveTaint(t *testing.T) {
	tmpTime := metav1.Now()
	tests := []struct {
		name   string
		taint  corev1.Taint
		taints []corev1.Taint
		expect []corev1.Taint
	}{
		{
			"remove the taint",
			corev1.Taint{
				Key:    "foo",
				Effect: corev1.TaintEffectNoSchedule,
			},
			[]corev1.Taint{
				{
					Key:       "foo",
					Value:     "bar",
					Effect:    corev1.TaintEffectNoSchedule,
					TimeAdded: &tmpTime,
				},
			},
			[]corev1.Taint{},
		},
		{
			"not containt the taint",
			corev1.Taint{
				Key:    "foo",
				Effect: corev1.TaintEffectNoSchedule,
			},
			[]corev1.Taint{
				{
					Key:       "foo",
					Value:     "bar",
					Effect:    corev1.TaintEffectNoExecute,
					TimeAdded: &tmpTime,
				},
			},
			[]corev1.Taint{
				{
					Key:       "foo",
					Value:     "bar",
					Effect:    corev1.TaintEffectNoExecute,
					TimeAdded: &tmpTime,
				},
			},
		},
	}
	for _, tt := range tests {
		st := tt
		tf := func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", st.name)
			{
				get := removeTaint(st.taint, st.taints)
				if !reflect.DeepEqual(get, st.expect) {
					t.Fatalf("\t%s\texpect %v, but get %v", failed, st.expect, get)
				}
				t.Logf("\t%s\texpect %v, get %v", succeed, st.expect, get)

			}
		}
		t.Run(st.name, tf)
	}
}
