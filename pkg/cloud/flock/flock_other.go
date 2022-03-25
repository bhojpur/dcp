//go:build !linux && !darwin && !freebsd && !openbsd && !netbsd && !dragonfly
// +build !linux,!darwin,!freebsd,!openbsd,!netbsd,!dragonfly

package flock

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

// Acquire is not implemented on non-unix systems.
func Acquire(path string) (int, error) {
	return -1, nil
}

// AcquireShared creates a shared lock on a file for the duration of the process, or until Release(d).
// This method is reentrant.
func AcquireShared(path string) (int, error) {
	return 0, nil
}

// Release is not implemented on non-unix systems.
func Release(lock int) error {
	return nil
}

// CheckLock checks whether any process is using the lock
func CheckLock(path string) bool {
	return false
}
