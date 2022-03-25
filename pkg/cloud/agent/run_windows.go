//go:build windows
// +build windows

package agent

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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/bhojpur/dcp/pkg/cloud/cli/cmds"
	"github.com/bhojpur/dcp/pkg/cloud/daemons/config"
)

const (
	dockershimSock = "npipe:////./pipe/docker_engine"
	containerdSock = "npipe:////./pipe/containerd-containerd"
)

// setupCriCtlConfig creates the crictl config file and populates it
// with the given data from config.
func setupCriCtlConfig(cfg cmds.Agent, nodeConfig *config.Node) error {
	cre := nodeConfig.ContainerRuntimeEndpoint
	if cre == "" || strings.HasPrefix(cre, "npipe:") {
		switch {
		case cfg.Docker:
			cre = dockershimSock
		default:
			cre = containerdSock
		}
	} else {
		cre = containerdSock
	}
	agentConfDir := filepath.Join(cfg.DataDir, "agent", "etc")
	if _, err := os.Stat(agentConfDir); os.IsNotExist(err) {
		if err := os.MkdirAll(agentConfDir, 0700); err != nil {
			return err
		}
	}

	crp := "runtime-endpoint: " + cre + "\n"
	return ioutil.WriteFile(filepath.Join(agentConfDir, "crictl.yaml"), []byte(crp), 0600)
}
