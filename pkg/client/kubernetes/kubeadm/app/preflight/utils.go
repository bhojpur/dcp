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
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/version"
	utilsexec "k8s.io/utils/exec"
)

// GetKubeletVersion is helper function that returns version of kubelet available in $PATH
func GetKubeletVersion(execer utilsexec.Interface) (*version.Version, error) {
	kubeletVersionRegex := regexp.MustCompile(`^\s*Kubernetes v((0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)([-0-9a-zA-Z_\.+]*)?)\s*$`)

	command := execer.Command("kubelet", "--version")
	out, err := command.Output()
	if err != nil {
		return nil, errors.Wrap(err, "cannot execute 'kubelet --version'")
	}

	cleanOutput := strings.TrimSpace(string(out))
	subs := kubeletVersionRegex.FindAllStringSubmatch(cleanOutput, -1)
	if len(subs) != 1 || len(subs[0]) < 2 {
		return nil, errors.Errorf("Unable to parse output from Kubelet: %q", cleanOutput)
	}
	return version.ParseSemantic(subs[0][1])
}
