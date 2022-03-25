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
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// CreateBasic creates a basic, general KubeConfig object that then can be extended
func CreateBasic(serverURL, clusterName, userName string, caCert []byte) *clientcmdapi.Config {
	// Use the cluster and the username as the context name
	contextName := fmt.Sprintf("%s@%s", userName, clusterName)

	return &clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			clusterName: {
				Server:                   serverURL,
				CertificateAuthorityData: caCert,
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			contextName: {
				Cluster:  clusterName,
				AuthInfo: userName,
			},
		},
		AuthInfos:      map[string]*clientcmdapi.AuthInfo{},
		CurrentContext: contextName,
	}
}

// CreateWithCerts creates a KubeConfig object with access to the API server with client certificates
func CreateWithCerts(serverURL, clusterName, userName string, caCert []byte, clientKey []byte, clientCert []byte) *clientcmdapi.Config {
	config := CreateBasic(serverURL, clusterName, userName, caCert)
	config.AuthInfos[userName] = &clientcmdapi.AuthInfo{
		ClientKeyData:         clientKey,
		ClientCertificateData: clientCert,
	}
	return config
}

// CreateWithToken creates a KubeConfig object with access to the API server with a token
func CreateWithToken(serverURL, clusterName, userName string, caCert []byte, token string) *clientcmdapi.Config {
	config := CreateBasic(serverURL, clusterName, userName, caCert)
	config.AuthInfos[userName] = &clientcmdapi.AuthInfo{
		Token: token,
	}
	return config
}

// ClientSetFromFile returns a ready-to-use client from a kubeconfig file
func ClientSetFromFile(path string) (*clientset.Clientset, error) {
	config, err := clientcmd.LoadFromFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load admin kubeconfig")
	}
	return ToClientSet(config)
}

// ToClientSet converts a KubeConfig object to a client
func ToClientSet(config *clientcmdapi.Config) (*clientset.Clientset, error) {
	overrides := clientcmd.ConfigOverrides{Timeout: "10s"}
	clientConfig, err := clientcmd.NewDefaultClientConfig(*config, &overrides).ClientConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create API client configuration from kubeconfig")
	}

	client, err := clientset.NewForConfig(clientConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create API client")
	}
	return client, nil
}

// WriteToDisk writes a KubeConfig object down to disk with mode 0600
func WriteToDisk(filename string, kubeconfig *clientcmdapi.Config) error {
	err := clientcmd.WriteToFile(*kubeconfig, filename)
	if err != nil {
		return err
	}

	return nil
}

// GetClusterFromKubeConfig returns the default Cluster of the specified KubeConfig
func GetClusterFromKubeConfig(config *clientcmdapi.Config) *clientcmdapi.Cluster {
	// If there is an unnamed cluster object, use it
	if config.Clusters[""] != nil {
		return config.Clusters[""]
	}
	if config.Contexts[config.CurrentContext] != nil {
		return config.Clusters[config.Contexts[config.CurrentContext].Cluster]
	}
	return nil
}

// HasAuthenticationCredentials returns true if the current user has valid authentication credentials for
// token authentication, basic authentication or X509 authentication
func HasAuthenticationCredentials(config *clientcmdapi.Config) bool {
	authInfo := getCurrentAuthInfo(config)
	if authInfo == nil {
		return false
	}

	// token authentication
	if len(authInfo.Token) != 0 {
		return true
	}

	// basic authentication
	if len(authInfo.Username) != 0 && len(authInfo.Password) != 0 {
		return true
	}

	// X509 authentication
	if (len(authInfo.ClientCertificate) != 0 || len(authInfo.ClientCertificateData) != 0) &&
		(len(authInfo.ClientKey) != 0 || len(authInfo.ClientKeyData) != 0) {
		return true
	}

	return false
}

// EnsureAuthenticationInfoAreEmbedded check if some authentication info are provided as external key/certificate
// files, and eventually embeds such files into the kubeconfig file
func EnsureAuthenticationInfoAreEmbedded(config *clientcmdapi.Config) error {
	authInfo := getCurrentAuthInfo(config)
	if authInfo == nil {
		return errors.New("invalid kubeconfig file. AuthInfo is not defined for the current user")
	}

	if len(authInfo.ClientCertificateData) == 0 && len(authInfo.ClientCertificate) != 0 {
		clientCert, err := ioutil.ReadFile(authInfo.ClientCertificate)
		if err != nil {
			return errors.Wrap(err, "error while reading client cert file defined in kubeconfig")
		}
		authInfo.ClientCertificateData = clientCert
		authInfo.ClientCertificate = ""
	}
	if len(authInfo.ClientKeyData) == 0 && len(authInfo.ClientKey) != 0 {
		clientKey, err := ioutil.ReadFile(authInfo.ClientKey)
		if err != nil {
			return errors.Wrap(err, "error while reading client key file defined in kubeconfig")
		}
		authInfo.ClientKeyData = clientKey
		authInfo.ClientKey = ""
	}

	return nil
}

// EnsureCertificateAuthorityIsEmbedded check if the certificate authority is provided as an external
// file and eventually embeds it into the kubeconfig
func EnsureCertificateAuthorityIsEmbedded(cluster *clientcmdapi.Cluster) error {
	if cluster == nil {
		return errors.New("received nil value for Cluster")
	}

	if len(cluster.CertificateAuthorityData) == 0 && len(cluster.CertificateAuthority) != 0 {
		ca, err := ioutil.ReadFile(cluster.CertificateAuthority)
		if err != nil {
			return errors.Wrap(err, "error while reading certificate authority file defined in kubeconfig")
		}
		cluster.CertificateAuthorityData = ca
		cluster.CertificateAuthority = ""
	}

	return nil
}

// getCurrentAuthInfo returns current authInfo, if defined
func getCurrentAuthInfo(config *clientcmdapi.Config) *clientcmdapi.AuthInfo {
	if config == nil || config.CurrentContext == "" ||
		len(config.Contexts) == 0 || config.Contexts[config.CurrentContext] == nil {
		return nil
	}
	user := config.Contexts[config.CurrentContext].AuthInfo

	if user == "" || len(config.AuthInfos) == 0 || config.AuthInfos[user] == nil {
		return nil
	}

	return config.AuthInfos[user]
}
