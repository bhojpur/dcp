package nodeconfig

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
	"os"
	"testing"

	"github.com/bhojpur/dcp/pkg/cloud/version"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var FakeNodeWithNoAnnotation = &corev1.Node{
	TypeMeta: metav1.TypeMeta{
		Kind:       "Node",
		APIVersion: "v1",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name: "fakeNode-no-annotation",
	},
}

var TestEnvName = version.ProgramUpper + "_NODE_NAME"
var FakeNodeWithAnnotation = &corev1.Node{
	TypeMeta: metav1.TypeMeta{
		Kind:       "Node",
		APIVersion: "v1",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name: "fakeNode-with-annotation",
		Annotations: map[string]string{
			NodeArgsAnnotation:       `["server","--no-flannel"]`,
			NodeEnvAnnotation:        `{"` + TestEnvName + `":"fakeNode-with-annotation"}`,
			NodeConfigHashAnnotation: "LNQOAOIMOQIBRMEMACW7LYHXUNPZADF6RFGOSPIHJCOS47UVUJAA====",
		},
	},
}

func Test_UnitSetExistingNodeConfigAnnotations(t *testing.T) {
	// adding same config
	os.Args = []string{version.Program, "server", "--no-flannel"}
	os.Setenv(version.ProgramUpper+"_NODE_NAME", "fakeNode-with-annotation")
	nodeUpdated, err := SetNodeConfigAnnotations(FakeNodeWithAnnotation)
	if err != nil {
		t.Fatalf("Failed to set node config annotation: %v", err)
	}
	if nodeUpdated {
		t.Errorf("Test_UnitSetExistingNodeConfigAnnotations() expected false")
	}
}

func Test_UnitSetNodeConfigAnnotations(t *testing.T) {
	type args struct {
		node   *corev1.Node
		osArgs []string
	}
	setup := func(osArgs []string) error {
		os.Args = osArgs
		return os.Setenv(TestEnvName, "fakeNode-with-no-annotation")
	}
	teardown := func() error {
		return os.Unsetenv(TestEnvName)
	}
	tests := []struct {
		name               string
		args               args
		want               bool
		wantErr            bool
		wantNodeArgs       string
		wantNodeEnv        string
		wantNodeConfigHash string
	}{
		{
			name: "Set empty NodeConfigAnnotations",
			args: args{
				node:   FakeNodeWithAnnotation,
				osArgs: []string{version.Program, "server", "--no-flannel"},
			},
			want:               true,
			wantNodeArgs:       `["server","--no-flannel"]`,
			wantNodeEnv:        `{"` + TestEnvName + `":"fakeNode-with-no-annotation"}`,
			wantNodeConfigHash: "FBV4UQYLF2N7NH7EK42GKOTU5YA24TXB4WAYZHA5ZOFNGZHC4ZPA====",
		},
		{
			name: "Set args with equal",
			args: args{
				node:   FakeNodeWithNoAnnotation,
				osArgs: []string{version.Program, "server", "--no-flannel", "--write-kubeconfig-mode=777"},
			},
			want:         true,
			wantNodeArgs: `["server","--no-flannel","--write-kubeconfig-mode","777"]`,
			wantNodeEnv:  `{"` + TestEnvName + `":"fakeNode-with-no-annotation"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer teardown()
			if err := setup(tt.args.osArgs); err != nil {
				t.Errorf("Setup for SetNodeConfigAnnotations() failed = %v", err)
				return
			}
			got, err := SetNodeConfigAnnotations(tt.args.node)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetNodeConfigAnnotations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SetNodeConfigAnnotations() = %+v\nWantRes = %+v", got, tt.want)
			}
			nodeAnn := tt.args.node.Annotations
			if nodeAnn[NodeArgsAnnotation] != tt.wantNodeArgs {
				t.Errorf("SetNodeConfigAnnotations() = %+v\nWantAnn.nodeArgs = %+v", nodeAnn[NodeArgsAnnotation], tt.wantNodeArgs)
			}
			if nodeAnn[NodeEnvAnnotation] != tt.wantNodeEnv {
				t.Errorf("SetNodeConfigAnnotations() = %+v\nWantAnn.nodeEnv = %+v", nodeAnn[NodeEnvAnnotation], tt.wantNodeEnv)
			}
			if tt.wantNodeConfigHash != "" && nodeAnn[NodeConfigHashAnnotation] != tt.wantNodeConfigHash {
				t.Errorf("SetNodeConfigAnnotations() = %+v\nWantAnn.nodeConfigHash = %+v", nodeAnn[NodeConfigHashAnnotation], tt.wantNodeConfigHash)
			}
		})
	}
}
