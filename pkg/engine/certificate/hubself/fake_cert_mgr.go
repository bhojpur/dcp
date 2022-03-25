package hubself

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
	"path/filepath"

	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/cmd/server/config"
	"github.com/bhojpur/dcp/pkg/engine/certificate/interfaces"
	"github.com/bhojpur/dcp/pkg/engine/storage/disk"
	"github.com/bhojpur/dcp/pkg/projectinfo"
)

var (
	defaultCertificatePEM = `-----BEGIN CERTIFICATE-----
MIICRzCCAfGgAwIBAgIJALMb7ecMIk3MMA0GCSqGSIb3DQEBCwUAMH4xCzAJBgNV
BAYTAkdCMQ8wDQYDVQQIDAZMb25kb24xDzANBgNVBAcMBkxvbmRvbjEYMBYGA1UE
CgwPR2xvYmFsIFNlY3VyaXR5MRYwFAYDVQQLDA1JVCBEZXBhcnRtZW50MRswGQYD
VQQDDBJ0ZXN0LWNlcnRpZmljYXRlLTAwIBcNMTcwNDI2MjMyNjUyWhgPMjExNzA0
MDIyMzI2NTJaMH4xCzAJBgNVBAYTAkdCMQ8wDQYDVQQIDAZMb25kb24xDzANBgNV
BAcMBkxvbmRvbjEYMBYGA1UECgwPR2xvYmFsIFNlY3VyaXR5MRYwFAYDVQQLDA1J
VCBEZXBhcnRtZW50MRswGQYDVQQDDBJ0ZXN0LWNlcnRpZmljYXRlLTAwXDANBgkq
hkiG9w0BAQEFAANLADBIAkEAtBMa7NWpv3BVlKTCPGO/LEsguKqWHBtKzweMY2CV
tAL1rQm913huhxF9w+ai76KQ3MHK5IVnLJjYYA5MzP2H5QIDAQABo1AwTjAdBgNV
HQ4EFgQU22iy8aWkNSxv0nBxFxerfsvnZVMwHwYDVR0jBBgwFoAU22iy8aWkNSxv
0nBxFxerfsvnZVMwDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQsFAANBAEOefGbV
NcHxklaW06w6OBYJPwpIhCVozC1qdxGX1dg8VkEKzjOzjgqVD30m59OFmSlBmHsl
nkVA6wyOSDYBf3o=
-----END CERTIFICATE-----`
	defaultKeyPEM = `-----BEGIN RSA PRIVATE KEY-----
MIIBUwIBADANBgkqhkiG9w0BAQEFAASCAT0wggE5AgEAAkEAtBMa7NWpv3BVlKTC
PGO/LEsguKqWHBtKzweMY2CVtAL1rQm913huhxF9w+ai76KQ3MHK5IVnLJjYYA5M
zP2H5QIDAQABAkAS9BfXab3OKpK3bIgNNyp+DQJKrZnTJ4Q+OjsqkpXvNltPJosf
G8GsiKu/vAt4HGqI3eU77NvRI+mL4MnHRmXBAiEA3qM4FAtKSRBbcJzPxxLEUSwg
XSCcosCktbkXvpYrS30CIQDPDxgqlwDEJQ0uKuHkZI38/SPWWqfUmkecwlbpXABK
iQIgZX08DA8VfvcA5/Xj1Zjdey9FVY6POLXen6RPiabE97UCICp6eUW7ht+2jjar
e35EltCRCjoejRHTuN9TC0uCoVipAiAXaJIx/Q47vGwiw6Y8KXsNU6y54gTbOSxX
54LzHNk/+Q==
-----END RSA PRIVATE KEY-----`
)

type fakeEngineCertManager struct {
	certificatePEM   string
	keyPEM           string
	rootDir          string
	engineName       string
	engineConfigFile string
}

// NewFakeEngineCertManager new a EngineCertificateManager instance
func NewFakeEngineCertManager(rootDir, engineConfigFile, certificatePEM, keyPEM string) (interfaces.EngineCertificateManager, error) {
	hn := projectinfo.GetEngineName()
	if len(hn) == 0 {
		hn = EngineName
	}
	if len(certificatePEM) == 0 {
		certificatePEM = defaultCertificatePEM
	}
	if len(keyPEM) == 0 {
		keyPEM = defaultKeyPEM
	}

	rd := rootDir
	if len(rd) == 0 {
		rd = filepath.Join(EngineRootDir, hn)
	}

	fyc := &fakeEngineCertManager{
		certificatePEM:   certificatePEM,
		keyPEM:           keyPEM,
		rootDir:          rd,
		engineName:       hn,
		engineConfigFile: engineConfigFile,
	}

	return fyc, nil
}

// Start create the dcpsvr.conf file
func (fyc *fakeEngineCertManager) Start() {
	dStorage, err := disk.NewDiskStorage(fyc.rootDir)
	if err != nil {
		klog.Errorf("failed to create storage, %v", err)
	}
	fileName := fmt.Sprintf(engineConfigFileName, fyc.engineName)
	engineConf := filepath.Join(fyc.rootDir, fileName)
	if err := dStorage.Create(fileName, []byte(fyc.engineConfigFile)); err != nil {
		klog.Errorf("Unable to create the file %q: %v", engineConf, err)
	}
	return
}

// Stop do nothing
func (fyc *fakeEngineCertManager) Stop() {}

// Current returns the certificate created by the entered fyc.certificatePEM and fyc.keyPEM
func (fyc *fakeEngineCertManager) Current() *tls.Certificate {
	certificate, err := tls.X509KeyPair([]byte(fyc.certificatePEM), []byte(fyc.keyPEM))
	if err != nil {
		panic(fmt.Sprintf("Unable to initialize certificate: %v", err))
	}
	certs, err := x509.ParseCertificates(certificate.Certificate[0])
	if err != nil {
		panic(fmt.Sprintf("Unable to initialize certificate leaf: %v", err))
	}
	certificate.Leaf = certs[0]

	return &certificate
}

// ServerHealthy returns true
func (fyc *fakeEngineCertManager) ServerHealthy() bool {
	return true
}

// Update do nothing
func (fyc *fakeEngineCertManager) Update(_ *config.EngineConfiguration) error {
	return nil
}

// GetCaFile returns the empty path
func (fyc *fakeEngineCertManager) GetCaFile() string {
	return ""
}

// GetConfFilePath returns the path of Bhojpur DCP config file path
func (fyc *fakeEngineCertManager) GetConfFilePath() string {
	return fyc.getEngineConfFile()
}

// NotExpired returns true
func (fyc *fakeEngineCertManager) NotExpired() bool {
	return fyc.Current() != nil
}

// getEngineConfFile returns the path of Bhojpur DCP agent conf file.
func (fyc *fakeEngineCertManager) getEngineConfFile() string {
	return filepath.Join(fyc.rootDir, fmt.Sprintf(engineConfigFileName, fyc.engineName))
}
