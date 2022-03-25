package components

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
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"k8s.io/klog/v2"

	enutil "github.com/bhojpur/dcp/pkg/client/util/edgenode"
)

const (
	kubeletConfigRegularExpression = "\\-\\-kubeconfig=.*kubelet.conf"
	apiserverAddrRegularExpression = "server: (http(s)?:\\/\\/)?[\\w][-\\w]{0,62}(\\.[\\w][-\\w]{0,62})*(:[\\d]{1,5})?"

	kubeAdmFlagsEnvFile = "/var/lib/kubelet/kubeadm-flags.env"
	dirMode             = 0755
)

type kubeletOperator struct {
	dcpDir string
}

// NewKubeletOperator create kubeletOperator
func NewKubeletOperator(bhojpurDir string) *kubeletOperator {
	return &kubeletOperator{
		dcpDir: bhojpurDir,
	}
}

// RedirectTrafficToEngine
// add env config leads kubelet to visit Bhojpur DCP server engine as apiServer
func (op *kubeletOperator) RedirectTrafficToEngine() error {
	// 1. create a working dir to store revised kubelet.conf
	_, err := op.writeEngineKubeletConfig()
	if err != nil {
		return err
	}

	// 2. append /var/lib/kubelet/kubeadm-flags.env
	if err := op.appendConfig(); err != nil {
		return err
	}

	// 3. restart
	return restartKubeletService()
}

// UndoRedirectTrafficToEngine
// undo what's done to kubelet and restart
// to compatible the old convert way for a while , so do renameSvcBk
func (op *kubeletOperator) UndoRedirectTrafficToEngine() error {
	if err := op.undoAppendConfig(); err != nil {
		return err
	}

	if err := restartKubeletService(); err != nil {
		return err
	}

	if err := op.undoWriteEngineKubeletConfig(); err != nil {
		return err
	}
	klog.Info("revertKubelet: undoWriteEngineKubeletConfig finished")

	return nil
}

func (op *kubeletOperator) writeEngineKubeletConfig() (string, error) {
	err := os.MkdirAll(op.dcpDir, dirMode)
	if err != nil {
		return "", err
	}
	fullPath := op.getEngineKubeletConf()
	err = ioutil.WriteFile(fullPath, []byte(enutil.DcpKubeletConf), fileMode)
	if err != nil {
		return "", err
	}
	klog.Infof("revised kubeconfig %s is generated", fullPath)
	return fullPath, nil
}

func (op *kubeletOperator) undoWriteEngineKubeletConfig() error {
	dcpKubeletConf := op.getEngineKubeletConf()
	if _, err := enutil.FileExists(dcpKubeletConf); err != nil && os.IsNotExist(err) {
		return nil
	}

	return os.Remove(dcpKubeletConf)
}

func (op *kubeletOperator) appendConfig() error {
	// set env KUBELET_KUBEADM_ARGS, args set later will override before
	// ExecStart: kubelet $KUBELET_KUBECONFIG_ARGS $KUBELET_CONFIG_ARGS $KUBELET_KUBEADM_ARGS $KUBELET_EXTRA_ARGS
	// append setup: " --kubeconfig=$engineKubeletConf -bootstrap-kubeconfig= "
	kubeConfigSetup := op.getAppendSetting()

	// if wrote, return
	content, err := ioutil.ReadFile(kubeAdmFlagsEnvFile)
	if err != nil {
		return err
	}
	args := string(content)
	if strings.Contains(args, kubeConfigSetup) {
		klog.Info("kubeConfigSetup has wrote before")
		return nil
	}

	// append KUBELET_KUBEADM_ARGS
	argsRegexp := regexp.MustCompile(`KUBELET_KUBEADM_ARGS="(.+)"`)
	finding := argsRegexp.FindStringSubmatch(args)
	if len(finding) != 2 {
		return fmt.Errorf("kubeadm-flags.env error format. %s", args)
	}

	r := strings.Replace(args, finding[1], fmt.Sprintf("%s %s", finding[1], kubeConfigSetup), 1)
	err = ioutil.WriteFile(kubeAdmFlagsEnvFile, []byte(r), fileMode)
	if err != nil {
		return err
	}

	return nil
}

func (op *kubeletOperator) undoAppendConfig() error {
	kubeConfigSetup := op.getAppendSetting()
	contentbyte, err := ioutil.ReadFile(kubeAdmFlagsEnvFile)
	if err != nil {
		return err
	}

	content := strings.ReplaceAll(string(contentbyte), kubeConfigSetup, "")
	err = ioutil.WriteFile(kubeAdmFlagsEnvFile, []byte(content), fileMode)
	if err != nil {
		return err
	}
	klog.Info("revertKubelet: undoAppendConfig finished")

	return nil
}

func (op *kubeletOperator) getAppendSetting() string {
	configPath := op.getEngineKubeletConf()
	return fmt.Sprintf(" --kubeconfig=%s --bootstrap-kubeconfig= ", configPath)
}

func (op *kubeletOperator) getEngineKubeletConf() string {
	return filepath.Join(op.dcpDir, enutil.KubeletConfName)
}

func restartKubeletService() error {
	klog.Info("restartKubelet: " + enutil.DaemonReload)
	cmd := exec.Command("bash", "-c", enutil.DaemonReload)
	if err := enutil.Exec(cmd); err != nil {
		return err
	}
	klog.Info("restartKubelet: " + enutil.RestartKubeletSvc)
	cmd = exec.Command("bash", "-c", enutil.RestartKubeletSvc)
	if err := enutil.Exec(cmd); err != nil {
		return err
	}
	klog.Infof("restartKubelet: kubelet has been restarted")
	return nil
}

// GetApiServerAddress parse apiServer address from conf file
func GetApiServerAddress(kubeadmConfPaths []string) (string, error) {
	var kbcfg string
	for _, path := range kubeadmConfPaths {
		if exist, _ := enutil.FileExists(path); exist {
			kbcfg = path
			break
		}
	}
	if kbcfg == "" {
		return "", fmt.Errorf("get apiserverAddr err: no file exists in list %s", kubeadmConfPaths)
	}

	kubeletConfPath, err := enutil.GetSingleContentFromFile(kbcfg, kubeletConfigRegularExpression)
	if err != nil {
		return "", err
	}

	confArr := strings.Split(kubeletConfPath, "=")
	if len(confArr) != 2 {
		return "", fmt.Errorf("get kubeletConfPath format err:%s", kubeletConfPath)
	}
	kubeletConfPath = confArr[1]
	apiserverAddr, err := enutil.GetSingleContentFromFile(kubeletConfPath, apiserverAddrRegularExpression)
	if err != nil {
		return "", err
	}

	addrArr := strings.Split(apiserverAddr, " ")
	if len(addrArr) != 2 {
		return "", fmt.Errorf("get apiserverAddr format err:%s", apiserverAddr)
	}
	apiserverAddr = addrArr[1]
	return apiserverAddr, nil
}
