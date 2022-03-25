package hash

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
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"

	"golang.org/x/crypto/scrypt"
)

// Version is the hashing format version
const Version = 1
const hashFormat = "$%d:%x:%d:%d:%d:%s"

// SCrypt contains all of the variables needed for scrypt hashing
type SCrypt struct {
	N       int
	R       int
	P       int
	KeyLen  int
	SaltLen int
}

// NewSCrypt returns a scrypt hasher with recommended default values
func NewSCrypt() Hasher {
	return SCrypt{
		N:       15,
		R:       8,
		P:       1,
		KeyLen:  64,
		SaltLen: 8,
	}
}

// CreateHash will return a hashed version of the secretKey, or an error
func (s SCrypt) CreateHash(secretKey string) (string, error) {
	salt := make([]byte, s.SaltLen)

	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	dk, err := scrypt.Key([]byte(secretKey), salt, 1<<s.N, s.R, s.P, s.KeyLen)
	if err != nil {
		return "", err
	}

	enc := base64.RawStdEncoding.EncodeToString(dk)
	hash := fmt.Sprintf(hashFormat, Version, salt, s.N, s.R, s.P, enc)

	return hash, nil
}

// VerifyHash will compare a secretKey and a hash, and return nil if they match
func (s SCrypt) VerifyHash(hash, secretKey string) error {
	var (
		version, n uint
		r, p       int
		enc        string
		salt       []byte
	)
	_, err := fmt.Sscanf(hash, hashFormat, &version, &salt, &n, &r, &p, &enc)
	if err != nil {
		return err
	}
	if version != Version {
		return fmt.Errorf("hash version %d does not match package version %d", version, Version)
	}

	dk, err := base64.RawStdEncoding.DecodeString(enc)
	if err != nil {
		return err
	}

	verify, err := scrypt.Key([]byte(secretKey), salt, 1<<n, r, p, len(dk))
	if err != nil {
		return err
	}

	if subtle.ConstantTimeCompare(dk, verify) != 1 {
		return errors.New("hash does not match")
	}

	return nil
}
