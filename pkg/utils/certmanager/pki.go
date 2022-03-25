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
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/certificate"
)

// GenTLSConfigUseCertMgrAndCertPool generates a TLS configuration
// using the given certificate manager and x509 CertPool
func GenTLSConfigUseCertMgrAndCertPool(
	m certificate.Manager,
	root *x509.CertPool) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		// Can't use SSLv3 because of POODLE and BEAST
		// Can't use TLSv1.0 because of POODLE and BEAST using CBC cipher
		// Can't use TLSv1.1 because of RC4 cipher usage
		MinVersion: tls.VersionTLS12,
		ClientCAs:  root,
		ClientAuth: tls.VerifyClientCertIfGiven,
	}

	tlsConfig.GetClientCertificate =
		func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			cert := m.Current()
			if cert == nil {
				return &tls.Certificate{Certificate: nil}, nil
			}
			return cert, nil
		}
	tlsConfig.GetCertificate =
		func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
			cert := m.Current()
			if cert == nil {
				return &tls.Certificate{Certificate: nil}, nil
			}
			return cert, nil
		}

	return tlsConfig, nil
}

// GenRootCertPool generates a x509 CertPool based on the given kubeconfig,
// if the kubeConfig is empty, it will creates the CertPool using the CA file
func GenRootCertPool(kubeConfig, caFile string) (*x509.CertPool, error) {
	if kubeConfig != "" {
		// kubeconfig is given, generate the clientset based on it
		if _, err := os.Stat(kubeConfig); os.IsNotExist(err) {
			return nil, err
		}

		// load the root ca from the given kubeconfig file
		config, err := clientcmd.LoadFromFile(kubeConfig)
		if err != nil || config == nil {
			return nil, fmt.Errorf("failed to load the kubeconfig file(%s), %v",
				kubeConfig, err)
		}

		if len(config.CurrentContext) == 0 {
			return nil, fmt.Errorf("'current context' is not set in %s",
				kubeConfig)
		}

		ctx, ok := config.Contexts[config.CurrentContext]
		if !ok || ctx == nil {
			return nil, fmt.Errorf("'current context(%s)' is not found in %s",
				config.CurrentContext, kubeConfig)
		}

		cluster, ok := config.Clusters[ctx.Cluster]
		if !ok || cluster == nil {
			return nil, fmt.Errorf("'cluster(%s)' is not found in %s",
				ctx.Cluster, kubeConfig)
		}

		if len(cluster.CertificateAuthorityData) == 0 {
			return nil, fmt.Errorf("'certificate authority data of the cluster(%s) is not set in %s",
				ctx.Cluster, kubeConfig)
		}

		rootCertPool := x509.NewCertPool()
		rootCertPool.AppendCertsFromPEM(cluster.CertificateAuthorityData)
		return rootCertPool, nil
	}

	// kubeConfig is missing, generate the cluster root ca based on the given ca file
	return GenCertPoolUseCA(caFile)
}

// GenTLSConfigUseCertMgrAndCA generates a TLS configuration based on the
// given certificate manager and the CA file
func GenTLSConfigUseCertMgrAndCA(
	m certificate.Manager,
	serverAddr, caFile string) (*tls.Config, error) {
	root, err := GenCertPoolUseCA(caFile)
	if err != nil {
		return nil, err
	}

	host, _, err := net.SplitHostPort(serverAddr)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		// Can't use SSLv3 because of POODLE and BEAST
		// Can't use TLSv1.0 because of POODLE and BEAST using CBC cipher
		// Can't use TLSv1.1 because of RC4 cipher usage
		MinVersion: tls.VersionTLS12,
		ServerName: host,
		RootCAs:    root,
	}

	tlsConfig.GetClientCertificate =
		func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			cert := m.Current()
			if cert == nil {
				return &tls.Certificate{Certificate: nil}, nil
			}
			return cert, nil
		}
	tlsConfig.GetCertificate =
		func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
			cert := m.Current()
			if cert == nil {
				return &tls.Certificate{Certificate: nil}, nil
			}
			return cert, nil
		}

	return tlsConfig, nil
}

// GenCertPoolUseCA generates a x509 CertPool based on the given CA file
func GenCertPoolUseCA(caFile string) (*x509.CertPool, error) {
	if caFile == "" {
		return nil, errors.New("CA file is not set")
	}

	if _, err := os.Stat(caFile); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("CA file(%s) doesn't exist", caFile)
		}
		return nil, fmt.Errorf("fail to stat the CA file(%s): %s", caFile, err)
	}

	caData, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caData)
	return certPool, nil
}
