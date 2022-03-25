package configfilearg

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

func Test_UnitMustParse(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		config string
		want   []string
	}{
		{
			name: "Basic server",
			args: []string{"dcp", "server"},

			want: []string{"dcp", "server"},
		},
		{
			name: "Server with known flags",
			args: []string{"dcp", "server", "-t 12345", "--write-kubeconfig-mode 644"},

			want: []string{"dcp", "server", "-t 12345", "--write-kubeconfig-mode 644"},
		},
		{
			name:   "Server with known flags and config with known and unknown flags",
			args:   []string{"dcp", "server", "--write-kubeconfig-mode 644"},
			config: "./testdata/defaultdata.yaml",
			want: []string{"dcp", "server", "--token=12345", "--node-label=DEAFBEEF",
				"--etcd-s3=true", "--etcd-s3-bucket=my-backup", "--kubelet-arg=max-pods=999", "--write-kubeconfig-mode 644"},
		},
		{
			name: "Basic etcd-snapshot",
			args: []string{"dcp", "etcd-snapshot", "save"},

			want: []string{"dcp", "etcd-snapshot", "save"},
		},
		{
			name: "Etcd-snapshot with known flags",
			args: []string{"dcp", "etcd-snapshot", "save", "--s3=true"},

			want: []string{"dcp", "etcd-snapshot", "save", "--s3=true"},
		},
		{
			name:   "Etcd-snapshot with config with known and unknown flags",
			args:   []string{"dcp", "etcd-snapshot", "save"},
			config: "./testdata/defaultdata.yaml",
			want:   []string{"dcp", "etcd-snapshot", "save", "--etcd-s3=true", "--etcd-s3-bucket=my-backup"},
		},
		{
			name: "Agent with known flags",
			args: []string{"dcp", "agent", "--token=12345"},

			want: []string{"dcp", "agent", "--token=12345"},
		},
		{
			name:   "Agent with config with known and unknown flags, flags are not skipped",
			args:   []string{"dcp", "agent"},
			config: "./testdata/defaultdata.yaml",
			want: []string{"dcp", "agent", "--token=12345", "--node-label=DEAFBEEF",
				"--etcd-s3=true", "--etcd-s3-bucket=my-backup", "--notaflag=true", "--kubelet-arg=max-pods=999"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DefaultParser.DefaultConfig = tt.config
			if got := MustParse(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MustParse() = %+v\nWant = %+v", got, tt.want)
			}
		})
	}
}
