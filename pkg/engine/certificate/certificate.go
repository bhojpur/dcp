package certificate

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
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/cmd/server/config"
	"github.com/bhojpur/dcp/pkg/engine/certificate/interfaces"
)

// Factory is a function that returns an EngineCertificateManager.
// The cfg parameter provides the common info for certificate manager
type Factory func(cfg *config.EngineConfiguration) (interfaces.EngineCertificateManager, error)

// CertificateManagerRegistry is a object for holding all certificate managers
type CertificateManagerRegistry struct {
	sync.Mutex
	registry map[string]Factory
}

// NewCertificateManagerRegistry creates an *CertificateManagerRegistry object
func NewCertificateManagerRegistry() *CertificateManagerRegistry {
	return &CertificateManagerRegistry{}
}

// Register register a Factory func for creating certificate manager
func (cmr *CertificateManagerRegistry) Register(name string, cm Factory) {
	cmr.Lock()
	defer cmr.Unlock()

	if cmr.registry == nil {
		cmr.registry = map[string]Factory{}
	}

	_, found := cmr.registry[name]
	if found {
		klog.Fatalf("certificate manager %s was registered twice", name)
	}

	klog.Infof("Registered certificate manager %s", name)
	cmr.registry[name] = cm
}

// New creates a EngineCertificateManager with specified name of registered certificate manager
func (cmr *CertificateManagerRegistry) New(name string, cfg *config.EngineConfiguration) (interfaces.EngineCertificateManager, error) {
	f, found := cmr.registry[name]
	if !found {
		return nil, fmt.Errorf("certificate manager %s is not registered", name)
	}

	cm, err := f(cfg)
	if err != nil {
		return nil, err
	}

	cm.Start()
	err = wait.PollImmediate(5*time.Second, 4*time.Minute, func() (bool, error) {
		curr := cm.Current()
		if curr != nil {
			return true, nil
		}

		klog.Infof("waiting for preparing client certificate")
		return false, nil
	})
	if err != nil {
		klog.Errorf("client certificate preparation failed, %v", err)
		return nil, err
	}

	return cm, nil
}
