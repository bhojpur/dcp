package generator

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

// Artifacts hosts a private key, its corresponding serving certificate and
// the CA certificate that signs the serving certificate.
type Artifacts struct {
	// PEM encoded private key
	Key []byte
	// PEM encoded serving certificate
	Cert []byte
	// PEM encoded CA private key
	CAKey []byte
	// PEM encoded CA certificate
	CACert []byte
	// Resource version of the certs
	ResourceVersion string
}

// CertGenerator is an interface to provision the serving certificate.
type CertGenerator interface {
	// Generate returns a Artifacts struct.
	Generate(CommonName string) (*Artifacts, error)
	// SetCA sets the PEM-encoded CA private key and CA cert for signing the generated serving cert.
	SetCA(caKey, caCert []byte)
}
