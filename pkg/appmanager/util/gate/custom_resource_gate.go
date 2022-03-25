package gate

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
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/bhojpur/dcp/pkg/appmanager/apis"
)

const (
	envCustomResourceEnable = "CUSTOM_RESOURCE_ENABLE"
)

var (
	internalScheme  = runtime.NewScheme()
	discoveryClient discovery.DiscoveryInterface

	isNotNotFound = func(err error) bool { return !errors.IsNotFound(err) }
)

func init() {
	_ = apis.AddToScheme(internalScheme)
	cfg, err := config.GetConfig()
	if err == nil {
		discoveryClient = discovery.NewDiscoveryClientForConfigOrDie(cfg)
	}
}

// ResourceEnabled help runnable check if the custom resource is valid and enabled
// 1. If this CRD is not found from kueb-apiserver, it is invalid.
// 2. If 'CUSTOM_RESOURCE_ENABLE' env is not empty and this CRD kind is not in ${CUSTOM_RESOURCE_ENABLE}.
func ResourceEnabled(obj runtime.Object) bool {
	gvk, err := apiutil.GVKForObject(obj, internalScheme)
	if err != nil {
		klog.Warningf("custom resource gate not recognized object %T in scheme: %v", obj, err)
		return false
	}

	return discoveryEnabled(gvk) && envEnabled(gvk)
}

func discoveryEnabled(gvk schema.GroupVersionKind) bool {
	if discoveryClient == nil {
		return true
	}
	var resourceList *metav1.APIResourceList
	err := retry.OnError(retry.DefaultBackoff, isNotNotFound, func() error {
		var err error
		resourceList, err = discoveryClient.ServerResourcesForGroupVersion(gvk.GroupVersion().String())
		if err != nil && !errors.IsNotFound(err) {
			klog.Infof("custom resource gate failed to get groupVersionKind %v in discovery: %v", gvk, err)
		}
		return err
	})
	if err != nil {
		if errors.IsNotFound(err) {
			klog.Infof("custom resource gate not found groupVersionKind %v in discovery: %v", gvk, err)
			return false
		}
		// This might be caused by abnormal apiserver or etcd, ignore the discovery and just use envEnable
		return true
	}

	for _, r := range resourceList.APIResources {
		if r.Kind == gvk.Kind {
			return true
		}
	}

	return false
}

func envEnabled(gvk schema.GroupVersionKind) bool {
	limits := strings.TrimSpace(os.Getenv(envCustomResourceEnable))
	if len(limits) == 0 {
		// all enabled by default
		return true
	}

	if !sets.NewString(strings.Split(limits, ",")...).Has(gvk.Kind) {
		klog.Warningf("custom resource gate not found groupVersionKind %v in CUSTOM_RESOURCE_ENABLE: %v", gvk, limits)
		return false
	}

	return true
}
