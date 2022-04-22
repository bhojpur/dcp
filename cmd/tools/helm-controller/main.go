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
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	helmv1 "github.com/bhojpur/dcp/pkg/generated/controllers/helm.bhojpur.net"
	helmcontroller "github.com/bhojpur/dcp/pkg/helm-controller/helm"
	"github.com/bhojpur/dcp/pkg/version"
	"github.com/bhojpur/host/pkg/common/apply"
	"github.com/bhojpur/host/pkg/common/signals"
	"github.com/bhojpur/host/pkg/common/start"
	batchv1 "github.com/bhojpur/host/pkg/generated/controllers/batch"
	corev1 "github.com/bhojpur/host/pkg/generated/controllers/core"
	rbacv1 "github.com/bhojpur/host/pkg/generated/controllers/rbac"
	"github.com/bhojpur/host/pkg/machine/log"
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

var released = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)

var appHelpTemplate = `Usage: {{.Name}} {{if .Flags}}[OPTIONS] {{end}}COMMAND [arg...]
{{.Usage}}
Version: {{.Version}}{{if or .Author .Email}}
Author:{{if .Author}}
  {{.Author}}{{if .Email}} - <{{.Email}}>{{end}}{{else}}
  {{.Email}}{{end}}{{end}}
{{if .Flags}}
Options:
  {{range .Flags}}{{.}}
  {{end}}{{end}}
Commands:
  {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
  {{end}}
Run '{{.Name}} COMMAND --help' for more information on a command.
`

var commandHelpTemplate = `Usage: helmctl {{.Name}}{{if .Flags}} [OPTIONS]{{end}} [arg...]
{{.Usage}}{{if .Description}}
Description:
   {{.Description}}{{end}}{{if .Flags}}
Options:
   {{range .Flags}}
   {{.}}{{end}}{{ end }}
`

func setDebugOutputLevel() {
	// check -D, --debug and -debug, if set force debug and env var
	for _, f := range os.Args {
		if f == "-D" || f == "--debug" || f == "-debug" {
			os.Setenv("BHOJPUR_HELM_DEBUG", "1")
			log.SetDebug(true)
			return
		}
	}

	// check env
	debugEnv := os.Getenv("BHOJPUR_HELM_DEBUG")
	if debugEnv != "" {
		showDebug, err := strconv.ParseBool(debugEnv)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing boolean value from BHOJPUR_HELM_DEBUG: %s\n", err)
			os.Exit(1)
		}
		log.SetDebug(showDebug)
	}
}

func main() {
	cli.AppHelpTemplate = appHelpTemplate
	cli.CommandHelpTemplate = commandHelpTemplate

	logrus.SetOutput(colorable.NewColorableStdout())
	setDebugOutputLevel()

	if err := mainErr(); err != nil {
		logrus.Fatal(err)
	}
}

func mainErr() error {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Author = "Bhojpur Consulting Private Limited, India"
	app.Email = "https://www.bhojpur-consulting.com"

	app.Usage = "Helm Controller, to help with Helm deployments. Options kubeconfig or masterurl are required."
	app.Version = version.FullVersion()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "kubeconfig, k",
			EnvVar: "KUBECONFIG",
			Value:  "",
			Usage:  "Kubernetes config files, e.g. $HOME/.kube/config",
		},
		cli.StringFlag{
			Name:   "master, m",
			EnvVar: "MASTERURL",
			Value:  "",
			Usage:  "Kubernetes cluster master URL.",
		},
		cli.StringFlag{
			Name:   "namespace, n",
			EnvVar: "NAMESPACE",
			Value:  "",
			Usage:  "Namespace to watch, empty means it will watch CRDs in all namespaces.",
		},
		cli.IntFlag{
			Name:   "threads, t",
			EnvVar: "THREADS",
			Value:  2,
			Usage:  "Threadiness level to set, defaults to 2.",
		},
	}
	app.Action = run

	if err := app.Run(os.Args); err != nil {
		klog.Fatal(err)
	}
	return nil
}

func run(c *cli.Context) error {
	masterURL := c.String("master")
	kubeconfig := c.String("kubeconfig")
	namespace := c.String("namespace")
	threadiness := c.Int("threads")

	if threadiness <= 0 {
		klog.Infof("Can not start with thread count of %d, please pass a proper thread count.", threadiness)
		return nil
	}

	klog.Infof("Starting Bhojpur Helm Controller with %d threads.", threadiness)

	if namespace == "" {
		klog.Info("Starting helm controller with no namespace.")
	} else {
		klog.Infof("Starting helm controller in namespace: %s.", namespace)
	}

	ctx := signals.SetupSignalHandler(context.Background())

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Fatalf("Error building config from flags: %s", err.Error())
	}

	helms, err := helmv1.NewFactoryFromConfigWithNamespace(cfg, namespace)
	if err != nil {
		klog.Fatalf("Error building sample controllers: %s", err.Error())
	}

	batches, err := batchv1.NewFactoryFromConfigWithNamespace(cfg, namespace)
	if err != nil {
		klog.Fatalf("Error building sample controllers: %s", err.Error())
	}

	rbacs, err := rbacv1.NewFactoryFromConfigWithNamespace(cfg, namespace)
	if err != nil {
		klog.Fatalf("Error building sample controllers: %s", err.Error())
	}

	cores, err := corev1.NewFactoryFromConfigWithNamespace(cfg, namespace)
	if err != nil {
		klog.Fatalf("Error building sample controllers: %s", err.Error())
	}

	k8sClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes client: %s", err.Error())
	}

	discoverClient, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building discovery client: %s", err.Error())
	}

	objectSetApply := apply.New(discoverClient, apply.NewClientFactory(cfg))

	helmcontroller.Register(ctx,
		k8sClient,
		objectSetApply,
		helms.Helm().V1().HelmChart(),
		helms.Helm().V1().HelmChartConfig(),
		batches.Batch().V1().Job(),
		rbacs.Rbac().V1().ClusterRoleBinding(),
		cores.Core().V1().ServiceAccount(),
		cores.Core().V1().ConfigMap())

	if err := start.All(ctx, threadiness, helms, batches, rbacs, cores); err != nil {
		klog.Fatalf("Error starting: %s", err.Error())
	}

	<-ctx.Done()
	return nil
}
