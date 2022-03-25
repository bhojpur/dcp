//go:build linux
// +build linux

package containerd

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
	"io/fs"
	"path/filepath"

	"github.com/bhojpur/dcp/pkg/cloud/agent/templates"
	"github.com/sirupsen/logrus"
)

// findNvidiaContainerRuntimes returns a list of nvidia container runtimes that
// are available on the system. It checks install locations used by the nvidia
// gpu operator and by system package managers. The gpu operator installation
// takes precedence over the system package manager installation.
// The given fs.FS should represent the filesystem root directory to search in.
func findNvidiaContainerRuntimes(root fs.FS) map[string]templates.ContainerdRuntimeConfig {
	// Check these locations in order. The GPU operator's installation should
	// take precedence over the package manager's installation.
	locationsToCheck := []string{
		"usr/local/nvidia/toolkit", // Path when installing via GPU Operator
		"usr/bin",                  // Path when installing via package manager
	}

	// Fill in the binary location with just the name of the binary,
	// and check against each of the possible locations. If a match is found,
	// set the location to the full path.
	potentialRuntimes := map[string]templates.ContainerdRuntimeConfig{
		"nvidia": {
			RuntimeType: "io.containerd.runc.v2",
			BinaryName:  "nvidia-container-runtime",
		},
		"nvidia-experimental": {
			RuntimeType: "io.containerd.runc.v2",
			BinaryName:  "nvidia-container-runtime-experimental",
		},
	}
	foundRuntimes := map[string]templates.ContainerdRuntimeConfig{}
RUNTIME:
	for runtimeName, runtimeConfig := range potentialRuntimes {
		for _, location := range locationsToCheck {
			binaryPath := filepath.Join(location, runtimeConfig.BinaryName)
			logrus.Debugf("Searching for %s container runtime at /%s", runtimeName, binaryPath)
			if info, err := fs.Stat(root, binaryPath); err == nil {
				if info.IsDir() {
					logrus.Debugf("Found %s container runtime at /%s, but it is a directory. Skipping.", runtimeName, binaryPath)
					continue
				}
				runtimeConfig.BinaryName = filepath.Join("/", binaryPath)
				logrus.Infof("Found %s container runtime at %s", runtimeName, runtimeConfig.BinaryName)
				foundRuntimes[runtimeName] = runtimeConfig
				// Skip to the next runtime to enforce precedence.
				continue RUNTIME
			} else {
				if errors.Is(err, fs.ErrNotExist) {
					logrus.Debugf("%s container runtime not found at /%s", runtimeName, binaryPath)
				} else {
					logrus.Errorf("Error searching for %s container runtime at /%s: %v", runtimeName, binaryPath, err)
				}
			}
		}
	}
	return foundRuntimes
}
