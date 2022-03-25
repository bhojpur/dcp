package transport

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
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/pkg/engine/certificate/interfaces"
	"github.com/bhojpur/dcp/pkg/engine/util"
)

// Interface is an transport interface for managing clients that used to connecting kube-apiserver
type Interface interface {
	// concurrent use by multiple goroutines
	// CurrentTransport get transport that used by load balancer
	CurrentTransport() http.RoundTripper
	// BearerTransport returns transport for proxying request with bearer token in header
	BearerTransport() http.RoundTripper
	// close all net connections that specified by address
	Close(address string)
}

type transportManager struct {
	currentTransport *http.Transport
	bearerTransport  *http.Transport
	certManager      interfaces.EngineCertificateManager
	closeAll         func()
	close            func(string)
	stopCh           <-chan struct{}
}

// NewTransportManager create an transport interface object.
func NewTransportManager(certMgr interfaces.EngineCertificateManager, stopCh <-chan struct{}) (Interface, error) {
	caFile := certMgr.GetCaFile()
	if len(caFile) == 0 {
		return nil, fmt.Errorf("ca cert file was not prepared when new tranport")
	}
	klog.V(2).Infof("use %s ca cert file to access remote server", caFile)

	cfg, err := tlsConfig(certMgr, caFile)
	if err != nil {
		klog.Errorf("could not get tls config when new transport, %v", err)
		return nil, err
	}

	d := util.NewDialer("transport manager")
	t := utilnet.SetTransportDefaults(&http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     cfg,
		MaxIdleConnsPerHost: 25,
		DialContext:         d.DialContext,
	})

	bearerTLSCfg, err := tlsConfig(nil, caFile)
	if err != nil {
		klog.Errorf("could not get tls config when new bearer transport, %v", err)
		return nil, err
	}

	bt := utilnet.SetTransportDefaults(&http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     bearerTLSCfg,
		MaxIdleConnsPerHost: 25,
		DialContext:         d.DialContext,
	})

	tm := &transportManager{
		currentTransport: t,
		bearerTransport:  bt,
		certManager:      certMgr,
		closeAll:         d.CloseAll,
		close:            d.Close,
		stopCh:           stopCh,
	}
	tm.start()

	return tm, nil
}

func (tm *transportManager) CurrentTransport() http.RoundTripper {
	return tm.currentTransport
}

func (tm *transportManager) BearerTransport() http.RoundTripper {
	return tm.bearerTransport
}

func (tm *transportManager) Close(address string) {
	tm.close(address)
}

func (tm *transportManager) start() {
	lastCert := tm.certManager.Current()

	go wait.Until(func() {
		curr := tm.certManager.Current()

		if lastCert == nil && curr == nil {
			// maybe at Bhojpur DCP engine startup, just wait for cert generated, do nothing
		} else if lastCert == nil && curr != nil {
			// cert generated, close all client connections for load new cert
			klog.Infof("new cert generated, so close all client connections for loading new cert")
			tm.closeAll()
			lastCert = curr
		} else if lastCert != nil && curr != nil {
			if lastCert == curr {
				// cert is not rotate, just wait
			} else {
				// cert rotated
				klog.Infof("cert rotated, so close all client connections for loading new cert")
				tm.closeAll()
				lastCert = curr
			}
		} else {
			// lastCet != nil && curr == nil
			// certificate expired or deleted unintentionally, just wait for cert updated by bootstrap config, do nothing
		}
	}, 10*time.Second, tm.stopCh)
}

func tlsConfig(certMgr interfaces.EngineCertificateManager, caFile string) (*tls.Config, error) {
	root, err := rootCertPool(caFile)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		// Can't use SSLv3 because of POODLE and BEAST
		// Can't use TLSv1.0 because of POODLE and BEAST using CBC cipher
		// Can't use TLSv1.1 because of RC4 cipher usage
		MinVersion: tls.VersionTLS12,
		RootCAs:    root,
	}

	if certMgr != nil {
		tlsConfig.GetClientCertificate = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			cert := certMgr.Current()
			if cert == nil {
				return &tls.Certificate{Certificate: nil}, nil
			}
			return cert, nil
		}
	}

	return tlsConfig, nil
}

func rootCertPool(caFile string) (*x509.CertPool, error) {
	if len(caFile) > 0 {
		if caFileExists, err := util.FileExists(caFile); err != nil {
			return nil, err
		} else if caFileExists {
			caData, err := ioutil.ReadFile(caFile)
			if err != nil {
				return nil, err
			}

			certPool := x509.NewCertPool()
			certPool.AppendCertsFromPEM(caData)
			return certPool, nil
		}
	}

	return nil, fmt.Errorf("failed to load ca file(%s)", caFile)
}
