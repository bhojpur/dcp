package file

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
	"encoding/json"
	"os"

	"github.com/bhojpur/dcp/pkg/cloud/dynamiclistener"
	v1 "k8s.io/api/core/v1"
)

func New(file string) dynamiclistener.TLSStorage {
	return &storage{
		file: file,
	}
}

type storage struct {
	file string
}

func (s *storage) Get() (*v1.Secret, error) {
	f, err := os.Open(s.file)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	defer f.Close()

	secret := v1.Secret{}
	return &secret, json.NewDecoder(f).Decode(&secret)
}

func (s *storage) Update(secret *v1.Secret) error {
	f, err := os.Create(s.file)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(secret)
}
