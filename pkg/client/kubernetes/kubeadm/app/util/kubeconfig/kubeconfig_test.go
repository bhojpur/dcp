package kubeconfig

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
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	configOut1 = `apiVersion: v1
clusters:
- cluster:
    server: ""
  name: k8s
contexts:
- context:
    cluster: k8s
    user: user1
  name: user1@k8s
current-context: user1@k8s
kind: Config
preferences: {}
users:
- name: user1
  user:
    token: abc
`
	configOut2 = `apiVersion: v1
clusters:
- cluster:
    server: localhost:8080
  name: kubernetes
contexts:
- context:
    cluster: kubernetes
    user: user2
  name: user2@kubernetes
current-context: user2@kubernetes
kind: Config
preferences: {}
users:
- name: user2
  user:
    token: cba
`
)

type configClient struct {
	clusterName string
	userName    string
	serverURL   string
	caCert      []byte
}

type configClientWithCerts struct {
	clientKey  []byte
	clientCert []byte
}

type configClientWithToken struct {
	token string
}

func TestCreateWithCerts(t *testing.T) {
	var createBasicTest = []struct {
		name        string
		cc          configClient
		ccWithCerts configClientWithCerts
		expected    string
	}{
		{"empty config", configClient{}, configClientWithCerts{}, ""},
		{"clusterName kubernetes", configClient{clusterName: "kubernetes"}, configClientWithCerts{}, ""},
	}
	for _, rt := range createBasicTest {
		t.Run(rt.name, func(t *testing.T) {
			cwc := CreateWithCerts(
				rt.cc.serverURL,
				rt.cc.clusterName,
				rt.cc.userName,
				rt.cc.caCert,
				rt.ccWithCerts.clientKey,
				rt.ccWithCerts.clientCert,
			)
			if cwc.Kind != rt.expected {
				t.Errorf(
					"failed CreateWithCerts:\n\texpected: %s\n\t  actual: %s",
					rt.expected,
					cwc.Kind,
				)
			}
		})
	}
}

func TestCreateWithToken(t *testing.T) {
	var createBasicTest = []struct {
		name        string
		cc          configClient
		ccWithToken configClientWithToken
		expected    string
	}{
		{"empty config", configClient{}, configClientWithToken{}, ""},
		{"clusterName kubernetes", configClient{clusterName: "kubernetes"}, configClientWithToken{}, ""},
	}
	for _, rt := range createBasicTest {
		t.Run(rt.name, func(t *testing.T) {
			cwc := CreateWithToken(
				rt.cc.serverURL,
				rt.cc.clusterName,
				rt.cc.userName,
				rt.cc.caCert,
				rt.ccWithToken.token,
			)
			if cwc.Kind != rt.expected {
				t.Errorf(
					"failed CreateWithToken:\n\texpected: %s\n\t  actual: %s",
					rt.expected,
					cwc.Kind,
				)
			}
		})
	}
}

func TestWriteKubeconfigToDisk(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Couldn't create tmpdir")
	}
	defer os.RemoveAll(tmpdir)

	var writeConfig = []struct {
		name        string
		cc          configClient
		ccWithToken configClientWithToken
		expected    error
		file        []byte
	}{
		{"test1", configClient{clusterName: "k8s", userName: "user1"}, configClientWithToken{token: "abc"}, nil, []byte(configOut1)},
		{"test2", configClient{clusterName: "kubernetes", userName: "user2", serverURL: "localhost:8080"}, configClientWithToken{token: "cba"}, nil, []byte(configOut2)},
	}
	for _, rt := range writeConfig {
		t.Run(rt.name, func(t *testing.T) {
			c := CreateWithToken(
				rt.cc.serverURL,
				rt.cc.clusterName,
				rt.cc.userName,
				rt.cc.caCert,
				rt.ccWithToken.token,
			)
			configPath := fmt.Sprintf("%s/etc/kubernetes/%s.conf", tmpdir, rt.name)
			err := WriteToDisk(configPath, c)
			if err != rt.expected {
				t.Errorf(
					"failed WriteToDisk with an error:\n\texpected: %s\n\t  actual: %s",
					rt.expected,
					err,
				)
			}
			newFile, _ := ioutil.ReadFile(configPath)
			if !bytes.Equal(newFile, rt.file) {
				t.Errorf(
					"failed WriteToDisk config write:\n\texpected: %s\n\t  actual: %s",
					rt.file,
					newFile,
				)
			}
		})
	}
}

