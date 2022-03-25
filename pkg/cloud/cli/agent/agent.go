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
	"fmt"
	"os"
	"runtime"

	"github.com/bhojpur/dcp/pkg/cloud/agent"
	"github.com/bhojpur/dcp/pkg/cloud/cli/cmds"
	"github.com/bhojpur/dcp/pkg/cloud/datadir"
	"github.com/bhojpur/dcp/pkg/cloud/netutil"
	"github.com/bhojpur/dcp/pkg/cloud/token"
	"github.com/bhojpur/dcp/pkg/cloud/version"
	"github.com/erikdubbelboer/gspt"
	"github.com/rancher/wrangler/pkg/signals"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func Run(ctx *cli.Context) error {
	// hide process arguments from ps output, since they may contain
	// database credentials or other secrets.
	gspt.SetProcTitle(os.Args[0] + " agent")

	// Evacuate cgroup v2 before doing anything else that may fork.
	if err := cmds.EvacuateCgroup2(); err != nil {
		return err
	}

	// Initialize logging, and subprocess reaping if necessary.
	// Log output redirection and subprocess reaping both require forking.
	if err := cmds.InitLogging(); err != nil {
		return err
	}

	if os.Getuid() != 0 && runtime.GOOS != "windows" {
		return fmt.Errorf("agent must be ran as root")
	}

	if cmds.AgentConfig.TokenFile != "" {
		token, err := token.ReadFile(cmds.AgentConfig.TokenFile)
		if err != nil {
			return err
		}
		cmds.AgentConfig.Token = token
	}

	if cmds.AgentConfig.Token == "" && cmds.AgentConfig.ClusterSecret != "" {
		cmds.AgentConfig.Token = cmds.AgentConfig.ClusterSecret
	}

	if cmds.AgentConfig.Token == "" {
		return fmt.Errorf("--token is required")
	}

	if cmds.AgentConfig.ServerURL == "" {
		return fmt.Errorf("--server is required")
	}

	if cmds.AgentConfig.FlannelIface != "" && len(cmds.AgentConfig.NodeIP) == 0 {
		cmds.AgentConfig.NodeIP.Set(netutil.GetIPFromInterface(cmds.AgentConfig.FlannelIface))
	}

	logrus.Info("Starting " + version.Program + " agent " + ctx.App.Version)

	dataDir, err := datadir.LocalHome(cmds.AgentConfig.DataDir, cmds.AgentConfig.Rootless)
	if err != nil {
		return err
	}

	cfg := cmds.AgentConfig
	cfg.Debug = ctx.GlobalBool("debug")
	cfg.DataDir = dataDir

	contextCtx := signals.SetupSignalContext()

	return agent.Run(contextCtx, cfg)
}
