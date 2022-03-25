package fake

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
	"bytes"
	"fmt"

	"github.com/bhojpur/dcp/pkg/appmanager/webhook/util/generator"
)

// CertGenerator is a certGenerator for testing.
type CertGenerator struct {
	CAKey                  []byte
	CACert                 []byte
	DNSNameToCertArtifacts map[string]*generator.Artifacts
}

var _ generator.CertGenerator = &CertGenerator{}

// SetCA sets the PEM-encoded CA private key and CA cert for signing the generated serving cert.
func (cp *CertGenerator) SetCA(CAKey, CACert []byte) {
	cp.CAKey = CAKey
	cp.CACert = CACert
}

// Generate generates certificates by matching a common name.
func (cp *CertGenerator) Generate(commonName string) (*generator.Artifacts, error) {
	certs, found := cp.DNSNameToCertArtifacts[commonName]
	if !found {
		return nil, fmt.Errorf("failed to find common name %q in the certGenerator", commonName)
	}
	if cp.CAKey != nil && cp.CACert != nil &&
		!bytes.Contains(cp.CAKey, []byte("invalid")) && !bytes.Contains(cp.CACert, []byte("invalid")) {
		certs.CAKey = cp.CAKey
		certs.CACert = cp.CACert
	}
	return certs, nil
}
