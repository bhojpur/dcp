package apiaddresses

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
	"bytes"
	"context"
	"encoding/json"

	"github.com/bhojpur/dcp/pkg/cloud/daemons/config"
	"github.com/bhojpur/dcp/pkg/cloud/etcd"
	"github.com/bhojpur/dcp/pkg/cloud/util"
	"github.com/bhojpur/dcp/pkg/cloud/version"
	controllerv1 "github.com/rancher/wrangler/pkg/generated/controllers/core/v1"
	clientv3 "go.etcd.io/etcd/client/v3"
	v1 "k8s.io/api/core/v1"
)

type EndpointsControllerGetter func() controllerv1.EndpointsController

func Register(ctx context.Context, runtime *config.ControlRuntime, endpoints controllerv1.EndpointsController) error {
	h := &handler{
		endpointsController: endpoints,
		runtime:             runtime,
		ctx:                 ctx,
	}
	endpoints.OnChange(ctx, version.Program+"-apiserver-lb-controller", h.sync)

	cl, err := etcd.GetClient(h.ctx, h.runtime, "https://127.0.0.1:2379")
	if err != nil {
		return err
	}
	h.etcdClient = cl

	go func() {
		<-ctx.Done()
		h.etcdClient.Close()
	}()

	return nil
}

type handler struct {
	endpointsController controllerv1.EndpointsController
	runtime             *config.ControlRuntime
	ctx                 context.Context
	etcdClient          *clientv3.Client
}

// This controller will update the version.program/apiaddresses etcd key with a list of
// api addresses endpoints found in the kubernetes service in the default namespace
func (h *handler) sync(key string, endpoint *v1.Endpoints) (*v1.Endpoints, error) {
	if endpoint != nil &&
		endpoint.Namespace == "default" &&
		endpoint.Name == "kubernetes" {
		w := &bytes.Buffer{}
		if err := json.NewEncoder(w).Encode(util.GetAddresses(endpoint)); err != nil {
			return nil, err
		}
		_, err := h.etcdClient.Put(h.ctx, etcd.AddressKey, w.String())
		if err != nil {
			return nil, err
		}
	}
	return endpoint, nil
}
