//go:build linux
// +build linux

package containerd

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
	"io/fs"
	"reflect"
	"testing"
	"testing/fstest"

	"github.com/bhojpur/dcp/pkg/cloud/agent/templates"
)

func Test_UnitFindNvidiaContainerRuntimes(t *testing.T) {
	executable := &fstest.MapFile{Mode: 0755}
	type args struct {
		root fs.FS
	}
	tests := []struct {
		name string
		args args
		want map[string]templates.ContainerdRuntimeConfig
	}{
		{
			name: "No runtimes",
			args: args{
				root: fstest.MapFS{},
			},
			want: map[string]templates.ContainerdRuntimeConfig{},
		},
		{
			name: "Nvidia runtime in /usr/bin",
			args: args{
				root: fstest.MapFS{
					"usr/bin/nvidia-container-runtime": executable,
				},
			},
			want: map[string]templates.ContainerdRuntimeConfig{
				"nvidia": {
					RuntimeType: "io.containerd.runc.v2",
					BinaryName:  "/usr/bin/nvidia-container-runtime",
				},
			},
		},
		{
			name: "Experimental runtime in /usr/local/nvidia/toolkit",
			args: args{
				root: fstest.MapFS{
					"usr/local/nvidia/toolkit/nvidia-container-runtime": executable,
				},
			},
			want: map[string]templates.ContainerdRuntimeConfig{
				"nvidia": {
					RuntimeType: "io.containerd.runc.v2",
					BinaryName:  "/usr/local/nvidia/toolkit/nvidia-container-runtime",
				},
			},
		},
		{
			name: "Two runtimes in separate directories",
			args: args{
				root: fstest.MapFS{
					"usr/bin/nvidia-container-runtime":                  executable,
					"usr/local/nvidia/toolkit/nvidia-container-runtime": executable,
				},
			},
			want: map[string]templates.ContainerdRuntimeConfig{
				"nvidia": {
					RuntimeType: "io.containerd.runc.v2",
					BinaryName:  "/usr/local/nvidia/toolkit/nvidia-container-runtime",
				},
			},
		},
		{
			name: "Experimental runtime in /usr/bin",
			args: args{
				root: fstest.MapFS{
					"usr/bin/nvidia-container-runtime-experimental": executable,
				},
			},
			want: map[string]templates.ContainerdRuntimeConfig{
				"nvidia-experimental": {
					RuntimeType: "io.containerd.runc.v2",
					BinaryName:  "/usr/bin/nvidia-container-runtime-experimental",
				},
			},
		},
		{
			name: "Same runtime in two directories",
			args: args{
				root: fstest.MapFS{
					"usr/bin/nvidia-container-runtime-experimental":                  executable,
					"usr/local/nvidia/toolkit/nvidia-container-runtime-experimental": executable,
				},
			},
			want: map[string]templates.ContainerdRuntimeConfig{
				"nvidia-experimental": {
					RuntimeType: "io.containerd.runc.v2",
					BinaryName:  "/usr/local/nvidia/toolkit/nvidia-container-runtime-experimental",
				},
			},
		},
		{
			name: "Both runtimes in /usr/bin",
			args: args{
				root: fstest.MapFS{
					"usr/bin/nvidia-container-runtime-experimental": executable,
					"usr/bin/nvidia-container-runtime":              executable,
				},
			},
			want: map[string]templates.ContainerdRuntimeConfig{
				"nvidia": {
					RuntimeType: "io.containerd.runc.v2",
					BinaryName:  "/usr/bin/nvidia-container-runtime",
				},
				"nvidia-experimental": {
					RuntimeType: "io.containerd.runc.v2",
					BinaryName:  "/usr/bin/nvidia-container-runtime-experimental",
				},
			},
		},
		{
			name: "Both runtimes in both directories",
			args: args{
				root: fstest.MapFS{
					"usr/local/nvidia/toolkit/nvidia-container-runtime":              executable,
					"usr/local/nvidia/toolkit/nvidia-container-runtime-experimental": executable,
					"usr/bin/nvidia-container-runtime":                               executable,
					"usr/bin/nvidia-container-runtime-experimental":                  executable,
				},
			},
			want: map[string]templates.ContainerdRuntimeConfig{
				"nvidia": {
					RuntimeType: "io.containerd.runc.v2",
					BinaryName:  "/usr/local/nvidia/toolkit/nvidia-container-runtime",
				},
				"nvidia-experimental": {
					RuntimeType: "io.containerd.runc.v2",
					BinaryName:  "/usr/local/nvidia/toolkit/nvidia-container-runtime-experimental",
				},
			},
		},
		{
			name: "Both runtimes in /usr/local/nvidia/toolkit",
			args: args{
				root: fstest.MapFS{
					"usr/local/nvidia/toolkit/nvidia-container-runtime":              executable,
					"usr/local/nvidia/toolkit/nvidia-container-runtime-experimental": executable,
				},
			},
			want: map[string]templates.ContainerdRuntimeConfig{
				"nvidia": {
					RuntimeType: "io.containerd.runc.v2",
					BinaryName:  "/usr/local/nvidia/toolkit/nvidia-container-runtime",
				},
				"nvidia-experimental": {
					RuntimeType: "io.containerd.runc.v2",
					BinaryName:  "/usr/local/nvidia/toolkit/nvidia-container-runtime-experimental",
				},
			},
		},
		{
			name: "Both runtimes in /usr/bin and one duplicate in /usr/local/nvidia/toolkit",
			args: args{
				root: fstest.MapFS{
					"usr/bin/nvidia-container-runtime":                               executable,
					"usr/bin/nvidia-container-runtime-experimental":                  executable,
					"usr/local/nvidia/toolkit/nvidia-container-runtime-experimental": executable,
				},
			},
			want: map[string]templates.ContainerdRuntimeConfig{
				"nvidia": {
					RuntimeType: "io.containerd.runc.v2",
					BinaryName:  "/usr/bin/nvidia-container-runtime",
				},
				"nvidia-experimental": {
					RuntimeType: "io.containerd.runc.v2",
					BinaryName:  "/usr/local/nvidia/toolkit/nvidia-container-runtime-experimental",
				},
			},
		},
		{
			name: "Runtime is a directory",
			args: args{
				root: fstest.MapFS{
					"usr/bin/nvidia-container-runtime": &fstest.MapFile{
						Mode: fs.ModeDir,
					},
				},
			},
			want: map[string]templates.ContainerdRuntimeConfig{},
		},
		{
			name: "Runtime in both directories, but one is a directory",
			args: args{
				root: fstest.MapFS{
					"usr/bin/nvidia-container-runtime": executable,
					"usr/local/nvidia/toolkit/nvidia-container-runtime": &fstest.MapFile{
						Mode: fs.ModeDir,
					},
				},
			},
			want: map[string]templates.ContainerdRuntimeConfig{
				"nvidia": {
					RuntimeType: "io.containerd.runc.v2",
					BinaryName:  "/usr/bin/nvidia-container-runtime",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findNvidiaContainerRuntimes(tt.args.root); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findNvidiaContainerRuntimes() = %+v\nWant = %+v", got, tt.want)
			}
		})
	}
}
