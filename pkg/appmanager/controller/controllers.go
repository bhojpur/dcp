package controller

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

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	dcpappdaemon "github.com/bhojpur/dcp/pkg/appmanager/controller/dcpappdaemon"
	dcpingress "github.com/bhojpur/dcp/pkg/appmanager/controller/dcpingress"
	"github.com/bhojpur/dcp/pkg/appmanager/controller/nodepool"
	"github.com/bhojpur/dcp/pkg/appmanager/controller/uniteddeployment"
)

var controllerAddFuncs []func(manager.Manager, context.Context) error

func init() {
	controllerAddFuncs = append(controllerAddFuncs, uniteddeployment.Add, nodepool.Add, dcpappdaemon.Add, dcpingress.Add)
}

func SetupWithManager(m manager.Manager, ctx context.Context) error {
	for _, f := range controllerAddFuncs {
		if err := f(m, ctx); err != nil {
			if kindMatchErr, ok := err.(*meta.NoKindMatchError); ok {
				klog.Infof("CRD %v is not installed, its controller will perform noops!", kindMatchErr.GroupKind)
				continue
			}
			return err
		}
	}
	return nil
}
