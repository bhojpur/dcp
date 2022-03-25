package preflight

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

	"github.com/pkg/errors"
	utilsexec "k8s.io/utils/exec"
	fakeexec "k8s.io/utils/exec/testing"
)

func TestGetKubeletVersion(t *testing.T) {
	cases := []struct {
		output   string
		expected string
		err      error
		valid    bool
	}{
		{"Kubernetes v1.7.0", "1.7.0", nil, true},
		{"Kubernetes v1.8.0-alpha.2.1231+afabd012389d53a", "1.8.0-alpha.2.1231+afabd012389d53a", nil, true},
		{"something-invalid", "", nil, false},
		{"command not found", "", errors.New("kubelet not found"), false},
		{"", "", nil, false},
	}

	for _, tc := range cases {
		t.Run(tc.output, func(t *testing.T) {
			fcmd := fakeexec.FakeCmd{
				OutputScript: []fakeexec.FakeAction{
					func() ([]byte, []byte, error) { return []byte(tc.output), nil, tc.err },
				},
			}
			fexec := &fakeexec.FakeExec{
				CommandScript: []fakeexec.FakeCommandAction{
					func(cmd string, args ...string) utilsexec.Cmd { return fakeexec.InitFakeCmd(&fcmd, cmd, args...) },
				},
			}
			ver, err := GetKubeletVersion(fexec)
			switch {
			case err != nil && tc.valid:
				t.Errorf("GetKubeletVersion: unexpected error for %q. Error: %v", tc.output, err)
			case err == nil && !tc.valid:
				t.Errorf("GetKubeletVersion: error expected for key %q, but result is %q", tc.output, ver)
			case ver != nil && ver.String() != tc.expected:
				t.Errorf("GetKubeletVersion: unexpected version result for key %q. Expected: %q Actual: %q", tc.output, tc.expected, ver)
			}
		})
	}
}
