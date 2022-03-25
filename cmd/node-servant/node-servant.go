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
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/bhojpur/dcp/cmd/node-servant/convert"
	preflightconvert "github.com/bhojpur/dcp/cmd/node-servant/preflight-convert"
	"github.com/bhojpur/dcp/cmd/node-servant/revert"
	"github.com/bhojpur/dcp/pkg/projectinfo"
)

// node-servant
// running on specific node, do convert/revert job
// client convert/revert join/reset, cluster operator shall start a k8s job to run this.
func main() {
	rand.Seed(time.Now().UnixNano())

	version := fmt.Sprintf("%#v", projectinfo.Get())
	rootCmd := &cobra.Command{
		Use:     "node-servant",
		Short:   "node-servant do preflight-convert/convert/revert specific node",
		Version: version,
	}
	rootCmd.PersistentFlags().String("kubeconfig", "", "The path to the kubeconfig file")
	rootCmd.AddCommand(convert.NewConvertCmd())
	rootCmd.AddCommand(revert.NewRevertCmd())
	rootCmd.AddCommand(preflightconvert.NewxPreflightConvertCmd())

	if err := rootCmd.Execute(); err != nil { // run command
		os.Exit(1)
	}
}
