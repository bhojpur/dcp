package main

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
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/bhojpur/dcp/pkg/cloud/cli/agent"
	"github.com/bhojpur/dcp/pkg/cloud/cli/cert"
	"github.com/bhojpur/dcp/pkg/cloud/cli/cmds"
	"github.com/bhojpur/dcp/pkg/cloud/cli/crictl"
	"github.com/bhojpur/dcp/pkg/cloud/cli/ctr"
	"github.com/bhojpur/dcp/pkg/cloud/cli/etcdsnapshot"
	"github.com/bhojpur/dcp/pkg/cloud/cli/kubectl"
	"github.com/bhojpur/dcp/pkg/cloud/cli/secretsencrypt"
	"github.com/bhojpur/dcp/pkg/cloud/cli/server"
	"github.com/bhojpur/dcp/pkg/cloud/configfilearg"
	"github.com/bhojpur/dcp/pkg/cloud/containerd"
	ctr2 "github.com/bhojpur/dcp/pkg/cloud/ctr"
	kubectl2 "github.com/bhojpur/dcp/pkg/cloud/kubectl"
	"github.com/docker/docker/pkg/reexec"
	crictl2 "github.com/bhojpur/dcp/pkg/cloud/cli/crictl"
	"github.com/bhojpur/host/pkg/machine/log"
	"github.com/bhojpur/dcp/pkg/version"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func init() {
	reexec.Register("containerd", containerd.Main)
	reexec.Register("kubectl", kubectl2.Main)
	reexec.Register("crictl", crictl2.Main)
	reexec.Register("ctr", ctr2.Main)
}

func main() {
	cmd := os.Args[0]
	os.Args[0] = filepath.Base(os.Args[0])
	if reexec.Init() {
		return
	}
	os.Args[0] = cmd

	app := cmds.NewApp()
	app.Author = "Bhojpur Consulting Private Limited, India"
	app.Email = "https://www.bhojpur-consulting.com"

	app.Commands = []cli.Command{
		cmds.NewServerCommand(server.Run),
		cmds.NewAgentCommand(agent.Run),
		cmds.NewKubectlCommand(kubectl.Run),
		cmds.NewCRICTL(crictl.Run),
		cmds.NewCtrCommand(ctr.Run),
		cmds.NewEtcdSnapshotCommand(etcdsnapshot.Save,
			cmds.NewEtcdSnapshotSubcommands(
				etcdsnapshot.Delete,
				etcdsnapshot.List,
				etcdsnapshot.Prune,
				etcdsnapshot.Save),
		),
		cmds.NewSecretsEncryptCommand(cli.ShowAppHelp,
			cmds.NewSecretsEncryptSubcommands(
				secretsencrypt.Status,
				secretsencrypt.Enable,
				secretsencrypt.Disable,
				secretsencrypt.Prepare,
				secretsencrypt.Rotate,
				secretsencrypt.Reencrypt),
		),
		cmds.NewCertCommand(
			cmds.NewCertSubcommands(
				cert.Run),
		),
	}
	app.CommandNotFound = cmdNotFound

	if err := app.Run(configfilearg.MustParse(os.Args)); err != nil && !errors.Is(err, context.Canceled) {
		logrus.Fatal(err)
	}
}

func cmdNotFound(c *cli.Context, command string) {
	log.Errorf(
		"%s: '%s' is not a %s command. See '%s --help'.",
		c.App.Name,
		command,
		c.App.Name,
		os.Args[0],
	)
	os.Exit(1)
}