package cmds

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
	"github.com/bhojpur/dcp/pkg/cloud/version"
	"github.com/urfave/cli"
)

const CertCommand = "certificate"

var (
	ServicesList     cli.StringSlice
	CertCommandFlags = []cli.Flag{
		DebugFlag,
		ConfigFlag,
		LogFile,
		AlsoLogToStderr,
		cli.StringFlag{
			Name:        "data-dir,d",
			Usage:       "(data) Folder to hold state default /var/lib/bhojpur/" + version.Program + " or ${HOME}/.bhojpur/" + version.Program + " if not root",
			Destination: &ServerConfig.DataDir,
		},
		cli.StringSliceFlag{
			Name:  "service,s",
			Usage: "List of services to rotate certificates for. Options include (admin, api-server, controller-manager, scheduler, " + version.Program + "-controller, " + version.Program + "-server, cloud-controller, etcd, auth-proxy, kubelet, kube-proxy)",
			Value: &ServicesList,
		},
	}
)

func NewCertCommand(subcommands []cli.Command) cli.Command {
	return cli.Command{
		Name:            CertCommand,
		Usage:           "Certificates management",
		SkipFlagParsing: false,
		SkipArgReorder:  true,
		Subcommands:     subcommands,
		Flags:           CertCommandFlags,
	}
}

func NewCertSubcommands(rotate func(ctx *cli.Context) error) []cli.Command {
	return []cli.Command{
		{
			Name:            "rotate",
			Usage:           "Certificate rotation",
			SkipFlagParsing: false,
			SkipArgReorder:  true,
			Action:          rotate,
			Flags:           CertCommandFlags,
		},
	}
}
