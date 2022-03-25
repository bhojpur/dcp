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
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"

	enutil "github.com/bhojpur/dcp/pkg/client/util/edgenode"
	"github.com/bhojpur/dcp/pkg/client/util/templates"
	"github.com/bhojpur/dcp/pkg/engine/certificate/hubself"
	"github.com/bhojpur/dcp/pkg/engine/storage/disk"
	"github.com/bhojpur/dcp/pkg/engine/util"
)

const (
	hubHealthzCheckFrequency = 10 * time.Second
	fileMode                 = 0666
)

type engineOperator struct {
	apiServerAddr            string
	engineImage              string
	joinToken                string
	workingMode              util.WorkingMode
	engineHealthCheckTimeout time.Duration
}

// NewEngineOperator new engineOperator struct
func NewEngineOperator(apiServerAddr string, engineImage string, joinToken string,
	workingMode util.WorkingMode, engineHealthCheckTimeout time.Duration) *engineOperator {
	return &engineOperator{
		apiServerAddr:            apiServerAddr,
		engineImage:              engineImage,
		joinToken:                joinToken,
		workingMode:              workingMode,
		engineHealthCheckTimeout: engineHealthCheckTimeout,
	}
}

// Install set Bhojpur DCP server engine yaml to static path to start pod
func (op *engineOperator) Install() error {

	// 1. put dcpsvr yaml into /etc/kubernetes/manifests
	klog.Infof("setting up Bhojpur DCP server engine on node")

	// 1-1. replace variables in yaml file
	klog.Infof("setting up Bhojpur DCP server engine apiServer addr")
	engineTemplate, err := templates.SubsituteTemplate(enutil.EngineTemplate, map[string]string{
		"kubernetesServerAddr": op.apiServerAddr,
		"image":                op.engineImage,
		"joinToken":            op.joinToken,
		"workingMode":          string(op.workingMode),
	})
	if err != nil {
		return err
	}

	// 1-2. create dcpsvr.yaml
	podManifestPath := enutil.GetPodManifestPath()
	if err := enutil.EnsureDir(podManifestPath); err != nil {
		return err
	}
	if err := ioutil.WriteFile(getEngineYaml(podManifestPath), []byte(engineTemplate), fileMode); err != nil {
		return err
	}
	klog.Infof("create the %s/dcpsvr.yaml", podManifestPath)

	// 2. wait Bhojpur DCP server engine pod to be ready
	return engineHealthcheck(op.engineHealthCheckTimeout)
}

// UnInstall remove yaml and configs of Bhojpur DCP server engine
func (op *engineOperator) UnInstall() error {
	// 1. remove the dcpsvr.yaml to delete the dcpsvr
	podManifestPath := enutil.GetPodManifestPath()
	engineYamlPath := getEngineYaml(podManifestPath)
	if _, err := enutil.FileExists(engineYamlPath); os.IsNotExist(err) {
		klog.Infof("UninstallEngine: %s is not exists, skip delete", engineYamlPath)
	} else {
		err := os.Remove(engineYamlPath)
		if err != nil {
			return err
		}
		klog.Infof("UninstallEngine: %s has been removed", engineYamlPath)
	}

	// 2. remove dcpsvr config directory and certificates in it
	engineConf := getEngineConf()
	if _, err := enutil.FileExists(engineConf); os.IsNotExist(err) {
		klog.Infof("UninstallEngine: dir %s is not exists, skip delete", engineConf)
		return nil
	}
	err := os.RemoveAll(engineConf)
	if err != nil {
		return err
	}
	klog.Infof("UninstallEngine: config dir %s  has been removed", engineConf)

	// 3. remove Bhojpur DCP server engine cache dir
	// since k8s may takes a while to notice and remove Bhojpur DCP server engine pod, we have to wait for that.
	// because, if we delete dir before Bhojpur DCP exit, Bhojpur DCP may recreate cache/kubelet dir before exit.
	err = waitUntilEngineExit(time.Duration(60)*time.Second, time.Duration(1)*time.Second)
	if err != nil {
		return err
	}
	cacheDir := getEngineCacheDir()
	err = os.RemoveAll(cacheDir)
	if err != nil {
		return err
	}
	klog.Infof("UninstallEngine: cache dir %s  has been removed", cacheDir)

	return nil
}

func getEngineYaml(podManifestPath string) string {
	return filepath.Join(podManifestPath, enutil.EngineYamlName)
}

func getEngineConf() string {
	return filepath.Join(hubself.EngineRootDir, hubself.EngineName)
}

func getEngineCacheDir() string {
	// get default dir
	return disk.CacheBaseDir
}

func waitUntilEngineExit(timeout time.Duration, period time.Duration) error {
	klog.Info("wait for dcpsvr exit")
	serverHealthzURL, _ := url.Parse(fmt.Sprintf("http://%s", enutil.ServerHealthzServer))
	serverHealthzURL.Path = enutil.ServerHealthzURLPath

	return wait.PollImmediate(period, timeout, func() (bool, error) {
		_, err := pingClusterHealthz(http.DefaultClient, serverHealthzURL.String())
		if err != nil { // means Bhojpur DCP server engine has exited
			klog.Infof("Bhojpur DCP server engine is not running, with ping result: %v", err)
			return true, nil
		}
		klog.Infof("Bhojpur DCP server engine is still running")
		return false, nil
	})
}

// engineHealthcheck will check the status of Bhojpur DCP server engine pod
func engineHealthcheck(timeout time.Duration) error {
	serverHealthzURL, err := url.Parse(fmt.Sprintf("http://%s", enutil.ServerHealthzServer))
	if err != nil {
		return err
	}
	serverHealthzURL.Path = enutil.ServerHealthzURLPath

	start := time.Now()
	return wait.PollImmediate(hubHealthzCheckFrequency, timeout, func() (bool, error) {
		_, err := pingClusterHealthz(http.DefaultClient, serverHealthzURL.String())
		if err != nil {
			klog.Infof("Bhojpur DCP server engine is not ready, ping cluster healthz with result: %v", err)
			return false, nil
		}
		klog.Infof("Bhojpur DCP server engine healthz is OK after %f seconds", time.Since(start).Seconds())
		return true, nil
	})
}

func pingClusterHealthz(client *http.Client, addr string) (bool, error) {
	if client == nil {
		return false, fmt.Errorf("http client is invalid")
	}

	resp, err := client.Get(addr)
	if err != nil {
		return false, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return false, fmt.Errorf("failed to read response of cluster healthz, %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("response status code is %d", resp.StatusCode)
	}

	if strings.ToLower(string(b)) != "ok" {
		return false, fmt.Errorf("cluster healthz is %s", string(b))
	}

	return true, nil
}
