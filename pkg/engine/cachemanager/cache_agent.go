package cachemanager

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
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/pkg/engine/util"
)

const (
	sepForAgent = ","
)

func (cm *cacheManager) initCacheAgents() error {
	if cm.sharedFactory == nil {
		return nil
	}
	configmapInformer := cm.sharedFactory.Core().V1().ConfigMaps().Informer()
	configmapInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    cm.addConfigmap,
		UpdateFunc: cm.updateConfigmap,
	})

	klog.Infof("init cache agents to %v", cm.cacheAgents)
	return nil
}

func (cm *cacheManager) addConfigmap(obj interface{}) {
	cfg, ok := obj.(*corev1.ConfigMap)
	if !ok {
		return
	}

	deletedAgents := cm.updateCacheAgents(cfg.Data[util.CacheUserAgentsKey], "add")
	cm.deleteAgentCache(deletedAgents)
}

func (cm *cacheManager) updateConfigmap(oldObj, newObj interface{}) {
	oldCfg, ok := oldObj.(*corev1.ConfigMap)
	if !ok {
		return
	}

	newCfg, ok := newObj.(*corev1.ConfigMap)
	if !ok {
		return
	}

	if oldCfg.Data[util.CacheUserAgentsKey] == newCfg.Data[util.CacheUserAgentsKey] {
		return
	}

	deletedAgents := cm.updateCacheAgents(newCfg.Data[util.CacheUserAgentsKey], "update")
	cm.deleteAgentCache(deletedAgents)
}

// updateCacheAgents update cache agents
func (cm *cacheManager) updateCacheAgents(cacheAgents, action string) sets.String {
	newAgents := sets.NewString()
	for _, agent := range strings.Split(cacheAgents, sepForAgent) {
		agent = strings.TrimSpace(agent)
		if len(agent) != 0 {
			newAgents.Insert(agent)
		}
	}

	cm.Lock()
	defer cm.Unlock()
	cm.cacheAgents = cm.cacheAgents.Delete(util.DefaultCacheAgents...)
	if cm.cacheAgents.Equal(newAgents) {
		// add default cache agents
		cm.cacheAgents = cm.cacheAgents.Insert(util.DefaultCacheAgents...)
		return sets.String{}
	}

	// get deleted and added agents
	deletedAgents := cm.cacheAgents.Difference(newAgents)
	addedAgents := newAgents.Difference(cm.cacheAgents)

	// construct new cache agents
	cm.cacheAgents = cm.cacheAgents.Delete(deletedAgents.List()...)
	cm.cacheAgents = cm.cacheAgents.Insert(addedAgents.List()...)
	cm.cacheAgents = cm.cacheAgents.Insert(util.DefaultCacheAgents...)
	klog.Infof("current cache agents: %v after %s, deleted agents: %v", cm.cacheAgents, action, deletedAgents)

	// return deleted agents
	return deletedAgents
}

func (cm *cacheManager) deleteAgentCache(deletedAgents sets.String) {
	// delete cache data for deleted agents
	if deletedAgents.Len() > 0 {
		keys := deletedAgents.List()
		for i := range keys {
			if err := cm.storage.DeleteCollection(keys[i]); err != nil {
				klog.Errorf("failed to cleanup cache for deleted agent(%s), %v", keys[i], err)
			} else {
				klog.Infof("cleanup cache for agent(%s) successfully", keys[i])
			}
		}
	}
}
