package rest

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
	"net/url"

	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/cmd/server/config"
	"github.com/bhojpur/dcp/pkg/engine/certificate/interfaces"
	"github.com/bhojpur/dcp/pkg/engine/healthchecker"
	"github.com/bhojpur/dcp/pkg/engine/util"
	"github.com/bhojpur/dcp/pkg/projectinfo"
)

type RestConfigManager struct {
	remoteServers []*url.URL
	certMgrMode   string
	checker       healthchecker.HealthChecker
	certManager   interfaces.EngineCertificateManager
}

// NewRestConfigManager creates a *RestConfigManager object
func NewRestConfigManager(cfg *config.EngineConfiguration, certMgr interfaces.EngineCertificateManager, healthChecker healthchecker.HealthChecker) (*RestConfigManager, error) {
	mgr := &RestConfigManager{
		remoteServers: cfg.RemoteServers,
		certMgrMode:   cfg.CertMgrMode,
		checker:       healthChecker,
		certManager:   certMgr,
	}
	return mgr, nil
}

// GetRestConfig gets rest client config according to the mode of certificateManager
func (rcm *RestConfigManager) GetRestConfig(needHealthyServer bool) *rest.Config {
	certMgrMode := rcm.certMgrMode
	switch certMgrMode {
	case util.EngineCertificateManagerName:
		return rcm.getHubselfRestConfig(needHealthyServer)
	default:
		return nil
	}
}

// getHubselfRestConfig gets rest client config from hub agent conf file.
func (rcm *RestConfigManager) getHubselfRestConfig(needHealthyServer bool) *rest.Config {
	healthyServer := rcm.remoteServers[0]
	if needHealthyServer {
		healthyServer = rcm.getHealthyServer()
		if healthyServer == nil {
			klog.Infof("all of remote servers are unhealthy, so return nil for rest config")
			return nil
		}
	}

	// certificate expired, rest config can not be used to connect remote server,
	// so return nil for rest config
	if rcm.certManager.Current() == nil {
		klog.Infof("certificate expired, so return nil for rest config")
		return nil
	}

	hubConfFile := rcm.certManager.GetConfFilePath()
	if isExist, _ := util.FileExists(hubConfFile); isExist {
		cfg, err := util.LoadRESTClientConfig(hubConfFile)
		if err != nil {
			klog.Errorf("could not get rest config for %s, %v", hubConfFile, err)
			return nil
		}

		// re-fix host connecting healthy server
		cfg.Host = healthyServer.String()
		klog.Infof("re-fix hub rest config host successfully with server %s", cfg.Host)
		return cfg
	}

	klog.Errorf("%s config file(%s) is not exist", projectinfo.GetEngineName(), hubConfFile)
	return nil
}

// getHealthyServer is used to get a healthy server
func (rcm *RestConfigManager) getHealthyServer() *url.URL {
	for _, server := range rcm.remoteServers {
		if rcm.checker.IsHealthy(server) {
			return server
		}
	}
	return nil
}
