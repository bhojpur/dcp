package etcd

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
	"time"

	controllerv1 "github.com/rancher/wrangler/pkg/generated/controllers/core/v1"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

func RegisterMetadataHandlers(ctx context.Context, etcd *ETCD, nodes controllerv1.NodeController) {
	h := &metadataHandler{
		etcd:           etcd,
		nodeController: nodes,
		ctx:            ctx,
	}
	nodes.OnChange(ctx, "managed-etcd-metadata-controller", h.sync)
}

type metadataHandler struct {
	etcd           *ETCD
	nodeController controllerv1.NodeController
	ctx            context.Context
}

func (m *metadataHandler) sync(key string, node *v1.Node) (*v1.Node, error) {
	if node == nil {
		return nil, nil
	}

	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		logrus.Debug("waiting for node to be assigned for etcd controller")
		m.nodeController.EnqueueAfter(key, 5*time.Second)
		return node, nil
	}

	if key == nodeName {
		return m.handleSelf(node)
	}

	return node, nil
}

func (m *metadataHandler) handleSelf(node *v1.Node) (*v1.Node, error) {
	if node.Annotations[NodeNameAnnotation] == m.etcd.name &&
		node.Annotations[NodeAddressAnnotation] == m.etcd.address &&
		node.Labels[EtcdRoleLabel] == "true" &&
		node.Labels[ControlPlaneLabel] == "true" ||
		m.etcd.config.DisableETCD {
		return node, nil
	}

	node = node.DeepCopy()
	if node.Annotations == nil {
		node.Annotations = map[string]string{}
	}

	node.Annotations[NodeNameAnnotation] = m.etcd.name
	node.Annotations[NodeAddressAnnotation] = m.etcd.address
	node.Labels[EtcdRoleLabel] = "true"
	node.Labels[MasterLabel] = "true"
	node.Labels[ControlPlaneLabel] = "true"

	return m.nodeController.Update(node)
}
