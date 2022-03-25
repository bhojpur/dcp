package store

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

	"k8s.io/client-go/util/certificate"
	"k8s.io/klog/v2"
)

// fileStoreWrapper is a wrapper for "k8s.io/client-go/util/certificate#FileStore"
// This wrapper increases tolerance for unexpected situations and is more robust.
type fileStoreWrapper struct {
	certificate.FileStore
}

// NewFileStoreWrapper returns a wrapper for "k8s.io/client-go/util/certificate#FileStore"
// This wrapper increases tolerance for unexpected situations and is more robust.
func NewFileStoreWrapper(pairNamePrefix, certDirectory, keyDirectory, certFile, keyFile string) (certificate.FileStore, error) {
	fileStore, err := certificate.NewFileStore(pairNamePrefix, certDirectory, keyDirectory, certFile, keyFile)
	if err != nil {
		return nil, err
	}
	return &fileStoreWrapper{
		FileStore: fileStore,
	}, nil
}

func (s *fileStoreWrapper) Current() (*tls.Certificate, error) {
	cert, err := s.FileStore.Current()
	// If an error occurs, just return the NoCertKeyError.
	// The cert-manager will regenerate the related certificates when it receives the NoCertKeyError.
	if err != nil {
		klog.Warningf("unexpected error occurred when loading the certificate: %v, will regenerate it", err)
		noCertKeyErr := certificate.NoCertKeyError("NO_VALID_CERT")
		return nil, &noCertKeyErr
	}
	return cert, nil
}
