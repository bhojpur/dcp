package node

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
	"strings"

	"github.com/bhojpur/dcp/pkg/cloud/nodepassword"
	"github.com/pkg/errors"
	coreclient "github.com/rancher/wrangler/pkg/generated/controllers/core/v1"
	"github.com/sirupsen/logrus"
	core "k8s.io/api/core/v1"
)

func Register(ctx context.Context,
	modCoreDNS bool,
	secretClient coreclient.SecretClient,
	configMap coreclient.ConfigMapController,
	nodes coreclient.NodeController,
) error {
	h := &handler{
		modCoreDNS:   modCoreDNS,
		secretClient: secretClient,
		configCache:  configMap.Cache(),
		configClient: configMap,
	}
	nodes.OnChange(ctx, "node", h.onChange)
	nodes.OnRemove(ctx, "node", h.onRemove)

	return nil
}

type handler struct {
	modCoreDNS   bool
	secretClient coreclient.SecretClient
	configCache  coreclient.ConfigMapCache
	configClient coreclient.ConfigMapClient
}

func (h *handler) onChange(key string, node *core.Node) (*core.Node, error) {
	if node == nil {
		return nil, nil
	}
	return h.updateHosts(node, false)
}

func (h *handler) onRemove(key string, node *core.Node) (*core.Node, error) {
	return h.updateHosts(node, true)
}

func (h *handler) updateHosts(node *core.Node, removed bool) (*core.Node, error) {
	var (
		nodeName    string
		nodeAddress string
	)
	nodeName = node.Name
	for _, address := range node.Status.Addresses {
		if address.Type == "InternalIP" {
			nodeAddress = address.Address
			break
		}
	}
	if removed {
		if err := h.removeNodePassword(nodeName); err != nil {
			logrus.Warn(errors.Wrap(err, "Unable to remove node password"))
		}
	}
	if h.modCoreDNS {
		if err := h.updateCoreDNSConfigMap(nodeName, nodeAddress, removed); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (h *handler) updateCoreDNSConfigMap(nodeName, nodeAddress string, removed bool) error {
	if nodeAddress == "" && !removed {
		logrus.Errorf("No InternalIP found for node " + nodeName)
		return nil
	}

	configMapCache, err := h.configCache.Get("kube-system", "coredns")
	if err != nil || configMapCache == nil {
		logrus.Warn(errors.Wrap(err, "Unable to fetch coredns config map"))
		return nil
	}

	configMap := configMapCache.DeepCopy()
	hosts := configMap.Data["NodeHosts"]
	hostsMap := map[string]string{}

	for _, line := range strings.Split(hosts, "\n") {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			logrus.Warnf("Unknown format for hosts line [%s]", line)
			continue
		}
		ip := fields[0]
		host := fields[1]
		if host == nodeName {
			if removed {
				continue
			}
			if ip == nodeAddress {
				return nil
			}
		}
		hostsMap[host] = ip
	}

	if !removed {
		hostsMap[nodeName] = nodeAddress
	}

	var newHosts string
	for host, ip := range hostsMap {
		newHosts += ip + " " + host + "\n"
	}
	configMap.Data["NodeHosts"] = newHosts

	if _, err := h.configClient.Update(configMap); err != nil {
		return err
	}

	var actionType string
	if removed {
		actionType = "Removed"
	} else {
		actionType = "Updated"
	}
	logrus.Infof("%s coredns node hosts entry [%s]", actionType, nodeAddress+" "+nodeName)
	return nil
}

func (h *handler) removeNodePassword(nodeName string) error {
	return nodepassword.Delete(h.secretClient, nodeName)
}
