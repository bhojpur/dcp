//go:build windows
// +build windows

package agent

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

	"github.com/bhojpur/dcp/pkg/cloud/daemons/config"
)

func TestCheckRuntimeEndpoint(t *testing.T) {
	type args struct {
		cfg *config.Agent
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Runtime endpoint unaltered",
			args: args{
				cfg: &config.Agent{RuntimeSocket: "npipe:////./pipe/containerd-containerd"},
			},
			want: "npipe:////./pipe/containerd-containerd",
		},
		{
			name: "Runtime endpoint altered",
			args: args{
				cfg: &config.Agent{RuntimeSocket: "//./pipe/containerd-containerd"},
			},
			want: "npipe:////./pipe/containerd-containerd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argsMap := map[string]string{}
			checkRuntimeEndpoint(tt.args.cfg, argsMap)
			if argsMap["container-runtime-endpoint"] != tt.want {
				got := argsMap["container-runtime-endpoint"]
				t.Errorf("error, input was " + tt.args.cfg.RuntimeSocket + " should be " + tt.want + ", but got " + got)
			}
		})

	}
}
