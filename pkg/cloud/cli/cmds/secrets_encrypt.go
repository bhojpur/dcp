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

const SecretsEncryptCommand = "secrets-encrypt"

var EncryptFlags = []cli.Flag{
	DataDirFlag,
	ServerToken,
	cli.StringFlag{
		Name:        "server, s",
		Usage:       "(cluster) Server to connect to",
		EnvVar:      version.ProgramUpper + "_URL",
		Value:       "https://127.0.0.1:6443",
		Destination: &ServerConfig.ServerURL,
	},
}

func NewSecretsEncryptCommand(action func(*cli.Context) error, subcommands []cli.Command) cli.Command {
	return cli.Command{
		Name:            SecretsEncryptCommand,
		Usage:           "Control secrets encryption and keys rotation",
		SkipFlagParsing: false,
		SkipArgReorder:  true,
		Action:          action,
		Subcommands:     subcommands,
	}
}

func NewSecretsEncryptSubcommands(status, enable, disable, prepare, rotate, reencrypt func(ctx *cli.Context) error) []cli.Command {
	return []cli.Command{
		{
			Name:            "status",
			Usage:           "Print current status of secrets encryption",
			SkipFlagParsing: false,
			SkipArgReorder:  true,
			Action:          status,
			Flags: append(EncryptFlags, &cli.StringFlag{
				Name:        "output,o",
				Usage:       "Status format. Default: text. Optional: json",
				Destination: &ServerConfig.EncryptOutput,
			}),
		},
		{
			Name:            "enable",
			Usage:           "Enable secrets encryption",
			SkipFlagParsing: false,
			SkipArgReorder:  true,
			Action:          enable,
			Flags:           EncryptFlags,
		},
		{
			Name:            "disable",
			Usage:           "Disable secrets encryption",
			SkipFlagParsing: false,
			SkipArgReorder:  true,
			Action:          disable,
			Flags:           EncryptFlags,
		},
		{
			Name:            "prepare",
			Usage:           "Prepare for encryption keys rotation",
			SkipFlagParsing: false,
			SkipArgReorder:  true,
			Action:          prepare,
			Flags: append(EncryptFlags, &cli.BoolFlag{
				Name:        "f,force",
				Usage:       "Force preparation.",
				Destination: &ServerConfig.EncryptForce,
			}),
		},
		{
			Name:            "rotate",
			Usage:           "Rotate secrets encryption keys",
			SkipFlagParsing: false,
			SkipArgReorder:  true,
			Action:          rotate,
			Flags: append(EncryptFlags, &cli.BoolFlag{
				Name:        "f,force",
				Usage:       "Force key rotation.",
				Destination: &ServerConfig.EncryptForce,
			}),
		},
		{
			Name:            "reencrypt",
			Usage:           "Reencrypt all data with new encryption key",
			SkipFlagParsing: false,
			SkipArgReorder:  true,
			Action:          reencrypt,
			Flags: append(EncryptFlags,
				&cli.BoolFlag{
					Name:        "f,force",
					Usage:       "Force secrets reencryption.",
					Destination: &ServerConfig.EncryptForce,
				},
				&cli.BoolFlag{
					Name:        "skip",
					Usage:       "Skip removing old key",
					Destination: &ServerConfig.EncryptSkip,
				}),
		},
	}
}
