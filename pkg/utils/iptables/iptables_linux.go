package iptables

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
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/sys/unix"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
)

type locker struct {
	lock16 *os.File
	lock14 *net.UnixListener
}

func (l *locker) Close() error {
	errList := []error{}
	if l.lock16 != nil {
		if err := l.lock16.Close(); err != nil {
			errList = append(errList, err)
		}
	}
	if l.lock14 != nil {
		if err := l.lock14.Close(); err != nil {
			errList = append(errList, err)
		}
	}
	return utilerrors.NewAggregate(errList)
}

func grabIptablesLocks(lockfilePath string) (iptablesLocker, error) {
	var err error
	var success bool

	l := &locker{}
	defer func(l *locker) {
		// Clean up immediately on failure
		if !success {
			l.Close()
		}
	}(l)

	// Grab both 1.6.x and 1.4.x-style locks; we don't know what the
	// iptables-restore version is if it doesn't support --wait, so we
	// can't assume which lock method it'll use.

	// Roughly duplicate iptables 1.6.x xtables_lock() function.
	l.lock16, err = os.OpenFile(lockfilePath, os.O_CREATE, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open iptables lock %s: %v", lockfilePath, err)
	}

	if err := wait.PollImmediate(200*time.Millisecond, 2*time.Second, func() (bool, error) {
		if err := grabIptablesFileLock(l.lock16); err != nil {
			return false, nil
		}
		return true, nil
	}); err != nil {
		return nil, fmt.Errorf("failed to acquire new iptables lock: %v", err)
	}

	// Roughly duplicate iptables 1.4.x xtables_lock() function.
	if err := wait.PollImmediate(200*time.Millisecond, 2*time.Second, func() (bool, error) {
		l.lock14, err = net.ListenUnix("unix", &net.UnixAddr{Name: "@xtables", Net: "unix"})
		if err != nil {
			return false, nil
		}
		return true, nil
	}); err != nil {
		return nil, fmt.Errorf("failed to acquire old iptables lock: %v", err)
	}

	success = true
	return l, nil
}

func grabIptablesFileLock(f *os.File) error {
	return unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB)
}
