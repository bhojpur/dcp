package storage

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
	"errors"
)

// ErrStorageAccessConflict is an error for accessing key conflict
var ErrStorageAccessConflict = errors.New("specified key is under accessing")

// ErrStorageNotFound is an error for not found accessing key
var ErrStorageNotFound = errors.New("specified key is not found")

// ErrKeyHasNoContent is an error for file key that has no contents
var ErrKeyHasNoContent = errors.New("specified key has no contents")

// ErrKeyIsEmpty is an error for key is empty
var ErrKeyIsEmpty = errors.New("specified key is empty")

// ErrRootKeyInvalid is an error for root key is invalid
var ErrRootKeyInvalid = errors.New("root key is invalid")

// Store is an interface for caching data into backend storage
type Store interface {
	Create(key string, contents []byte) error
	Delete(key string) error
	Get(key string) ([]byte, error)
	ListKeys(key string) ([]string, error)
	List(key string) ([][]byte, error)
	Update(key string, contents []byte) error
	Replace(rootKey string, contents map[string][]byte) error
	DeleteCollection(rootKey string) error
}
