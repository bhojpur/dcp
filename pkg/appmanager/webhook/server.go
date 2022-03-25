package webhook

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
	"fmt"
	"time"

	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/capabilities"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	webhookutil "github.com/bhojpur/dcp/pkg/appmanager/webhook/util"
	webhookcontroller "github.com/bhojpur/dcp/pkg/appmanager/webhook/util/controller"
	"github.com/bhojpur/dcp/pkg/appmanager/webhook/util/health"
)

var (
	// HandlerMap contains all admission webhook handlers.
	HandlerMap = map[string]webhookutil.Handler{}

	Checker = health.Checker
)

func init() {
	capabilities.Initialize(capabilities.Capabilities{
		AllowPrivileged: true,
		PrivilegedSources: capabilities.PrivilegedSources{
			HostNetworkSources: []string{},
			HostPIDSources:     []string{},
			HostIPCSources:     []string{},
		},
	})
}

func addHandlers(m map[string]webhookutil.Handler) {
	for path, handler := range m {
		if len(path) == 0 {
			klog.Warningf("Skip handler with empty path.")
			continue
		}
		if path[0] != '/' {
			path = "/" + path
		}
		_, found := HandlerMap[path]
		if found {
			klog.V(1).Infof("conflicting webhook builder path %v in handler map", path)
		}
		HandlerMap[path] = handler
	}
}

func SetupWithManager(mgr manager.Manager) error {
	server := mgr.GetWebhookServer()
	server.Host = "0.0.0.0"
	server.Port = webhookutil.GetPort()
	server.CertDir = webhookutil.GetCertDir()

	// register admission handlers
	for path, handler := range HandlerMap {
		handler.SetOptions(webhookutil.Options{
			Client: mgr.GetClient(),
		})
		server.Register(path, &webhook.Admission{Handler: handler})
		klog.V(3).Infof("Registered webhook handler %s", path)
	}

	// register health handler
	server.Register("/healthz", &health.Handler{})

	return nil
}

// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete

func Initialize(mgr manager.Manager, stopCh <-chan struct{}) error {
	cli, err := client.NewDelegatingClient(client.NewDelegatingClientInput{
		CacheReader: mgr.GetAPIReader(),
		Client:      mgr.GetClient(),
	})
	if err != nil {
		return err
	}

	c, err := webhookcontroller.New(cli, HandlerMap)
	if err != nil {
		return err
	}
	go func() {
		c.Start(stopCh)
	}()

	timer := time.NewTimer(time.Second * 5)
	defer timer.Stop()
	select {
	case <-webhookcontroller.Inited():
		return nil
	case <-timer.C:
		return fmt.Errorf("failed to start webhook controller for waiting more than 5s")
	}
}
