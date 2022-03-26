package tests

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
	"net"
	"os"
	"path/filepath"

	"github.com/bhojpur/dcp/pkg/cloud/daemons/config"
	"github.com/bhojpur/dcp/pkg/cloud/daemons/control/deps"
)

// GenerateDataDir creates a temporary directory at "/tmp/bhojpur/<RANDOM_STRING>/".
// The latest directory created with this function is soft linked to "/tmp/bhojpur/latest/".
// This allows tests to replicate the "/var/lib/bhojpur/dcp" directory structure.
func GenerateDataDir(cnf *config.Control) error {
	if err := os.MkdirAll(cnf.DataDir, 0700); err != nil {
		return err
	}
	testDir, err := os.MkdirTemp(cnf.DataDir, "*")
	if err != nil {
		return err
	}
	// Remove old symlink and add new one
	os.Remove(filepath.Join(cnf.DataDir, "latest"))
	if err = os.Symlink(testDir, filepath.Join(cnf.DataDir, "latest")); err != nil {
		return err
	}
	cnf.DataDir = testDir
	cnf.DataDir, err = filepath.Abs(cnf.DataDir)
	if err != nil {
		return err
	}
	return nil
}

// CleanupDataDir removes the associated "/tmp/bhojpur/<RANDOM_STRING>"
// directory along with the 'latest' symlink that points at it.
func CleanupDataDir(cnf *config.Control) {
	os.Remove(filepath.Join(cnf.DataDir, "..", "latest"))
	os.RemoveAll(cnf.DataDir)
}

// GenerateRuntime creates a temporary data dir and configures
// config.ControlRuntime with all the appropriate certificate keys.
func GenerateRuntime(cnf *config.Control) error {
	cnf.Runtime = &config.ControlRuntime{}
	if err := GenerateDataDir(cnf); err != nil {
		return err
	}

	os.MkdirAll(filepath.Join(cnf.DataDir, "tls"), 0700)
	os.MkdirAll(filepath.Join(cnf.DataDir, "cred"), 0700)

	deps.CreateRuntimeCertFiles(cnf)

	return deps.GenServerDeps(cnf)
}

func ClusterIPNet() *net.IPNet {
	_, clusterIPNet, _ := net.ParseCIDR("10.42.0.0/16")
	return clusterIPNet
}

func ServiceIPNet() *net.IPNet {
	_, serviceIPNet, _ := net.ParseCIDR("10.43.0.0/16")
	return serviceIPNet
}
