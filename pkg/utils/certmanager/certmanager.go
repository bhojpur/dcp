package certmanager

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
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"net"
	"os"
	"time"

	certificatesv1 "k8s.io/api/certificates/v1"
	certificatesv1beta1 "k8s.io/api/certificates/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/certificate"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/pkg/projectinfo"
	"github.com/bhojpur/dcp/pkg/tunnel/constants"
	"github.com/bhojpur/dcp/pkg/tunnel/server/serveraddr"
	"github.com/bhojpur/dcp/pkg/utils/certmanager/store"
)

const (
	EngineServerCSROrg = "system:masters"
	EngineCSROrg       = "bhojpur:dcpsvr"
	EngineServerCSRCN  = "kube-apiserver-kubelet-client"
)

// NewTunnelServerCertManager creates a certificate manager for
// the tunnel-server
func NewTunnelServerCertManager(
	clientset kubernetes.Interface,
	factory informers.SharedInformerFactory,
	certDir string,
	clCertNames []string,
	clIPs []net.IP,
	stopCh <-chan struct{}) (certificate.Manager, error) {
	var (
		dnsNames = []string{}
		ips      = []net.IP{}
		err      error
	)

	// the ips and dnsNames should be acquired through api-server at the first time, because the informer factory has not started yet.
	_ = wait.PollUntil(5*time.Second, func() (bool, error) {
		dnsNames, ips, err = serveraddr.GetTunnelServerDNSandIP(clientset)
		if err != nil {
			return false, err
		}

		// get clusterIP for tunnel server internal service
		svc, err := clientset.CoreV1().Services(constants.TunnelServerServiceNs).Get(context.Background(), constants.TunnelServerInternalServiceName, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			// compatible with versions that not supported x-tunnel-server-internal-svc
			return true, nil
		} else if err != nil {
			return false, err
		}

		if svc.Spec.ClusterIP != "" && net.ParseIP(svc.Spec.ClusterIP) != nil {
			ips = append(ips, net.ParseIP(svc.Spec.ClusterIP))
			dnsNames = append(dnsNames, serveraddr.GetDefaultDomainsForSvc(svc.Namespace, svc.Name)...)
		}

		return true, nil
	}, stopCh)
	// add user specified DNS names and IP addresses
	dnsNames = append(dnsNames, clCertNames...)
	ips = append(ips, clIPs...)
	klog.Infof("subject of tunnel server certificate, ips=%#+v, dnsNames=%#+v", ips, dnsNames)

	// the dynamic ip acquire func
	getIPs := func() ([]net.IP, error) {
		_, dynamicIPs, err := serveraddr.TunnelServerAddrManager(factory)
		dynamicIPs = append(dynamicIPs, clIPs...)
		return dynamicIPs, err
	}

	return newCertManager(
		clientset,
		projectinfo.GetServerName(),
		certDir,
		constants.TunneServerCSRCN,
		[]string{constants.TunneServerCSROrg, constants.TunnelCSROrg},
		dnsNames,
		[]certificatesv1.KeyUsage{
			certificatesv1.UsageKeyEncipherment,
			certificatesv1.UsageDigitalSignature,
			certificatesv1.UsageServerAuth,
			certificatesv1.UsageClientAuth,
		},
		ips,
		getIPs)
}

// NewTunnelAgentCertManager creates a certificate manager for
// the tunel-agent
func NewTunnelAgentCertManager(
	clientset kubernetes.Interface,
	certDir string) (certificate.Manager, error) {
	// As tunnel-agent will run on the edge node with Host network mode,
	// we can use the status.podIP as the node IP
	nodeIP := os.Getenv(constants.TunnelAgentPodIPEnv)
	if nodeIP == "" {
		return nil, fmt.Errorf("env %s is not set",
			constants.TunnelAgentPodIPEnv)
	}

	return newCertManager(
		clientset,
		projectinfo.GetAgentName(),
		certDir,
		constants.TunnelAgentCSRCN,
		[]string{constants.TunnelCSROrg},
		[]string{os.Getenv("NODE_NAME")},
		[]certificatesv1.KeyUsage{
			certificatesv1.UsageKeyEncipherment,
			certificatesv1.UsageDigitalSignature,
			certificatesv1.UsageClientAuth,
		},
		[]net.IP{net.ParseIP(nodeIP)},
		nil)
}

// NewEngineServerCertManager creates a certificate manager for
// the dcpsvr-server
func NewEngineServerCertManager(
	clientset kubernetes.Interface,
	certDir,
	proxyServerSecureDummyAddr string) (certificate.Manager, error) {

	klog.Infof("subject of Bhojpur DCP server engine certificate")
	host, _, err := net.SplitHostPort(proxyServerSecureDummyAddr)
	if err != nil {
		return nil, err
	}

	return newCertManager(
		clientset,
		fmt.Sprintf("%s-server", projectinfo.GetEngineName()),
		certDir,
		EngineServerCSRCN,
		[]string{EngineServerCSROrg, EngineCSROrg},
		nil,
		[]certificatesv1.KeyUsage{
			certificatesv1.UsageKeyEncipherment,
			certificatesv1.UsageDigitalSignature,
			certificatesv1.UsageServerAuth,
		},
		[]net.IP{net.ParseIP("127.0.0.1"), net.ParseIP(host)},
		nil)
}

// NewCertManager creates a certificate manager that will generates a
// certificate by sending a csr to the apiserver
func newCertManager(
	clientset kubernetes.Interface,
	componentName,
	certDir,
	commonName string,
	organizations,
	dnsNames []string,
	keyUsages []certificatesv1.KeyUsage,
	ips []net.IP,
	getIPs serveraddr.GetIPs) (certificate.Manager, error) {
	certificateStore, err :=
		store.NewFileStoreWrapper(componentName, certDir, certDir, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the server certificate store: %v", err)
	}

	getTemplate := func() *x509.CertificateRequest {
		// use dynamic ips
		if getIPs != nil {
			tmpIPs, err := getIPs()
			if err == nil && len(tmpIPs) != 0 {
				klog.V(4).Infof("the latest tunnel server's ips=%#+v", tmpIPs)
				ips = tmpIPs
			}
		}
		return &x509.CertificateRequest{
			Subject: pkix.Name{
				CommonName:   commonName,
				Organization: organizations,
			},
			DNSNames:    dnsNames,
			IPAddresses: ips,
		}
	}

	certManager, err := certificate.NewManager(&certificate.Config{
		ClientsetFn: func(current *tls.Certificate) (kubernetes.Interface, error) {
			return clientset, nil
		},
		SignerName:       certificatesv1beta1.LegacyUnknownSignerName,
		GetTemplate:      getTemplate,
		Usages:           keyUsages,
		CertificateStore: certificateStore,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize server certificate manager: %v", err)
	}

	return certManager, nil
}
