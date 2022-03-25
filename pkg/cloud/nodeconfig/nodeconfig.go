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
	"crypto/sha256"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/bhojpur/dcp/pkg/cloud/configfilearg"
	"github.com/bhojpur/dcp/pkg/cloud/version"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

var (
	NodeArgsAnnotation       = version.Program + ".io/node-args"
	NodeEnvAnnotation        = version.Program + ".io/node-env"
	NodeConfigHashAnnotation = version.Program + ".io/node-config-hash"
)

const (
	OmittedValue = "********"
)

func getNodeArgs() (string, error) {
	nodeArgsList := []string{}
	for _, arg := range configfilearg.MustParse(os.Args[1:]) {
		if strings.HasPrefix(arg, "--") && strings.Contains(arg, "=") {
			parsedArg := strings.SplitN(arg, "=", 2)
			nodeArgsList = append(nodeArgsList, parsedArg...)
			continue
		}
		nodeArgsList = append(nodeArgsList, arg)
	}
	for i, arg := range nodeArgsList {
		if isSecret(arg) {
			if i+1 < len(nodeArgsList) {
				nodeArgsList[i+1] = OmittedValue
			}
		}
	}
	nodeArgs, err := json.Marshal(nodeArgsList)
	if err != nil {
		return "", errors.Wrap(err, "Failed to retrieve argument list for node")
	}
	return string(nodeArgs), nil
}

func getNodeEnv() (string, error) {
	dcpEnv := make(map[string]string)
	for _, v := range os.Environ() {
		keyValue := strings.SplitN(v, "=", 2)
		if strings.HasPrefix(keyValue[0], version.ProgramUpper+"_") {
			dcpEnv[keyValue[0]] = keyValue[1]
		}
	}
	for key := range dcpEnv {
		if isSecret(key) {
			dcpEnv[key] = OmittedValue
		}
	}
	dcpEnvJSON, err := json.Marshal(dcpEnv)
	if err != nil {
		return "", errors.Wrap(err, "Failed to retrieve environment map for node")
	}
	return string(dcpEnvJSON), nil
}

func SetNodeConfigAnnotations(node *corev1.Node) (bool, error) {
	nodeArgs, err := getNodeArgs()
	if err != nil {
		return false, err
	}
	nodeEnv, err := getNodeEnv()
	if err != nil {
		return false, err
	}
	h := sha256.New()
	_, err = h.Write([]byte(nodeArgs + nodeEnv))
	if err != nil {
		return false, fmt.Errorf("Failed to hash the node config: %v", err)
	}
	if node.Annotations == nil {
		node.Annotations = make(map[string]string)
	}
	configHash := h.Sum(nil)
	encoded := base32.StdEncoding.EncodeToString(configHash[:])
	if node.Annotations[NodeConfigHashAnnotation] == encoded {
		return false, nil
	}

	node.Annotations[NodeEnvAnnotation] = nodeEnv
	node.Annotations[NodeArgsAnnotation] = nodeArgs
	node.Annotations[NodeConfigHashAnnotation] = encoded
	return true, nil
}

func isSecret(key string) bool {
	secretData := []string{
		version.ProgramUpper + "_TOKEN",
		version.ProgramUpper + "_DATASTORE_ENDPOINT",
		version.ProgramUpper + "_AGENT_TOKEN",
		version.ProgramUpper + "_CLUSTER_SECRET",
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"--token",
		"-t",
		"--agent-token",
		"--datastore-endpoint",
		"--cluster-secret",
		"--etcd-s3-access-key",
		"--etcd-s3-secret-key",
	}
	for _, secret := range secretData {
		if key == secret {
			return true
		}
	}
	return false
}
