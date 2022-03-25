package kubernetes

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
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/pkg/tunnel/constants"
)

// CreateClientSet creates a clientset based on the given kubeConfig. If the
// kubeConfig is empty, it will creates the clientset based on the in-cluster
// config
func CreateClientSet(kubeConfig string) (*kubernetes.Clientset, error) {

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

// CreateClientSet creates a clientset based on the given kubeconfig
func CreateClientSetKubeConfig(kubeConfig string) (*kubernetes.Clientset, error) {
	var (
		cfg *rest.Config
		err error
	)
	if kubeConfig == "" {
		return nil, errors.New("kubeconfig is not set")
	}
	if _, err := os.Stat(kubeConfig); err != nil && os.IsNotExist(err) {
		return nil, err
	}
	cfg, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("fail to create the clientset based on %s: %v",
			kubeConfig, err)
	}
	cliSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return cliSet, nil
}

// CreateClientSetApiserverAddr creates a clientset based on the given apiserverAddr.
// The clientset uses the serviceaccount's CA and Token for authentication and
// authorization.
func CreateClientSetApiserverAddr(apiserverAddr string) (*kubernetes.Clientset, error) {
	if apiserverAddr == "" {
		return nil, errors.New("apiserver addr can't be empty")
	}

	token, err := ioutil.ReadFile(constants.TunnelTokenFile)
	if err != nil {
		return nil, err
	}

	tlsClientConfig := rest.TLSClientConfig{}

	if _, err := certutil.NewPool(constants.TunnelCAFile); err != nil {
		klog.Errorf("Expected to load root CA config from %s, but got err: %v",
			constants.TunnelCAFile, err)
	} else {
		tlsClientConfig.CAFile = constants.TunnelCAFile
	}

	restConfig := rest.Config{
		Host:            "https://" + apiserverAddr,
		TLSClientConfig: tlsClientConfig,
		BearerToken:     string(token),
		BearerTokenFile: constants.TunnelTokenFile,
	}

	return kubernetes.NewForConfig(&restConfig)
}
