package factory

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
	"crypto"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/bhojpur/dcp/pkg/cloud/dynamiclistener/cert"
)

func GenCA() (*x509.Certificate, crypto.Signer, error) {
	caKey, err := NewPrivateKey()
	if err != nil {
		return nil, nil, err
	}

	caCert, err := NewSelfSignedCACert(caKey, "dynamiclistener-ca", "dynamiclistener-org")
	if err != nil {
		return nil, nil, err
	}

	return caCert, caKey, nil
}

func LoadOrGenCA() (*x509.Certificate, crypto.Signer, error) {
	cert, key, err := loadCA()
	if err == nil {
		return cert, key, nil
	}

	cert, key, err = GenCA()
	if err != nil {
		return nil, nil, err
	}

	certBytes, keyBytes, err := Marshal(cert, key)
	if err != nil {
		return nil, nil, err
	}

	if err := os.MkdirAll("./certs", 0700); err != nil {
		return nil, nil, err
	}

	if err := ioutil.WriteFile("./certs/ca.pem", certBytes, 0600); err != nil {
		return nil, nil, err
	}

	if err := ioutil.WriteFile("./certs/ca.key", keyBytes, 0600); err != nil {
		return nil, nil, err
	}

	return cert, key, nil
}

func loadCA() (*x509.Certificate, crypto.Signer, error) {
	return LoadCerts("./certs/ca.pem", "./certs/ca.key")
}

func LoadCA(caPem, caKey []byte) (*x509.Certificate, crypto.Signer, error) {
	key, err := cert.ParsePrivateKeyPEM(caKey)
	if err != nil {
		return nil, nil, err
	}
	signer, ok := key.(crypto.Signer)
	if !ok {
		return nil, nil, fmt.Errorf("key is not a crypto.Signer")
	}

	cert, err := ParseCertPEM(caPem)
	if err != nil {
		return nil, nil, err
	}

	return cert, signer, nil
}

func LoadCerts(certFile, keyFile string) (*x509.Certificate, crypto.Signer, error) {
	caPem, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, nil, err
	}
	caKey, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, nil, err
	}

	return LoadCA(caPem, caKey)
}