func TestGetCurrentAuthInfo(t *testing.T) {
	var testCases = []struct {
		name     string
		config   *clientcmdapi.Config
		expected bool
	}{
		{
			name:     "nil context",
			config:   nil,
			expected: false,
		},
		{
			name:     "no CurrentContext value",
			config:   &clientcmdapi.Config{},
			expected: false,
		},
		{
			name:     "no CurrentContext object",
			config:   &clientcmdapi.Config{CurrentContext: "kubernetes"},
			expected: false,
		},
		{
			name: "CurrentContext object with bad contents",
			config: &clientcmdapi.Config{
				CurrentContext: "kubernetes",
				Contexts:       map[string]*clientcmdapi.Context{"NOTkubernetes": {}},
			},
			expected: false,
		},
		{
			name: "no AuthInfo value",
			config: &clientcmdapi.Config{
				CurrentContext: "kubernetes",
				Contexts:       map[string]*clientcmdapi.Context{"kubernetes": {}},
			},
			expected: false,
		},
		{
			name: "no AuthInfo object",
			config: &clientcmdapi.Config{
				CurrentContext: "kubernetes",
				Contexts:       map[string]*clientcmdapi.Context{"kubernetes": {AuthInfo: "kubernetes"}},
			},
			expected: false,
		},
		{
			name: "AuthInfo object with bad contents",
			config: &clientcmdapi.Config{
				CurrentContext: "kubernetes",
				Contexts:       map[string]*clientcmdapi.Context{"kubernetes": {AuthInfo: "kubernetes"}},
				AuthInfos:      map[string]*clientcmdapi.AuthInfo{"NOTkubernetes": {}},
			},
			expected: false,
		},
		{
			name: "valid AuthInfo",
			config: &clientcmdapi.Config{
				CurrentContext: "kubernetes",
				Contexts:       map[string]*clientcmdapi.Context{"kubernetes": {AuthInfo: "kubernetes"}},
				AuthInfos:      map[string]*clientcmdapi.AuthInfo{"kubernetes": {}},
			},
			expected: true,
		},
	}
	for _, rt := range testCases {
		t.Run(rt.name, func(t *testing.T) {
			r := getCurrentAuthInfo(rt.config)
			if rt.expected != (r != nil) {
				t.Errorf(
					"failed TestHasCredentials:\n\texpected: %v\n\t  actual: %v",
					rt.expected,
					r,
				)
			}
		})
	}
}

func TestHasCredentials(t *testing.T) {
	var testCases = []struct {
		name     string
		config   *clientcmdapi.Config
		expected bool
	}{
		{
			name:     "no authInfo",
			config:   nil,
			expected: false,
		},
		{
			name: "no credentials",
			config: &clientcmdapi.Config{
				CurrentContext: "kubernetes",
				Contexts:       map[string]*clientcmdapi.Context{"kubernetes": {AuthInfo: "kubernetes"}},
				AuthInfos:      map[string]*clientcmdapi.AuthInfo{"kubernetes": {}},
			},
			expected: false,
		},
		{
			name: "token authentication credentials",
			config: &clientcmdapi.Config{
				CurrentContext: "kubernetes",
				Contexts:       map[string]*clientcmdapi.Context{"kubernetes": {AuthInfo: "kubernetes"}},
				AuthInfos:      map[string]*clientcmdapi.AuthInfo{"kubernetes": {Token: "123"}},
			},
			expected: true,
		},
		{
			name: "basic authentication credentials",
			config: &clientcmdapi.Config{
				CurrentContext: "kubernetes",
				Contexts:       map[string]*clientcmdapi.Context{"kubernetes": {AuthInfo: "kubernetes"}},
				AuthInfos:      map[string]*clientcmdapi.AuthInfo{"kubernetes": {Username: "A", Password: "B"}},
			},
			expected: true,
		},
		{
			name: "X509 authentication credentials",
			config: &clientcmdapi.Config{
				CurrentContext: "kubernetes",
				Contexts:       map[string]*clientcmdapi.Context{"kubernetes": {AuthInfo: "kubernetes"}},
				AuthInfos:      map[string]*clientcmdapi.AuthInfo{"kubernetes": {ClientKey: "A", ClientCertificate: "B"}},
			},
			expected: true,
		},
	}
	for _, rt := range testCases {
		t.Run(rt.name, func(t *testing.T) {
			r := HasAuthenticationCredentials(rt.config)
			if rt.expected != r {
				t.Errorf(
					"failed TestHasCredentials:\n\texpected: %v\n\t  actual: %v",
					rt.expected,
					r,
				)
			}
		})
	}
}
