package dns

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
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsEdgeNode(t *testing.T) {
	tests := []struct {
		desc   string
		node   *corev1.Node
		expect bool
	}{
		{
			desc: "node has edge worker label which equals true",
			node: &corev1.Node{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"bhojpur.net/is-edge-worker": "true",
					},
				},
				Spec:   corev1.NodeSpec{},
				Status: corev1.NodeStatus{},
			},
			expect: true,
		},
		{
			desc: "node has edge worker label which equals false",
			node: &corev1.Node{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"bhojpur.net/is-edge-worker": "false",
					},
				},
				Spec:   corev1.NodeSpec{},
				Status: corev1.NodeStatus{},
			},
			expect: false,
		},
		{
			desc: "node dose not has edge worker label",
			node: &corev1.Node{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"yin": "ruixing",
					},
				},
				Spec:   corev1.NodeSpec{},
				Status: corev1.NodeStatus{},
			},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			act := isEdgeNode(tt.node)
			if act != tt.expect {
				t.Errorf("the result we want is: %v, but the actual result is: %v\n", tt.expect, act)
			}
		})
	}
}

func TestFormatDnsRecord(t *testing.T) {
	var (
		ip     = "10.10.102.60"
		host   = "k8s-xing-master"
		expect = ip + "\t" + host
	)

	act := formatDNSRecord(ip, host)
	if act != expect {
		t.Errorf("the result we want is: %v, but the actual result is: %v\n", expect, act)
	}

}

func TestGetNodeHostIP(t *testing.T) {
	tests := []struct {
		desc   string
		node   *corev1.Node
		expect string
	}{
		{
			desc: "get node primary host ip",
			node: &corev1.Node{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       corev1.NodeSpec{},
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{
							Type:    corev1.NodeExternalIP,
							Address: "205.20.20.2",
						},
						{
							Type:    corev1.NodeInternalIP,
							Address: "102.10.10.60",
						},
						{
							Type:    corev1.NodeHostName,
							Address: "k8s-edge-node",
						},
					},
				},
			},
			expect: "102.10.10.60",
		},
		{
			desc: "get node primary host ip",
			node: &corev1.Node{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       corev1.NodeSpec{},
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{
							Type:    corev1.NodeExternalIP,
							Address: "205.20.20.2",
						},
						{
							Type:    corev1.NodeHostName,
							Address: "k8s-edge-node",
						},
					},
				},
			},
			expect: "205.20.20.2",
		},
		{
			desc: "get node primary host ip",
			node: &corev1.Node{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       corev1.NodeSpec{},
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{
							Type:    corev1.NodeHostName,
							Address: "k8s-edge-node",
						},
					},
				},
			},
			expect: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			act, _ := getNodeHostIP(tt.node)
			if act != tt.expect {
				t.Errorf("the result we want is: %v, but the actual result is: %v\n", tt.expect, act)
			}
		})
	}
}

func TestRemoveRecordByHostname(t *testing.T) {
	var (
		records  = []string{"10.1.218.68\tk8s-xing-61", "10.10.102.60\tk8s-xing-master"}
		hostname = "k8s-xing-61"
		expect   = []string{"10.10.102.60\tk8s-xing-master"}
	)

	act, changed := removeRecordByHostname(records, hostname)

	if !changed && !reflect.DeepEqual(act, records) {
		t.Errorf("the result we want is: %v, but the actual result is: %v\n", records, act)
	} else if !reflect.DeepEqual(act, expect) {
		t.Errorf("the result we want is: %v, but the actual result is: %v\n", expect, act)
	}
}

func TestParseHostnameFromDNSRecord(t *testing.T) {
	tests := []struct {
		desc   string
		record string
		expect string
	}{
		{
			desc:   "parse host name from dns record",
			record: "10.1.218.68\tk8s-xing-61",
			expect: "k8s-xing-61",
		},

		{
			desc:   "parse invalid dns recode",
			record: "10.10.102.2invalid dns record",
			expect: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			act, _ := parseHostnameFromDNSRecord(tt.record)
			if act != tt.expect {
				t.Errorf("the result we want is: %v, but the actual result is: %v\n", tt.expect, act)
			}
		})
	}
}

func TestAddOrUpdateRecord(t *testing.T) {
	tests := []struct {
		desc    string
		records []string
		record  string
		expect  []string
	}{
		//{
		//	desc: "test add record",
		//	records: []string{"10.1.218.68\tk8s-xing-61"},
		//	record: "10.1.10.62\tk8s-xing-62",
		//	expect: []string{"10.1.218.68\tk8s-xing-61","10.1.10.62\tk8s-xing-62"},
		//},
		//
		//{
		//	desc: "test update record",
		//	records: []string{"10.1.218.68\tk8s-xing-61"},
		//	record: "10.1.10.62\tk8s-xing-61",
		//	expect: []string{"10.1.10.62\tk8s-xing-61"},
		//},

		{
			desc:    "test idempotence",
			records: []string{"10.1.218.68\tk8s-xing-61"},
			record:  "10.1.10.62\tk8s-xing-61",
			expect:  []string{"10.1.218.68\tk8s-xing-61"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			act, _, err := addOrUpdateRecord(tt.records, tt.record)
			if err != nil {
				t.Error(err)
			}
			if stringSliceEqual(act, tt.expect) {
				t.Errorf("the result we want is: %v, but the actual result is: %v\n", tt.expect, act)
			}
		})
	}
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	if (a == nil) != (b == nil) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}
