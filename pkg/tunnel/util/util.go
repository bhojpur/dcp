package util

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
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/pkg/profile"
	"github.com/bhojpur/dcp/pkg/projectinfo"
)

const (
	// constants related dnat rules configmap
	TunnelServerDnatConfigMapNs    = "kube-system"
	dcpTunnelServerDnatDataKey     = "dnat-ports-pair"
	TunnelLocalHostProxyPorts      = "localhost-proxy-ports"
	dcpTunnelServerHTTPProxyPorts  = "http-proxy-ports"
	dcpTunnelServerHTTPSProxyPorts = "https-proxy-ports"
	PortsSeparator                 = ","
	PortPairSeparator              = "="

	KubeletHTTPSPort = "10250"
	KubeletHTTPPort  = "10255"

	MinPort = 1
	MaxPort = 65535
)

var (
	TunnelServerDnatConfigMapName = fmt.Sprintf("%s-tunnel-server-cfg",
		strings.TrimRightFunc(projectinfo.GetProjectPrefix(), func(c rune) bool { return c == '-' }))
)

// RunMetaServer start a http server for serving metrics and pprof requests.
func RunMetaServer(addr string) {
	muxHandler := mux.NewRouter()
	muxHandler.Handle("/metrics", promhttp.Handler())

	// register handler for pprof
	profile.Install(muxHandler)

	metaServer := &http.Server{
		Addr:           addr,
		Handler:        muxHandler,
		MaxHeaderBytes: 1 << 20,
	}

	klog.InfoS("start handling meta requests(metrics/pprof)", "server endpoint", addr)
	go func() {
		err := metaServer.ListenAndServe()
		if err != nil {
			klog.ErrorS(err, "meta server could not listen")
		}
		klog.InfoS("meta server stopped listening", "server endpoint", addr)
	}()
}

// GetConfiguredProxyPortsAndMappings returns the proxy ports and mappings that configured for tunnel server.
// field dnat-ports-pair will be deprecated in future version. it's recommended to use
// field http-proxy-ports and https-proxy-ports.
func GetConfiguredProxyPortsAndMappings(client clientset.Interface, insecureListenAddr, secureListenAddr string) ([]string, map[string]string, error) {
	c, err := client.CoreV1().
		ConfigMaps(TunnelServerDnatConfigMapNs).
		Get(context.Background(), TunnelServerDnatConfigMapName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return []string{}, map[string]string{}, fmt.Errorf("configmap %s/%s is not found",
				TunnelServerDnatConfigMapNs,
				TunnelServerDnatConfigMapName)
		} else {
			return []string{}, map[string]string{}, fmt.Errorf("fail to get configmap %s/%s: %v",
				TunnelServerDnatConfigMapNs,
				TunnelServerDnatConfigMapName, err)
		}
	}

	return resolveProxyPortsAndMappings(c, insecureListenAddr, secureListenAddr)
}

// resolveProxyPortsAndMappings get proxy ports and port mappings from specified configmap
func resolveProxyPortsAndMappings(cm *v1.ConfigMap, insecureListenAddr, secureListenAddr string) ([]string, map[string]string, error) {
	portMappings := make(map[string]string)
	proxyPorts := make([]string, 0)

	_, insecurePort, err := net.SplitHostPort(insecureListenAddr)
	if err != nil {
		return proxyPorts, portMappings, err
	}

	// field dnat-ports-pair will be deprecated in future version
	for _, port := range resolvePorts(cm.Data[dcpTunnelServerDnatDataKey], insecurePort) {
		portMappings[port] = insecureListenAddr
	}

	// resolve http-proxy-port field
	for _, port := range resolvePorts(cm.Data[dcpTunnelServerHTTPProxyPorts], "") {
		portMappings[port] = insecureListenAddr
	}

	// resolve https-proxy-port field
	for _, port := range resolvePorts(cm.Data[dcpTunnelServerHTTPSProxyPorts], "") {
		portMappings[port] = secureListenAddr
	}

	// cleanup 10250/10255 mappings
	delete(portMappings, KubeletHTTPSPort)
	delete(portMappings, KubeletHTTPPort)

	for port := range portMappings {
		proxyPorts = append(proxyPorts, port)
	}

	return proxyPorts, portMappings, nil
}

// resolvePorts parse the specified ports setting and return ports slice.
func resolvePorts(portsStr, insecurePort string) []string {
	ports := make([]string, 0)
	if len(strings.TrimSpace(portsStr)) == 0 {
		return ports
	}

	isPortPair := strings.Contains(portsStr, PortPairSeparator)
	parts := strings.Split(portsStr, PortsSeparator)
	for _, port := range parts {
		var proxyPort string
		if isPortPair {
			subParts := strings.Split(port, PortPairSeparator)
			if len(subParts) == 2 && strings.TrimSpace(subParts[1]) == insecurePort {
				proxyPort = strings.TrimSpace(subParts[0])
			}
		} else {
			proxyPort = strings.TrimSpace(port)
		}

		if len(proxyPort) != 0 {
			portInt, err := strconv.Atoi(proxyPort)
			if err != nil {
				klog.Errorf("failed to parse port %s, %v", port, err)
				continue
			} else if portInt < MinPort || portInt > MaxPort {
				klog.Errorf("port %s is not invalid port(should be range 1~65535)", port)
				continue
			}

			ports = append(ports, proxyPort)
		}
	}

	return ports
}
