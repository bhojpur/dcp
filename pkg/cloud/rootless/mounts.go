//go:build !windows
// +build !windows

package rootless

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
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

func setupMounts(stateDir string) error {
	// Remove symlinks to the rootful files, so that we can create our own files.
	removeList := []string{
		"/var/run/netns",
		"/run/containerd",
		"/run/xtables.lock",
	}
	for _, f := range removeList {
		_ = os.RemoveAll(f)
	}

	mountMap := [][]string{
		{"/var/log", filepath.Join(stateDir, "logs")},
		{"/var/lib/cni", filepath.Join(stateDir, "cni")},
		{"/var/lib/kubelet", filepath.Join(stateDir, "kubelet")},
		{"/etc/rancher", filepath.Join(stateDir, "etc", "rancher")},
	}

	for _, v := range mountMap {
		if err := setupMount(v[0], v[1]); err != nil {
			return errors.Wrapf(err, "failed to setup mount %s => %s", v[0], v[1])
		}
	}

	if devKmsg, err := os.Open("/dev/kmsg"); err == nil {
		devKmsg.Close()
	} else {
		// kubelet requires /dev/kmsg to be readable
		// https://github.com/rootless-containers/usernetes/issues/204
		// https://github.com/rootless-containers/usernetes/pull/214
		logrus.Debugf("`kernel.dmesg_restrict` seems to be set, bind-mounting /dev/null into /dev/kmsg")
		if err := unix.Mount("/dev/null", "/dev/kmsg", "none", unix.MS_BIND, ""); err != nil {
			return err
		}
	}

	return nil
}

func setupMount(target, dir string) error {
	toCreate := target
	for {
		if toCreate == "/" {
			return fmt.Errorf("missing /%s on the root filesystem", strings.Split(target, "/")[0])
		}

		if err := os.MkdirAll(toCreate, 0700); err == nil {
			break
		}

		toCreate = filepath.Base(toCreate)
	}

	if err := os.MkdirAll(toCreate, 0700); err != nil {
		return errors.Wrapf(err, "failed to create directory %s", toCreate)
	}

	logrus.Debug("Mounting none ", toCreate, " tmpfs")
	if err := unix.Mount("none", toCreate, "tmpfs", 0, ""); err != nil {
		return errors.Wrapf(err, "failed to mount tmpfs to %s", toCreate)
	}

	if err := os.MkdirAll(target, 0700); err != nil {
		return errors.Wrapf(err, "failed to create directory %s", target)
	}

	if dir == "" {
		return nil
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return errors.Wrapf(err, "failed to create directory %s", dir)
	}

	logrus.Debug("Mounting ", dir, target, " none bind")
	return unix.Mount(dir, target, "none", unix.MS_BIND, "")
}
