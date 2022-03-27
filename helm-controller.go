//go:build helm-contoller
// +build helm-contoller

//go:generate go run pkg/helm-contoller/codegen/cleanup/main.go
//go:generate /bin/rm -rf pkg/helm-contoller/generated
//go:generate go run pkg/helm-contoller/codegen/main.go

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
	"os"

	helmv1 "github.com/bhojpur/dcp/pkg/helm-controller/generated/controllers/helm.bhojpur.net"
	helmcontroller "github.com/bhojpur/dcp/pkg/helm-controller/helm"
	"github.com/rancher/wrangler/pkg/apply"
	batchv1 "github.com/rancher/wrangler/pkg/generated/controllers/batch"
	corev1 "github.com/rancher/wrangler/pkg/generated/controllers/core"
	rbacv1 "github.com/rancher/wrangler/pkg/generated/controllers/rbac"
	"github.com/rancher/wrangler/pkg/signals"
	"github.com/rancher/wrangler/pkg/start"
	"github.com/urfave/cli"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

var (
	VERSION = "v0.0.0-dev"
)

func main() {
	app := cli.NewApp()
	app.Name = "helm-controller"
	app.Version = VERSION
	app.Usage = "Helm Controller, to help with Helm deployments. Options kubeconfig or masterurl are required."
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

	klog.Infof("Starting helm controller with %d threads.", threadiness)

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
