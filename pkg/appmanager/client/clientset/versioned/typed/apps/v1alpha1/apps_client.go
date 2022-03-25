// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

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
	v1alpha1 "github.com/bhojpur/dcp/pkg/appmanager/apis/apps/v1alpha1"
	"github.com/bhojpur/dcp/pkg/appmanager/client/clientset/versioned/scheme"
	rest "k8s.io/client-go/rest"
)

type AppsV1alpha1Interface interface {
	RESTClient() rest.Interface
	NodePoolsGetter
	UnitedDeploymentsGetter
	DcpAppDaemonsGetter
	DcpIngressesGetter
}

// AppsV1alpha1Client is used to interact with features provided by the apps.bhojpur.net group.
type AppsV1alpha1Client struct {
	restClient rest.Interface
}

func (c *AppsV1alpha1Client) NodePools() NodePoolInterface {
	return newNodePools(c)
}

func (c *AppsV1alpha1Client) UnitedDeployments(namespace string) UnitedDeploymentInterface {
	return newUnitedDeployments(c, namespace)
}

func (c *AppsV1alpha1Client) DcpAppDaemons(namespace string) DcpAppDaemonInterface {
	return newDcpAppDaemons(c, namespace)
}

func (c *AppsV1alpha1Client) DcpIngresses() DcpIngressInterface {
	return newDcpIngresses(c)
}

// NewForConfig creates a new AppsV1alpha1Client for the given config.
func NewForConfig(c *rest.Config) (*AppsV1alpha1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &AppsV1alpha1Client{client}, nil
}

// NewForConfigOrDie creates a new AppsV1alpha1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *AppsV1alpha1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new AppsV1alpha1Client for the given RESTClient.
func New(c rest.Interface) *AppsV1alpha1Client {
	return &AppsV1alpha1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1alpha1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *AppsV1alpha1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}