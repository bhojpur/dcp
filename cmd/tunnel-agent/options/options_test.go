package options

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

import "testing"

func TestAgentIdentifiersAreValid(t *testing.T) {
	testcases := map[string]struct {
		agentIdentifiers string
		result           bool
	}{
		"empty agent identifiers": {
			"",
			true,
		},

		"valid agent identifiers": {
			"host=node-test",
			true,
		},

		"invalid agent identifiers without value": {
			"host",
			false,
		},

		"invalid agent identifiers with invalid key": {
			"foo=node-test",
			false,
		},
	}

	for k, tc := range testcases {
		result := agentIdentifiersAreValid(tc.agentIdentifiers)
		if result != tc.result {
			t.Errorf("%s: agent identifiers are valid  verification result, expect %v, but got %v", k, tc.result, result)
		}
	}
}
