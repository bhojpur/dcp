package pubkeypin

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

// It provides primitives for x509 public key pinning in the style of RFC7469.

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"strings"

	"github.com/pkg/errors"
)

const (
	// formatSHA256 is the prefix for pins that are full-length SHA-256 hashes encoded in base 16 (hex)
	formatSHA256 = "sha256"
)

// Set is a set of pinned x509 public keys.
type Set struct {
	sha256Hashes map[string]bool
}

// NewSet returns a new, empty PubKeyPinSet
func NewSet() *Set {
	return &Set{make(map[string]bool)}
}

// Allow adds an allowed public key hash to the Set
func (s *Set) Allow(pubKeyHashes ...string) error {
	for _, pubKeyHash := range pubKeyHashes {
		parts := strings.Split(pubKeyHash, ":")
		if len(parts) != 2 {
			return errors.New("invalid public key hash, expected \"format:value\"")
		}
		format, value := parts[0], parts[1]

		switch strings.ToLower(format) {
		case "sha256":
			return s.allowSHA256(value)
		default:
			return errors.Errorf("unknown hash format %q", format)
		}
	}
	return nil
}

// CheckAny checks if at least one certificate matches one of the public keys in the set
func (s *Set) CheckAny(certificates []*x509.Certificate) error {
	var hashes []string

	for _, certificate := range certificates {
		if s.checkSHA256(certificate) {
			return nil
		}

		hashes = append(hashes, Hash(certificate))
	}
	return errors.Errorf("none of the public keys %q are pinned", strings.Join(hashes, ":"))
}

// Empty returns true if the Set contains no pinned public keys.
func (s *Set) Empty() bool {
	return len(s.sha256Hashes) == 0
}

// Hash calculates the SHA-256 hash of the Subject Public Key Information (SPKI)
// object in an x509 certificate (in DER encoding). It returns the full hash as a
// hex encoded string (suitable for passing to Set.Allow).
func Hash(certificate *x509.Certificate) string {
	spkiHash := sha256.Sum256(certificate.RawSubjectPublicKeyInfo)
	return formatSHA256 + ":" + strings.ToLower(hex.EncodeToString(spkiHash[:]))
}

// allowSHA256 validates a "sha256" format hash and adds a canonical version of it into the Set
func (s *Set) allowSHA256(hash string) error {
	// validate that the hash is the right length to be a full SHA-256 hash
	hashLength := hex.DecodedLen(len(hash))
	if hashLength != sha256.Size {
		return errors.Errorf("expected a %d byte SHA-256 hash, found %d bytes", sha256.Size, hashLength)
	}

	// validate that the hash is valid hex
	_, err := hex.DecodeString(hash)
	if err != nil {
		return err
	}

	// in the end, just store the original hex string in memory (in lowercase)
	s.sha256Hashes[strings.ToLower(hash)] = true
	return nil
}

// checkSHA256 returns true if the certificate's "sha256" hash is pinned in the Set
func (s *Set) checkSHA256(certificate *x509.Certificate) bool {
	actualHash := sha256.Sum256(certificate.RawSubjectPublicKeyInfo)
	actualHashHex := strings.ToLower(hex.EncodeToString(actualHash[:]))
	return s.sha256Hashes[actualHashHex]
}
