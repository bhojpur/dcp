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
	"fmt"
	"os"
	"runtime"

	"github.com/bhojpur/dcp/pkg/cloud/statebase/endpoint"
	"github.com/bhojpur/dcp/pkg/cloud/statebase/version"
	"github.com/bhojpur/host/pkg/common/signals"
	"github.com/bhojpur/host/pkg/machine/log"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	config endpoint.Config
)

func main() {
	app := cli.NewApp()
	app.Name = "statebase"

	app.Author = "Bhojpur Consulting Private Limited, India"
	app.Email = "https://www.bhojpur-consulting.com"

	app.Usage = "Minimal etcd v3 API to support custom Kubernetes storage engines"
	app.Version = fmt.Sprintf("%s (%s)", version.Version, version.GitCommit)
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("%s version %s\n", app.Name, app.Version)
		fmt.Printf("Go version %s\n", runtime.Version())
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "listen-address",
			Value:       "0.0.0.0:2379",
			Destination: &config.Listener,
		},
		cli.StringFlag{
			Name:        "endpoint",
			Usage:       "Storage endpoint (default is sqlite)",
			Destination: &config.Endpoint,
		},
		cli.StringFlag{
			Name:        "ca-file",
			Usage:       "CA cert for DB connection",
			Destination: &config.BackendTLSConfig.CAFile,
		},
		cli.StringFlag{
			Name:        "cert-file",
			Usage:       "Certificate for DB connection",
			Destination: &config.BackendTLSConfig.CertFile,
		},
		cli.StringFlag{
			Name:        "server-cert-file",
			Usage:       "Certificate for etcd connection",
			Destination: &config.ServerTLSConfig.CertFile,
		},
		cli.StringFlag{
			Name:        "server-key-file",
			Usage:       "Key file for etcd connection",
			Destination: &config.ServerTLSConfig.KeyFile,
		},
		cli.IntFlag{
			Name:        "datastore-max-idle-connections",
			Usage:       "Maximum number of idle connections retained by datastore. If value = 0, the system default will be used. If value < 0, idle connections will not be reused.",
			Destination: &config.ConnectionPoolConfig.MaxIdle,
			Value:       0,
		},
		cli.IntFlag{
			Name:        "datastore-max-open-connections",
			Usage:       "Maximum number of open connections used by datastore. If value <= 0, then there is no limit",
			Destination: &config.ConnectionPoolConfig.MaxOpen,
			Value:       0,
		},
		cli.DurationFlag{
			Name:        "datastore-connection-max-lifetime",
			Usage:       "Maximum amount of time a connection may be reused. If value <= 0, then there is no limit.",
			Destination: &config.ConnectionPoolConfig.MaxLifetime,
			Value:       0,
		},
		cli.StringFlag{
			Name:        "key-file",
			Usage:       "Key file for DB connection",
			Destination: &config.BackendTLSConfig.KeyFile,
		},
		cli.BoolFlag{Name: "debug"},
	}
	app.Action = run
	app.CommandNotFound = cmdNotFound

	if err := app.Run(os.Args); err != nil {
		if !errors.Is(err, context.Canceled) {
			logrus.Fatal(err)
		}
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

func run(c *cli.Context) error {
	if c.Bool("debug") {
		logrus.SetLevel(logrus.TraceLevel)
	}
	ctx := signals.SetupSignalHandler()
	_, err := endpoint.Listen(context.Background(), config)
	if err != nil {
		return err
	}

	log.Debugf("%s", ctx)
	return err
}
