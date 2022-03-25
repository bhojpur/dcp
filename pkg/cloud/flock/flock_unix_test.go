//go:build linux || darwin || freebsd || openbsd || netbsd || dragonfly
// +build linux darwin freebsd openbsd netbsd dragonfly

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

import (
	"testing"
)

func Test_UnitFlock(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		wantCheck bool
		wantErr   bool
	}{
		{
			name: "Basic Flock Test",
			path: "/tmp/testlock.test",

			wantCheck: true,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lock, err := Acquire(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Acquire() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got := CheckLock(tt.path); got != tt.wantCheck {
				t.Errorf("CheckLock() = %+v\nWant = %+v", got, tt.wantCheck)
			}

			if err := Release(lock); (err != nil) != tt.wantErr {
				t.Errorf("Release() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got := CheckLock(tt.path); got == tt.wantCheck {
				t.Errorf("CheckLock() = %+v\nWant = %+v", got, !tt.wantCheck)
			}
		})
	}
}
