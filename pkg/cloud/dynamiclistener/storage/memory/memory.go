package memory

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
	"github.com/bhojpur/dcp/pkg/cloud/dynamiclistener"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

func New() dynamiclistener.TLSStorage {
	return &memory{}
}

func NewBacked(storage dynamiclistener.TLSStorage) dynamiclistener.TLSStorage {
	return &memory{storage: storage}
}

type memory struct {
	storage dynamiclistener.TLSStorage
	secret  *v1.Secret
}

func (m *memory) Get() (*v1.Secret, error) {
	if m.secret == nil && m.storage != nil {
		secret, err := m.storage.Get()
		if err != nil {
			return nil, err
		}
		m.secret = secret
	}

	return m.secret, nil
}

func (m *memory) Update(secret *v1.Secret) error {
	if m.secret == nil || m.secret.ResourceVersion == "" || m.secret.ResourceVersion != secret.ResourceVersion {
		if m.storage != nil {
			if err := m.storage.Update(secret); err != nil {
				return err
			}
		}

		logrus.Infof("Active TLS secret %s (ver=%s) (count %d): %v", secret.Name, secret.ResourceVersion, len(secret.Annotations)-1, secret.Annotations)
		m.secret = secret
	}
	return nil
}
