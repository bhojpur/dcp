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
	"bytes"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/client-go/rest"

	"github.com/bhojpur/dcp/cmd/server/config"
	"github.com/bhojpur/dcp/pkg/engine/certificate/hubself"
	"github.com/bhojpur/dcp/pkg/engine/certificate/interfaces"
	"github.com/bhojpur/dcp/pkg/engine/healthchecker"
	"github.com/bhojpur/dcp/pkg/engine/storage/disk"
)

var (
	certificatePEM = []byte(`-----BEGIN CERTIFICATE-----
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
-----END CERTIFICATE-----`)
	keyPEM = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIBUwIBADANBgkqhkiG9w0BAQEFAASCAT0wggE5AgEAAkEAtBMa7NWpv3BVlKTC
PGO/LEsguKqWHBtKzweMY2CVtAL1rQm913huhxF9w+ai76KQ3MHK5IVnLJjYYA5M
zP2H5QIDAQABAkAS9BfXab3OKpK3bIgNNyp+DQJKrZnTJ4Q+OjsqkpXvNltPJosf
G8GsiKu/vAt4HGqI3eU77NvRI+mL4MnHRmXBAiEA3qM4FAtKSRBbcJzPxxLEUSwg
XSCcosCktbkXvpYrS30CIQDPDxgqlwDEJQ0uKuHkZI38/SPWWqfUmkecwlbpXABK
iQIgZX08DA8VfvcA5/Xj1Zjdey9FVY6POLXen6RPiabE97UCICp6eUW7ht+2jjar
e35EltCRCjoejRHTuN9TC0uCoVipAiAXaJIx/Q47vGwiw6Y8KXsNU6y54gTbOSxX
54LzHNk/+Q==
-----END RSA PRIVATE KEY-----`)
	engineCon = `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: temp
    server: https://10.10.10.113:6443
  name: default-cluster
contexts:
- context:
    cluster: default-cluster
    namespace: default
    user: default-auth
  name: default-context
current-context: default-context
kind: Config
preferences: {}
users:
- name: default-auth
  user:
    client-certificate: /tmp/pki/dcpsvr-current.pem
    client-key: /tmp/pki/dcpsvr-current.pem
`
	testDir = "/tmp/pki/"
)

func TestGetRestConfig(t *testing.T) {
	remoteServers := map[string]int{"https://10.10.10.113:6443": 2}
	u, _ := url.Parse("https://10.10.10.113:6443")
	fakeHealthchecker := healthchecker.NewFakeChecker(true, remoteServers)
	dStorage, err := disk.NewDiskStorage(testDir)
	defer func() {
		if err := os.RemoveAll(testDir); err != nil {
			t.Errorf("Unable to clean up test directory %q: %v", testDir, err)
		}
	}()

	// store the kubelet ca file
	caFile := filepath.Join(testDir, "ca.crt")
	if err := dStorage.Create("ca.crt", certificatePEM); err != nil {
		t.Fatalf("Unable to create the file %q: %v", caFile, err)
	}

	// store the kubelet-pair.pem file
	pairFile := filepath.Join(testDir, "kubelet-pair.pem")
	pd := bytes.Join([][]byte{certificatePEM, keyPEM}, []byte("\n"))
	if err := dStorage.Create("kubelet-pair.pem", pd); err != nil {
		t.Fatalf("Unable to create the file %q: %v", pairFile, err)
	}

	// store the dcpsvr-current.pem
	engineCurrent := filepath.Join(testDir, "dcpsvr-current.pem")
	if err := dStorage.Create("dcpsvr-current.pem", pd); err != nil {
		t.Fatalf("Unable to create the file %q: %v", engineCurrent, err)
	}

	// set the EngineConfiguration
	cfg := &config.EngineConfiguration{
		RootDir:               testDir,
		RemoteServers:         []*url.URL{u},
		KubeletRootCAFilePath: caFile,
		KubeletPairFilePath:   pairFile,
	}

	tests := []struct {
		desc string
		mode string
	}{
		{desc: "hubself mode", mode: "hubself"},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			cfg.CertMgrMode = tt.mode
			var certMgr interfaces.EngineCertificateManager
			if tt.mode == "hubself" {
				certMgr, err = hubself.NewFakeEngineCertManager(testDir, engineCon, string(certificatePEM), string(keyPEM))
				certMgr.Start()
			}

			if err != nil {
				t.Errorf("failed to create %s certManager: %v", tt.mode, err)
			}

			rcm, err := NewRestConfigManager(cfg, certMgr, fakeHealthchecker)
			if err != nil {
				t.Errorf("failed to create RestConfigManager: %v", err)
			}

			var rc *rest.Config
			rc = rcm.GetRestConfig(true)
			if tt.mode == "hubself" {
				if rc.Host != u.String() || rc.TLSClientConfig.CertFile != engineCurrent || rc.TLSClientConfig.KeyFile != engineCurrent {
					t.Errorf("The information in rest.Config is not correct: %s", tt.mode)
				}
			} else {
				if rc.Host != u.String() || rc.TLSClientConfig.CAFile != caFile || rc.TLSClientConfig.KeyFile != pairFile {
					t.Errorf("The information in rest.Config is not correct: %s", tt.mode)
				}
			}
		})
	}
}
