package ctr

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

	"github.com/containerd/containerd/cmd/ctr/app"
	"github.com/containerd/containerd/pkg/seed"
	"github.com/urfave/cli"
)

func Main() {
	main()
}

func main() {
	seed.WithTimeAndRand()
	app := app.New()
	for i, flag := range app.Flags {
		if sFlag, ok := flag.(cli.StringFlag); ok {
			if sFlag.Name == "address, a" {
				sFlag.Value = "/run/dcp/containerd/containerd.sock"
				app.Flags[i] = sFlag
			} else if sFlag.Name == "namespace, n" {
				sFlag.Value = "bhojpur.net"
				app.Flags[i] = sFlag
			}
		}
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "ctr: %s\n", err)
		os.Exit(1)
	}
}
