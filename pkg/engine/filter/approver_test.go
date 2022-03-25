package filter

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
)

func TestApprove(t *testing.T) {
	testcases := map[string]struct {
		comp           string
		resource       string
		verbs          []string
		comp2          string
		resource2      string
		verb2          string
		expectedResult bool
	}{
		"normal case": {
			"kubelet", "services", []string{"list", "watch"},
			"kubelet", "services", "list",
			true,
		},
		"components are not equal": {
			"kubelet", "services", []string{"list", "watch"},
			"kube-proxy", "services", "list",
			false,
		},
		"resources are not equal": {
			"kubelet", "services", []string{"list", "watch"},
			"kubelet", "pods", "list",
			false,
		},
		"verb is not in verbs set": {
			"kubelet", "services", []string{"list", "watch"},
			"kubelet", "services", "get",
			false,
		},
	}

	for k, tt := range testcases {
		t.Run(k, func(t *testing.T) {
			approver := NewApprover(tt.comp, tt.resource, tt.verbs...)
			result := approver.Approve(tt.comp2, tt.resource2, tt.verb2)

			if result != tt.expectedResult {
				t.Errorf("Approve error: expected %v, but got %v\n", tt.expectedResult, result)
			}
		})
	}
}
