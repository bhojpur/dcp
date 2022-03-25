package kubectl

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
	"strings"
	"time"

	"github.com/bhojpur/dcp/pkg/cloud/server"
	"github.com/sirupsen/logrus"
	"k8s.io/component-base/cli"
	"k8s.io/kubectl/pkg/cmd"
	"k8s.io/kubectl/pkg/cmd/util"
)

func Main() {
	kubenv := os.Getenv("KUBECONFIG")
	for i, arg := range os.Args {
		if strings.HasPrefix(arg, "--kubeconfig=") {
			kubenv = strings.Split(arg, "=")[1]
		} else if strings.HasPrefix(arg, "--kubeconfig") && i+1 < len(os.Args) {
			kubenv = os.Args[i+1]
		}
	}
	if kubenv == "" {
		config, err := server.HomeKubeConfig(false, false)
		if _, serr := os.Stat(config); err == nil && serr == nil {
			os.Setenv("KUBECONFIG", config)
		}
		if err := checkReadConfigPermissions(config); err != nil {
			logrus.Warn(err)
		}
	}

	main()
}

func main() {
	rand.Seed(time.Now().UnixNano())

	command := cmd.NewDefaultKubectlCommand()
	if err := cli.RunNoErrOutput(command); err != nil {
		util.CheckErr(err)
	}
}

func checkReadConfigPermissions(configFile string) error {
	file, err := os.OpenFile(configFile, os.O_RDONLY, 0600)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("Unable to read %s, please start server "+
				"with --write-kubeconfig-mode to modify kube config permissions", configFile)
		}
	}
	file.Close()
	return nil
}
