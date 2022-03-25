package provider

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
	"strings"

	"github.com/bhojpur/dcp/pkg/cloud/version"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	cloudprovider "k8s.io/cloud-provider"
)

var (
	InternalIPKey = version.Program + "bhojpur.net/internal-ip"
	ExternalIPKey = version.Program + "bhojpur.net/external-ip"
	HostnameKey   = version.Program + "bhojpur.net/hostname"
)

func (k *dcp) AddSSHKeyToAllInstances(ctx context.Context, user string, keyData []byte) error {
	return cloudprovider.NotImplemented
}

func (k *dcp) CurrentNodeName(ctx context.Context, hostname string) (types.NodeName, error) {
	return types.NodeName(hostname), nil
}

func (k *dcp) InstanceExistsByProviderID(ctx context.Context, providerID string) (bool, error) {
	return true, nil
}

func (k *dcp) InstanceID(ctx context.Context, nodeName types.NodeName) (string, error) {
	if k.nodeInformerHasSynced == nil || !k.nodeInformerHasSynced() {
		return "", errors.New("Node informer has not synced yet")
	}

	node, err := k.nodeInformer.Lister().Get(string(nodeName))
	if err != nil {
		return "", fmt.Errorf("failed to get node %s: %w", nodeName, err)
	}
	if (node.Annotations[InternalIPKey] == "") && (node.Labels[InternalIPKey] == "") {
		return string(nodeName), errors.New("address annotations not yet set")
	}
	return string(nodeName), nil
}

func (k *dcp) InstanceShutdownByProviderID(ctx context.Context, providerID string) (bool, error) {
	return true, cloudprovider.NotImplemented
}

func (k *dcp) InstanceType(ctx context.Context, name types.NodeName) (string, error) {
	_, err := k.InstanceID(ctx, name)
	if err != nil {
		return "", err
	}
	return version.Program, nil
}

func (k *dcp) InstanceTypeByProviderID(ctx context.Context, providerID string) (string, error) {
	return "", cloudprovider.NotImplemented
}

func (k *dcp) NodeAddresses(ctx context.Context, name types.NodeName) ([]corev1.NodeAddress, error) {
	addresses := []corev1.NodeAddress{}
	if k.nodeInformerHasSynced == nil || !k.nodeInformerHasSynced() {
		return nil, errors.New("Node informer has not synced yet")
	}

	node, err := k.nodeInformer.Lister().Get(string(name))
	if err != nil {
		return nil, fmt.Errorf("Failed to find node %s: %v", name, err)
	}
	// check internal address
	if address := node.Annotations[InternalIPKey]; address != "" {
		for _, v := range strings.Split(address, ",") {
			addresses = append(addresses, corev1.NodeAddress{Type: corev1.NodeInternalIP, Address: v})
		}
	} else if address = node.Labels[InternalIPKey]; address != "" {
		addresses = append(addresses, corev1.NodeAddress{Type: corev1.NodeInternalIP, Address: address})
	} else {
		logrus.Infof("Couldn't find node internal ip annotation or label on node %s", name)
	}

	// check external address
	if address := node.Annotations[ExternalIPKey]; address != "" {
		for _, v := range strings.Split(address, ",") {
			addresses = append(addresses, corev1.NodeAddress{Type: corev1.NodeExternalIP, Address: v})
		}
	} else if address = node.Labels[ExternalIPKey]; address != "" {
		addresses = append(addresses, corev1.NodeAddress{Type: corev1.NodeExternalIP, Address: address})
	}

	// check hostname
	if address := node.Annotations[HostnameKey]; address != "" {
		addresses = append(addresses, corev1.NodeAddress{Type: corev1.NodeHostName, Address: address})
	} else if address = node.Labels[HostnameKey]; address != "" {
		addresses = append(addresses, corev1.NodeAddress{Type: corev1.NodeHostName, Address: address})
	} else {
		logrus.Infof("Couldn't find node hostname annotation or label on node %s", name)
	}

	return addresses, nil
}

func (k *dcp) NodeAddressesByProviderID(ctx context.Context, providerID string) ([]corev1.NodeAddress, error) {
	return nil, cloudprovider.NotImplemented
}
