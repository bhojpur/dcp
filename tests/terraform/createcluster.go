package e2e

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
	"flag"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
)

var destroy = flag.Bool("destroy", false, "a bool")
var nodeOs = flag.String("node_os", "centos8", "a string")
var externalDb = flag.String("external_db", "mysql", "a string")
var arch = flag.String("arch", "amd64", "a string")
var clusterType = flag.String("cluster_type", "etcd", "a string")
var resourceName = flag.String("resource_name", "etcd", "a string")
var sshuser = flag.String("sshuser", "ubuntu", "a string")
var sshkey = flag.String("sshkey", "", "a string")

var (
	kubeConfigFile string
	masterIPs      string
	workerIPs      string
)

func BuildCluster(nodeOs, clusterType, externalDb, resourceName string, t *testing.T, destroy bool) (string, string, string, error) {

	tDir := "./modules/dcpcluster"
	vDir := "/config/" + nodeOs + clusterType + ".tfvars"

	if externalDb != "" {
		vDir = "/config/" + nodeOs + externalDb + ".tfvars"
	}

	tfDir, err := filepath.Abs(tDir)
	if err != nil {
		return "", "", "", err
	}
	varDir, err := filepath.Abs(vDir)
	if err != nil {
		return "", "", "", err
	}
	TerraformOptions := &terraform.Options{
		TerraformDir: tfDir,
		VarFiles:     []string{varDir},
		Vars: map[string]interface{}{
			"cluster_type":  clusterType,
			"resource_name": resourceName,
			"external_db":   externalDb,
		},
	}

	if destroy {
		fmt.Printf("Cluster is being deleted")
		terraform.Destroy(t, TerraformOptions)
		return "", "", "", err
	}

	fmt.Printf("Creating Cluster")
	terraform.InitAndApply(t, TerraformOptions)
	kubeconfig := terraform.Output(t, TerraformOptions, "kubeconfig") + "_kubeconfig"
	masterIPs := terraform.Output(t, TerraformOptions, "master_ips")
	workerIPs := terraform.Output(t, TerraformOptions, "worker_ips")
	kubeconfigFile := "/config/" + kubeconfig
	return kubeconfigFile, masterIPs, workerIPs, err
}