package kubelet

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
	"strings"
	"testing"
)

func TestConstructNodeLabels(t *testing.T) {
	edgeWorkerLabel := "bhojpur.net/is-edge-worker"
	testcases := map[string]struct {
		nodeLabels map[string]string
		mode       string
		result     map[string]string
	}{
		"no input node labels with cloud mode": {
			mode: "cloud",
			result: map[string]string{
				"bhojpur.net/is-edge-worker": "false",
			},
		},
		"one input node labels with cloud mode": {
			nodeLabels: map[string]string{"foo": "bar"},
			mode:       "cloud",
			result: map[string]string{
				"bhojpur.net/is-edge-worker": "false",
				"foo":                        "bar",
			},
		},
		"more than one input node labels with cloud mode": {
			nodeLabels: map[string]string{
				"foo":  "bar",
				"foo2": "bar2",
			},
			mode: "cloud",
			result: map[string]string{
				"bhojpur.net/is-edge-worker": "false",
				"foo":                        "bar",
				"foo2":                       "bar2",
			},
		},
		"no input node labels with edge mode": {
			mode: "edge",
			result: map[string]string{
				"bhojpur.net/is-edge-worker": "true",
			},
		},
		"one input node labels with edge mode": {
			nodeLabels: map[string]string{"foo": "bar"},
			mode:       "edge",
			result: map[string]string{
				"bhojpur.net/is-edge-worker": "true",
				"foo":                        "bar",
			},
		},
	}

	for k, tc := range testcases {
		t.Run(k, func(t *testing.T) {
			constructedLabelsStr := constructNodeLabels(tc.nodeLabels, tc.mode, edgeWorkerLabel)
			constructedLabels := make(map[string]string)
			parts := strings.Split(constructedLabelsStr, ",")
			for i := range parts {
				kv := strings.Split(parts[i], "=")
				if len(kv) == 2 {
					constructedLabels[kv[0]] = kv[1]
				}
			}

			if !reflect.DeepEqual(constructedLabels, tc.result) {
				t.Errorf("expected node labels: %v, but got %v", tc.result, constructedLabels)
			}
		})
	}

}
