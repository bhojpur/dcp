package localhostproxy

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
)

func TestReplaceLocalHostPorts(t *testing.T) {
	testcases := map[string]struct {
		initPorts      []string
		localhostPorts string
		resultPorts    map[string]struct{}
	}{
		"no init ports for representing configmap is added": {
			localhostPorts: "10250, 10255, 10256",
			resultPorts: map[string]struct{}{
				"10250": {},
				"10255": {},
				"10256": {},
			},
		},
		"with init ports for representing configmap is updated": {
			initPorts:      []string{"10250", "10255", "10256"},
			localhostPorts: "10250, 10255, 10256, 10257",
			resultPorts: map[string]struct{}{
				"10250": {},
				"10255": {},
				"10256": {},
				"10257": {},
			},
		},
	}

	plh := &localHostProxyMiddleware{
		localhostPorts: make(map[string]struct{}),
	}

	for k, tc := range testcases {
		t.Run(k, func(t *testing.T) {
			// prepare localhost ports
			for i := range tc.initPorts {
				plh.localhostPorts[tc.initPorts[i]] = struct{}{}
			}

			// run replaceLocalHostPorts
			plh.replaceLocalHostPorts(tc.localhostPorts)

			// compare replace result
			ok := reflect.DeepEqual(plh.localhostPorts, tc.resultPorts)
			if !ok {
				t.Errorf("expect localhost ports: %v, but got %v", tc.resultPorts, plh.localhostPorts)
			}

			// cleanup localhost ports
			for port := range plh.localhostPorts {
				delete(plh.localhostPorts, port)
			}
		})
	}
}
