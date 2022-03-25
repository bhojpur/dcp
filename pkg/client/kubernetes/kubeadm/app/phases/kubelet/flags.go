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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/pkg/client/cmd/join/joindata"
	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/constants"
	kubeadmutil "github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/util"
	"github.com/bhojpur/dcp/pkg/projectinfo"
)

// WriteKubeletDynamicEnvFile writes an environment file with dynamic flags to the kubelet.
// Used at "kubeadm init" and "kubeadm join" time.
func WriteKubeletDynamicEnvFile(data joindata.DcpJoinData, kubeletDir string) error {
	stringMap := buildKubeletArgMap(data)
	argList := kubeadmutil.BuildArgumentListFromMap(stringMap, map[string]string{})
	envFileContent := fmt.Sprintf("%s=%q\n", constants.KubeletEnvFileVariableName, strings.Join(argList, " "))

	return writeKubeletFlagBytesToDisk([]byte(envFileContent), kubeletDir)
}

//buildKubeletArgMapCommon takes a kubeletFlagsOpts object and builds based on that a string-string map with flags
//that are common to both Linux and Windows
func buildKubeletArgMapCommon(data joindata.DcpJoinData) map[string]string {
	kubeletFlags := map[string]string{}

	nodeReg := data.NodeRegistration()
	if nodeReg.CRISocket == constants.DefaultDockerCRISocket {
		// These flags should only be set when running docker
		kubeletFlags["network-plugin"] = "cni"
		if data.PauseImage() != "" {
			kubeletFlags["pod-infra-container-image"] = data.PauseImage()
		}
	} else {
		kubeletFlags["container-runtime"] = "remote"
		kubeletFlags["container-runtime-endpoint"] = nodeReg.CRISocket
	}

	hostname, err := os.Hostname()
	if err != nil {
		klog.Warning(err)
	}
	if nodeReg.Name != hostname {
		klog.V(1).Infof("setting kubelet hostname-override to %q", nodeReg.Name)
		kubeletFlags["hostname-override"] = nodeReg.Name
	}

	kubeletFlags["node-labels"] = constructNodeLabels(data.NodeLabels(), nodeReg.WorkingMode, projectinfo.GetEdgeWorkerLabelKey())

	kubeletFlags["rotate-certificates"] = "false"

	return kubeletFlags
}

// constructNodeLabels make up node labels string
func constructNodeLabels(nodeLabels map[string]string, workingMode, edgeWorkerLabel string) string {
	if nodeLabels == nil {
		nodeLabels = make(map[string]string)
	}
	if _, ok := nodeLabels[edgeWorkerLabel]; !ok {
		if workingMode == "cloud" {
			nodeLabels[edgeWorkerLabel] = "false"
		} else {
			nodeLabels[edgeWorkerLabel] = "true"
		}
	}
	var labelsStr string
	for k, v := range nodeLabels {
		if len(labelsStr) == 0 {
			labelsStr = fmt.Sprintf("%s=%s", k, v)
		} else {
			labelsStr = fmt.Sprintf("%s,%s=%s", labelsStr, k, v)
		}
	}

	return labelsStr
}

// writeKubeletFlagBytesToDisk writes a byte slice down to disk at the specific location of the kubelet flag overrides file
func writeKubeletFlagBytesToDisk(b []byte, kubeletDir string) error {
	kubeletEnvFilePath := filepath.Join(kubeletDir, constants.KubeletEnvFileName)
	fmt.Printf("[kubelet-start] Writing kubelet environment file with flags to file %q\n", kubeletEnvFilePath)

	// creates target folder if not already exists
	if err := os.MkdirAll(kubeletDir, 0700); err != nil {
		return errors.Wrapf(err, "failed to create directory %q", kubeletDir)
	}
	if err := ioutil.WriteFile(kubeletEnvFilePath, b, 0644); err != nil {
		return errors.Wrapf(err, "failed to write kubelet configuration to the file %q", kubeletEnvFilePath)
	}
	return nil
}
