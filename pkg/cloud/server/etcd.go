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
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/bhojpur/dcp/pkg/cloud/etcd"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// setETCDLabelsAndAnnotations will set the etcd role label if not exists also it
// sets special annotations on the node object which are etcd node id and etcd node
// address, the function will also remove the controlplane and master role labels if
// they exist on the node
func setETCDLabelsAndAnnotations(ctx context.Context, config *Config) error {
	<-config.ControlConfig.Runtime.APIServerReady
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for range t.C {
		controlConfig := &config.ControlConfig

		sc, err := NewContext(ctx, controlConfig.Runtime.KubeConfigAdmin)
		if err != nil {
			logrus.Infof("Failed to set etcd role label: %v", err)
			continue
		}

		if err := sc.Start(ctx); err != nil {
			logrus.Infof("Failed to set etcd role label: %v", err)
			continue
		}

		controlConfig.Runtime.Core = sc.Core
		nodes := sc.Core.Core().V1().Node()

		nodeName := os.Getenv("NODE_NAME")
		if nodeName == "" {
			logrus.Info("Failed to set etcd role label: node name not set")
			continue
		}
		node, err := nodes.Get(nodeName, metav1.GetOptions{})
		if err != nil {
			logrus.Infof("Failed to set etcd role label: %v", err)
			continue
		}

		if node.Labels == nil {
			node.Labels = make(map[string]string)
		}

		// remove controlplane label if role label exists
		var controlRoleLabelExists bool
		if _, ok := node.Labels[MasterRoleLabelKey]; ok {
			delete(node.Labels, MasterRoleLabelKey)
			controlRoleLabelExists = true
		}
		if _, ok := node.Labels[ControlPlaneRoleLabelKey]; ok {
			delete(node.Labels, ControlPlaneRoleLabelKey)
			controlRoleLabelExists = true
		}

		if v, ok := node.Labels[ETCDRoleLabelKey]; ok && v == "true" && !controlRoleLabelExists {
			break
		}

		node.Labels[ETCDRoleLabelKey] = "true"

		// this is replacement to the etcd controller handleself function
		if node.Annotations == nil {
			node.Annotations = map[string]string{}
		}
		fileName := filepath.Join(controlConfig.DataDir, "db", "etcd", "name")

		data, err := ioutil.ReadFile(fileName)
		if err != nil {
			logrus.Infof("Waiting for etcd node name file to be available: %v", err)
			continue
		}
		etcdNodeName := string(data)
		node.Annotations[etcd.NodeNameAnnotation] = etcdNodeName

		address, err := etcd.GetAdvertiseAddress(controlConfig.PrivateIP)
		if err != nil {
			logrus.Infof("Waiting for etcd node address to be available: %v", err)
			continue
		}
		node.Annotations[etcd.NodeAddressAnnotation] = address

		_, err = nodes.Update(node)
		if err == nil {
			logrus.Infof("Successfully set etcd role label and annotations on node %s", nodeName)
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}
