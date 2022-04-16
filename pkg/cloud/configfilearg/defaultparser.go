package configfilearg

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
	"github.com/bhojpur/dcp/pkg/cloud/cli/cmds"
	"github.com/bhojpur/dcp/pkg/cloud/version"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var DefaultParser = &Parser{
	After:         []string{"server", "agent", "etcd-snapshot:1"},
	FlagNames:     []string{"--config", "-c"},
	EnvName:       version.ProgramUpper + "_CONFIG_FILE",
	DefaultConfig: "/etc/bhojpur/" + version.Program + "/config.yaml",
	ValidFlags:    map[string][]cli.Flag{"server": cmds.ServerFlags, "etcd-snapshot": cmds.EtcdSnapshotFlags},
}

func MustParse(args []string) []string {
	result, err := DefaultParser.Parse(args)
	if err != nil {
		logrus.Fatal(err)
	}
	return result
}

func MustFindString(args []string, target string) string {
	parser := &Parser{
		After:         []string{},
		FlagNames:     []string{},
		EnvName:       version.ProgramUpper + "_CONFIG_FILE",
		DefaultConfig: "/etc/bhojpur/" + version.Program + "/config.yaml",
	}
	result, err := parser.FindString(args, target)
	if err != nil {
		logrus.Fatal(err)
	}
	return result
}
