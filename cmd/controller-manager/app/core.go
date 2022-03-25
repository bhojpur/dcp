package app

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
	"net/http"
	"time"

	"github.com/bhojpur/dcp/pkg/controller/certificates"
	lifecyclecontroller "github.com/bhojpur/dcp/pkg/controller/nodelifecycle"
)

func startNodeLifecycleController(ctx ControllerContext) (http.Handler, bool, error) {
	lifecycleController, err := lifecyclecontroller.NewNodeLifecycleController(
		ctx.InformerFactory.Coordination().V1().Leases(),
		ctx.InformerFactory.Core().V1().Pods(),
		ctx.InformerFactory.Core().V1().Nodes(),
		ctx.InformerFactory.Apps().V1().DaemonSets(),
		// node lifecycle controller uses existing cluster role from node-controller
		ctx.ClientBuilder.ClientOrDie("node-controller"),
		//ctx.ComponentConfig.KubeCloudShared.NodeMonitorPeriod.Duration,
		5*time.Second,
		ctx.ComponentConfig.NodeLifecycleController.NodeStartupGracePeriod.Duration,
		ctx.ComponentConfig.NodeLifecycleController.NodeMonitorGracePeriod.Duration,
		ctx.ComponentConfig.NodeLifecycleController.PodEvictionTimeout.Duration,
		ctx.ComponentConfig.NodeLifecycleController.NodeEvictionRate,
		ctx.ComponentConfig.NodeLifecycleController.SecondaryNodeEvictionRate,
		ctx.ComponentConfig.NodeLifecycleController.LargeClusterSizeThreshold,
		ctx.ComponentConfig.NodeLifecycleController.UnhealthyZoneThreshold,
		*ctx.ComponentConfig.NodeLifecycleController.EnableTaintManager,
	)
	if err != nil {
		return nil, true, err
	}
	go lifecycleController.Run(ctx.Stop)
	return nil, true, nil
}

func startDcpCSRApproverController(ctx ControllerContext) (http.Handler, bool, error) {
	clientSet := ctx.ClientBuilder.ClientOrDie("csr-controller")
	csrApprover, err := certificates.NewCSRApprover(clientSet, ctx.InformerFactory)
	if err != nil {
		return nil, false, err
	}
	go csrApprover.Run(certificates.DcpCSRApproverThreadiness, ctx.Stop)

	return nil, true, nil
}
