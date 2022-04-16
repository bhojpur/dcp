package static

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
	"github.com/bhojpur/dcp/pkg/cloud/dynamiclistener/factory"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Storage struct {
	Secret *v1.Secret
}

func New(certPem, keyPem []byte) *Storage {
	return &Storage{
		Secret: &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					factory.Static: "true",
				},
			},
			Data: map[string][]byte{
				v1.TLSCertKey:       certPem,
				v1.TLSPrivateKeyKey: keyPem,
			},
			Type: v1.SecretTypeTLS,
		},
	}
}

func (s *Storage) Get() (*v1.Secret, error) {
	return s.Secret, nil
}

func (s *Storage) Update(_ *v1.Secret) error {
	return nil
}
