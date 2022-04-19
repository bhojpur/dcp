package gc

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
	"time"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/cmd/grid/dcpsvr/config"
	"github.com/bhojpur/dcp/pkg/engine/cachemanager"
	"github.com/bhojpur/dcp/pkg/engine/kubernetes/rest"
	"github.com/bhojpur/dcp/pkg/engine/util"
)

var (
	defaultEventGcInterval = 60
)

// GCManager is responsible for cleanup garbage of Bhojpur DCP server engine
type GCManager struct {
	store             cachemanager.StorageWrapper
	restConfigManager *rest.RestConfigManager
	nodeName          string
	eventsGCFrequency time.Duration
	lastTime          time.Time
	stopCh            <-chan struct{}
}

// NewGCManager creates a *GCManager object
func NewGCManager(cfg *config.EngineConfiguration, restConfigManager *rest.RestConfigManager, stopCh <-chan struct{}) (*GCManager, error) {
	gcFrequency := cfg.GCFrequency
	if gcFrequency == 0 {
		gcFrequency = defaultEventGcInterval
	}
	mgr := &GCManager{
		store:             cfg.StorageWrapper,
		nodeName:          cfg.NodeName,
		restConfigManager: restConfigManager,
		eventsGCFrequency: time.Duration(gcFrequency) * time.Minute,
		stopCh:            stopCh,
	}
	_ = mgr.gcPodsWhenRestart()
	return mgr, nil
}

// Run starts GCManager
func (m *GCManager) Run() {
	// run gc events after a time duration between eventsGCFrequency and 3 * eventsGCFrequency
	m.lastTime = time.Now()
	go wait.JitterUntil(func() {
		klog.V(2).Infof("start gc events after waiting %v from previous gc", time.Since(m.lastTime))
		m.lastTime = time.Now()
		cfg := m.restConfigManager.GetRestConfig(true)
		if cfg == nil {
			klog.Errorf("could not get rest config, so skip gc")
			return
		}
		kubeClient, err := clientset.NewForConfig(cfg)
		if err != nil {
			klog.Errorf("could not new kube client, %v", err)
			return
		}

		m.gcEvents(kubeClient, "kubelet")
		m.gcEvents(kubeClient, "kube-proxy")
	}, m.eventsGCFrequency, 2, true, m.stopCh)
}

func (m *GCManager) gcPodsWhenRestart() error {
	localPodKeys, err := m.store.ListKeys("kubelet/pods")
	if err != nil || len(localPodKeys) == 0 {
		return nil
	}
	klog.Infof("list pod keys from storage, total: %d", len(localPodKeys))

	cfg := m.restConfigManager.GetRestConfig(true)
	if cfg == nil {
		klog.Errorf("could not get rest config, so skip gc pods when restart")
		return err
	}
	kubeClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Errorf("could not new kube client, %v", err)
		return err
	}

	listOpts := metav1.ListOptions{FieldSelector: fields.OneTermEqualSelector("spec.nodeName", m.nodeName).String()}
	podList, err := kubeClient.CoreV1().Pods(v1.NamespaceAll).List(context.Background(), listOpts)
	if err != nil {
		klog.Errorf("could not list pods for node(%s), %v", m.nodeName, err)
		return err
	}

	currentPodKeys := make(map[string]struct{}, len(podList.Items))
	for i := range podList.Items {
		name := podList.Items[i].Name
		ns := podList.Items[i].Namespace

		key, _ := util.KeyFunc("kubelet", "pods", ns, name)
		currentPodKeys[key] = struct{}{}
	}
	klog.V(2).Infof("list all of pod that on the node: total: %d", len(currentPodKeys))

	deletedPods := make([]string, 0)
	for i := range localPodKeys {
		if _, ok := currentPodKeys[localPodKeys[i]]; !ok {
			deletedPods = append(deletedPods, localPodKeys[i])
		}
	}

	if len(deletedPods) == len(localPodKeys) {
		klog.Infof("it's dangerous to gc all cache pods, so skip gc")
		return nil
	}

	for _, key := range deletedPods {
		if err := m.store.Delete(key); err != nil {
			klog.Errorf("failed to gc pod %s, %v", key, err)
		} else {
			klog.Infof("gc pod %s successfully", key)
		}
	}

	return nil
}

func (m *GCManager) gcEvents(kubeClient clientset.Interface, component string) {
	if kubeClient == nil {
		return
	}

	localEventKeys, err := m.store.ListKeys(fmt.Sprintf("%s/events", component))
	if err != nil {
		klog.Errorf("could not list keys for %s events, %v", component, err)
		return
	} else if len(localEventKeys) == 0 {
		klog.Infof("no %s events in local storage, skip %s events gc", component, component)
		return
	}
	klog.Infof("list %s event keys from storage, total: %d", component, len(localEventKeys))

	deletedEvents := make([]string, 0)
	for _, key := range localEventKeys {
		_, _, ns, name := util.SplitKey(key)
		if len(ns) == 0 || len(name) == 0 {
			klog.Infof("could not get namespace or name for event %s", key)
			continue
		}

		_, err := kubeClient.CoreV1().Events(ns).Get(context.Background(), name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			deletedEvents = append(deletedEvents, key)
		} else if err != nil {
			klog.Errorf("could not get %s %s event for node(%s), %v", component, key, m.nodeName, err)
			break
		}
	}

	for _, key := range deletedEvents {
		if err := m.store.Delete(key); err != nil {
			klog.Errorf("failed to gc events %s, %v", key, err)
		} else {
			klog.Infof("gc events %s successfully", key)
		}
	}
}
