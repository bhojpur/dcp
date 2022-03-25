package writer

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
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"time"

	"k8s.io/klog"

	"github.com/bhojpur/dcp/pkg/appmanager/webhook/util/generator"
)

const (
	// CAKeyName is the name of the CA private key
	CAKeyName = "ca-key.pem"
	// CACertName is the name of the CA certificate
	CACertName = "ca-cert.pem"
	// ServerKeyName is the name of the server private key
	ServerKeyName  = "key.pem"
	ServerKeyName2 = "tls.key"
	// ServerCertName is the name of the serving certificate
	ServerCertName  = "cert.pem"
	ServerCertName2 = "tls.crt"
)

// CertWriter provides method to handle webhooks.
type CertWriter interface {
	// EnsureCert provisions the cert for the webhookClientConfig.
	EnsureCert(dnsName string) (*generator.Artifacts, bool, error)
}

// handleCommon ensures the given webhook has a proper certificate.
// It uses the given certReadWriter to read and (or) write the certificate.
func handleCommon(dnsName string, ch certReadWriter) (*generator.Artifacts, bool, error) {
	if len(dnsName) == 0 {
		return nil, false, errors.New("dnsName should not be empty")
	}
	if ch == nil {
		return nil, false, errors.New("certReaderWriter should not be nil")
	}

	certs, changed, err := createIfNotExists(ch)
	if err != nil {
		return nil, changed, err
	}
	// Recreate the cert if it's invalid.
	valid := validCert(certs, dnsName)
	if !valid {
		klog.Info("cert is invalid or expiring, regenerating a new one")
		certs, err = ch.overwrite(certs.ResourceVersion)
		if err != nil {
			return nil, false, err
		}
		changed = true
	}
	return certs, changed, nil
}

func createIfNotExists(ch certReadWriter) (*generator.Artifacts, bool, error) {
	// Try to read first
	certs, err := ch.read()
	if isNotFound(err) {
		// Create if not exists
		certs, err = ch.write()
		switch {
		// This may happen if there is another racer.
		case isAlreadyExists(err):
			certs, err = ch.read()
			return certs, true, err
		default:
			return certs, true, err
		}
	}
	return certs, false, err
}

// certReadWriter provides methods for reading and writing certificates.
type certReadWriter interface {
	// read reads a webhook name and returns the certs for it.
	read() (*generator.Artifacts, error)
	// write writes the certs and return the certs it wrote.
	write() (*generator.Artifacts, error)
	// overwrite overwrites the existing certs and return the certs it wrote.
	overwrite(resourceVersion string) (*generator.Artifacts, error)
}

func validCert(certs *generator.Artifacts, dnsName string) bool {
	if certs == nil || certs.Cert == nil || certs.Key == nil || certs.CACert == nil {
		klog.Errorf("valid cert error is null")
		return false
	}

	// Verify key and cert are valid pair
	_, err := tls.X509KeyPair(certs.Cert, certs.Key)
	if err != nil {
		klog.Errorf("valid cert key pair error %v", err)
		return false
	}

	// Verify cert is good for desired DNS name and signed by CA and will be valid for desired period of time.
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(certs.CACert) {
		klog.Errorf("valid cert append cert error %v", err)
		return false
	}
	block, _ := pem.Decode(certs.Cert)
	if block == nil {
		klog.Errorf("valid cert decode error %v", err)
		return false
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		klog.Errorf("valid cert parse cert error %v", err)
		return false
	}
	ops := x509.VerifyOptions{
		DNSName:     dnsName,
		Roots:       pool,
		CurrentTime: time.Now().AddDate(0, 6, 0),
	}
	_, err = cert.Verify(ops)
	if err != nil {
		klog.Errorf("valid cert verify error %v", err)
	}
	return err == nil
}
