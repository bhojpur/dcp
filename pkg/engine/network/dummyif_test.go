//go:build linux
// +build linux

package network

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
	"net"
	"testing"
)

const (
	testDummyIfName = "test-dummyif"
)

func TestEnsureDummyInterface(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	testcases := map[string]struct {
		preparedIP string
		testIP     string
		resultIP   string
	}{
		"init ensure dummy interface": {
			testIP:   "169.254.2.1",
			resultIP: "169.254.2.1",
		},
		"ensure dummy interface after prepared": {
			preparedIP: "169.254.2.2",
			testIP:     "169.254.2.2",
			resultIP:   "169.254.2.2",
		},
		"ensure dummy interface with new ip": {
			preparedIP: "169.254.2.3",
			testIP:     "169.254.2.4",
			resultIP:   "169.254.2.4",
		},
	}

	mgr := NewDummyInterfaceController()
	for k, tc := range testcases {
		t.Run(k, func(t *testing.T) {
			if len(tc.preparedIP) != 0 {
				err := mgr.EnsureDummyInterface(testDummyIfName, net.ParseIP(tc.preparedIP))
				if err != nil {
					t.Errorf("failed to prepare dummy interface with ip(%s), %v", tc.preparedIP, err)
				}
				ips, err := mgr.ListDummyInterface(testDummyIfName)
				if err != nil || len(ips) == 0 {
					t.Errorf("failed to prepare dummy interface(%s: %s), %v", testDummyIfName, tc.preparedIP, err)
				}
			}

			err := mgr.EnsureDummyInterface(testDummyIfName, net.ParseIP(tc.testIP))
			if err != nil {
				t.Errorf("failed to ensure dummy interface with ip(%s), %v", tc.testIP, err)
			}

			ips2, err := mgr.ListDummyInterface(testDummyIfName)
			if err != nil || len(ips2) == 0 {
				t.Errorf("failed to list dummy interface(%s), %v", testDummyIfName, err)
			}

			sameIP := false
			for _, ip := range ips2 {
				if ip.String() == tc.resultIP {
					sameIP = true
					break
				}
			}

			if !sameIP {
				t.Errorf("dummy if with ip(%s) is not ensured, addrs: %s", tc.resultIP, ips2[0].String())
			}

			// delete dummy interface
			err = mgr.DeleteDummyInterface(testDummyIfName)
			if err != nil {
				t.Errorf("failed to delte dummy interface, %v", err)
			}
		})
	}
}
