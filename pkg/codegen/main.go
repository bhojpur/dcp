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
	"os"

	dcpv1 "github.com/bhojpur/dcp/pkg/apis/dcp.bhojpur.net/v1"
	helmv1 "github.com/bhojpur/dcp/pkg/apis/helm.bhojpur.net/v1"
	controllergen "github.com/bhojpur/host/pkg/common/controller-gen"
	"github.com/bhojpur/host/pkg/common/controller-gen/args"
	bindata "github.com/go-bindata/go-bindata"
	"github.com/sirupsen/logrus"
)

var (
	basePackage = "github.com/bhojpur/dcp/pkg/cloud/types"
)

func main() {
	os.Unsetenv("GOPATH")
	bc := &bindata.Config{
		Input: []bindata.InputConfig{
			{
				Path:      "build/data",
				Recursive: true,
			},
		},
		Package:    "data",
		NoCompress: true,
		NoMemCopy:  true,
		NoMetadata: true,
		Output:     "pkg/cloud/data/zz_generated_bindata.go",
	}
	if err := bindata.Translate(bc); err != nil {
		logrus.Fatal(err)
	}

	bc = &bindata.Config{
		Input: []bindata.InputConfig{
			{
				Path:      "manifests",
				Recursive: true,
			},
		},
		Package:    "deploy",
		NoMetadata: true,
		Prefix:     "manifests/",
		Output:     "pkg/cloud/deploy/zz_generated_bindata.go",
		Tags:       "!no_stage",
	}
	if err := bindata.Translate(bc); err != nil {
		logrus.Fatal(err)
	}

	bc = &bindata.Config{
		Input: []bindata.InputConfig{
			{
				Path:      "build/static",
				Recursive: true,
			},
		},
		Package:    "static",
		NoMetadata: true,
		Prefix:     "build/static/",
		Output:     "pkg/cloud/static/zz_generated_bindata.go",
		Tags:       "!no_stage",
	}
	if err := bindata.Translate(bc); err != nil {
		logrus.Fatal(err)
	}

	controllergen.Run(args.Options{
		OutputPackage: "github.com/bhojpur/dcp/pkg/generated",
		Boilerplate:   "scripts/boilerplate.go.txt",
		Groups: map[string]args.Group{
			"dcp.bhojpur.net": {
				Types: []interface{}{
					dcpv1.Addon{},
				},
				GenerateTypes:   true,
				GenerateClients: true,
			},
		},
		"helm.bhojpur.net": {
			Types: []interface{}{
				helmv1.HelmChart{},
				helmv1.HelmChartConfig{},
			},
			GenerateTypes:   true,
			GenerateClients: true,
		},
	})
}
