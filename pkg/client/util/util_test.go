package util

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
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DownloadFile(t *testing.T) {
	testCase := []struct {
		name     string
		url      string
		retry    int
		expected bool
	}{
		{"invalid download url", "http://1.2.3.4", 2, false},
	}

	for _, tc := range testCase {
		err := DownloadFile(tc.url, "", tc.retry)
		assert.Equal(t, tc.expected, err == nil)
	}
}

func Test_Untar(t *testing.T) {
	testCase := []struct {
		name     string
		filePath string
		expected bool
	}{
		{"invalid tar file", "/tmp", false},
	}

	for _, tc := range testCase {
		err := Untar(tc.name, "")
		assert.Equal(t, tc.expected, err == nil)
	}
}
