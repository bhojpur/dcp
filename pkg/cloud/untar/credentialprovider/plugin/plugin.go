package plugin

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
	"flag"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/google/go-containerregistry/pkg/authn"
	"k8s.io/klog/v2"
	kubecredentialprovider "k8s.io/kubernetes/pkg/credentialprovider"
	kubeplugin "k8s.io/kubernetes/pkg/credentialprovider/plugin"
)

type pluginWrapper struct {
	k kubecredentialprovider.DockerKeyring
}

// Explicit interface checks
var _ authn.Keychain = &pluginWrapper{}

// RegisterCredentialProviderPlugins loads the provided configuration into the credentialprovider plugin registry
// If the configuration is not valid or any configured plugins are missing, an error will be raised.
func RegisterCredentialProviderPlugins(imageCredentialProviderConfigFile, imageCredentialProviderBinDir string) (*pluginWrapper, error) {
	klogSetup()
	if err := kubeplugin.RegisterCredentialProviderPlugins(imageCredentialProviderConfigFile, imageCredentialProviderBinDir); err != nil {
		return nil, errors.Wrap(err, "failed to register CRI auth plugins")
	}
	return &pluginWrapper{k: kubecredentialprovider.NewDockerKeyring()}, nil
}

// Resolve returns an authenticator for the authn.Keychain interface. The authenticator provides
// credentials to a registry by calling the credentialprovider plugin registry's Lookup method,
// which in turn consults the configuration and executes plugins to obtain credentials.
func (p *pluginWrapper) Resolve(target authn.Resource) (authn.Authenticator, error) {
	// Lookup may provide multiple AuthConfigs (for credential rotation support) but the Keychain interface only allows us to return one.
	if configs, ok := p.k.Lookup(target.String()); ok {
		return authn.FromConfig(authn.AuthConfig{
			Username:      configs[0].Username,
			Password:      configs[0].Password,
			Auth:          configs[0].Auth,
			IdentityToken: configs[0].IdentityToken,
			RegistryToken: configs[0].RegistryToken,
		}), nil
	}

	return authn.Anonymous, nil
}

// klogSetup syncs the klog verbosity to the current Logrus log level. This is necessary because the
// auth plugin stuff all uses klog/v2 and there's no good translation layer between logrus and klog.
func klogSetup() {
	klogFlags := flag.NewFlagSet("klog", flag.ContinueOnError)
	klog.InitFlags(klogFlags)
	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		_ = klogFlags.Set("v", "9")
	}
	_ = klogFlags.Parse(nil)
}
