package server

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
	"runtime"

	"github.com/bhojpur/dcp/pkg/cloud/deploy"
	"github.com/bhojpur/dcp/pkg/cloud/version"
	"github.com/bhojpur/dcp/pkg/generated/controllers/dcp.bhojpur.net"
	"github.com/bhojpur/dcp/pkg/generated/controllers/helm.bhojpur.net"
	"github.com/bhojpur/host/pkg/common/apply"
	"github.com/bhojpur/host/pkg/common/crd"
	"github.com/bhojpur/host/pkg/common/start"
	"github.com/bhojpur/host/pkg/generated/controllers/apps"
	"github.com/bhojpur/host/pkg/generated/controllers/batch"
	"github.com/bhojpur/host/pkg/generated/controllers/core"
	"github.com/bhojpur/host/pkg/generated/controllers/rbac"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Context struct {
	Dcp   *dcp.Factory
	Helm  *helm.Factory
	Batch *batch.Factory
	Apps  *apps.Factory
	Auth  *rbac.Factory
	Core  *core.Factory
	K8s   kubernetes.Interface
	Apply apply.Apply
}

func (c *Context) Start(ctx context.Context) error {
	return start.All(ctx, 5, c.Dcp, c.Helm, c.Apps, c.Auth, c.Batch, c.Core)
}

func NewContext(ctx context.Context, cfg string) (*Context, error) {
	restConfig, err := clientcmd.BuildConfigFromFlags("", cfg)
	if err != nil {
		return nil, err
	}

	// Construct a custom user-agent string for the apply client used by the deploy controller
	// so that we can track which node's deploy controller most recently modified a resource.
	nodeName := os.Getenv("NODE_NAME")
	managerName := deploy.ControllerName + "@" + nodeName
	if nodeName == "" || len(managerName) > validation.FieldManagerMaxLength {
		logrus.Warn("Deploy controller node name is empty or too long, and will not be tracked via server side apply field management")
		managerName = deploy.ControllerName
	}
	restConfig.UserAgent = fmt.Sprintf("%s/%s (%s/%s) %s/%s", managerName, version.Version, runtime.GOOS, runtime.GOARCH, version.Program, version.GitCommit)

	if err := crds(ctx, restConfig); err != nil {
		return nil, errors.Wrap(err, "failed to register CRDs")
	}

	k8s, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return &Context{
		Dcp:   dcp.NewFactoryFromConfigOrDie(restConfig),
		Helm:  helm.NewFactoryFromConfigOrDie(restConfig),
		K8s:   k8s,
		Auth:  rbac.NewFactoryFromConfigOrDie(restConfig),
		Apps:  apps.NewFactoryFromConfigOrDie(restConfig),
		Batch: batch.NewFactoryFromConfigOrDie(restConfig),
		Core:  core.NewFactoryFromConfigOrDie(restConfig),
		Apply: apply.New(k8s, apply.NewClientFactory(restConfig)).WithDynamicLookup(),
	}, nil
}

func crds(ctx context.Context, config *rest.Config) error {
	factory, err := crd.NewFactoryFromClient(config)
	if err != nil {
		return err
	}

	factory.BatchCreateCRDs(ctx, crd.NamespacedTypes(
		"Addon.dcp.bhojpur.net/v1",
		"HelmChart.helm.bhojpur.net/v1",
		"HelmChartConfig.helm.bhojpur.net/v1")...)

	return factory.BatchWait()
}
