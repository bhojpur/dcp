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
	"fmt"
	"sync"
	"time"

	"k8s.io/api/admissionregistration/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	admissionregistrationinformers "k8s.io/client-go/informers/admissionregistration/v1beta1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	admissionregistrationlisters "k8s.io/client-go/listers/admissionregistration/v1beta1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"

	extclient "github.com/bhojpur/dcp/pkg/appmanager/client"
	webhookutil "github.com/bhojpur/dcp/pkg/appmanager/webhook/util"
	"github.com/bhojpur/dcp/pkg/appmanager/webhook/util/configuration"
	"github.com/bhojpur/dcp/pkg/appmanager/webhook/util/generator"
	"github.com/bhojpur/dcp/pkg/appmanager/webhook/util/writer"
)

var (
	mutatingWebhookConfigurationName   = "app-mutating-webhook-configuration"
	validatingWebhookConfigurationName = "app-validating-webhook-configuration"

	namespace  = webhookutil.GetNamespace()
	secretName = webhookutil.GetSecretName()

	uninit   = make(chan struct{})
	onceInit = sync.Once{}
)

func Inited() chan struct{} {
	return uninit
}

type Controller struct {
	kubeClient    clientset.Interface
	runtimeClient client.Client
	handlers      map[string]webhookutil.Handler

	informerFactory    informers.SharedInformerFactory
	secretLister       corelisters.SecretNamespaceLister
	mutatingWCLister   admissionregistrationlisters.MutatingWebhookConfigurationLister
	validatingWCLister admissionregistrationlisters.ValidatingWebhookConfigurationLister
	synced             []cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

func New(cli client.Client, handlers map[string]webhookutil.Handler) (*Controller, error) {
	c := &Controller{
		kubeClient:    extclient.GetGenericClient().KubeClient,
		runtimeClient: cli,
		handlers:      handlers,

		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "webhook-controller"),
	}

	c.informerFactory = informers.NewSharedInformerFactory(c.kubeClient, 0)

	secretInformer := coreinformers.New(c.informerFactory, namespace, nil).Secrets()
	c.secretLister = secretInformer.Lister().Secrets(namespace)

	admissionRegistrationInformer := admissionregistrationinformers.New(c.informerFactory, v1.NamespaceAll, nil)
	c.mutatingWCLister = admissionRegistrationInformer.MutatingWebhookConfigurations().Lister()
	c.validatingWCLister = admissionRegistrationInformer.ValidatingWebhookConfigurations().Lister()

	c.synced = []cache.InformerSynced{
		secretInformer.Informer().HasSynced,
		admissionRegistrationInformer.MutatingWebhookConfigurations().Informer().HasSynced,
		admissionRegistrationInformer.ValidatingWebhookConfigurations().Informer().HasSynced,
	}

	secretInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			secret := obj.(*v1.Secret)
			if secret.Name == secretName {
				klog.Infof("Secret %s/%s added", secret.GetNamespace(), secretName)
				c.queue.Add("")
			}
		},
		UpdateFunc: func(old, cur interface{}) {
			secret := cur.(*v1.Secret)
			if secret.Name == secretName {
				klog.Infof("Secret %s/%s updated", secret.GetNamespace(), secretName)
				c.queue.Add("")
			}
		},
	})

	admissionRegistrationInformer.MutatingWebhookConfigurations().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			conf := obj.(*v1beta1.MutatingWebhookConfiguration)
			if conf.Name == mutatingWebhookConfigurationName {
				klog.Infof("MutatingWebhookConfiguration %s added", mutatingWebhookConfigurationName)
				c.queue.Add("")
			}
		},
		UpdateFunc: func(old, cur interface{}) {
			conf := cur.(*v1beta1.MutatingWebhookConfiguration)
			if conf.Name == mutatingWebhookConfigurationName {
				klog.Infof("MutatingWebhookConfiguration %s update", mutatingWebhookConfigurationName)
				c.queue.Add("")
			}
		},
	})

	admissionRegistrationInformer.ValidatingWebhookConfigurations().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			conf := obj.(*v1beta1.ValidatingWebhookConfiguration)
			if conf.Name == validatingWebhookConfigurationName {
				klog.Infof("ValidatingWebhookConfiguration %s added", validatingWebhookConfigurationName)
				c.queue.Add("")
			}
		},
		UpdateFunc: func(old, cur interface{}) {
			conf := cur.(*v1beta1.ValidatingWebhookConfiguration)
			if conf.Name == validatingWebhookConfigurationName {
				klog.Infof("ValidatingWebhookConfiguration %s updated", validatingWebhookConfigurationName)
				c.queue.Add("")
			}
		},
	})

	return c, nil
}

func (c *Controller) Start(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	klog.Infof("Starting webhook-controller")
	defer klog.Infof("Shutting down webhook-controller")

	c.informerFactory.Start(stopCh)
	if !cache.WaitForNamedCacheSync("webhook-controller", stopCh, c.synced...) {
		return
	}

	go wait.Until(func() {
		for c.processNextWorkItem() {
		}
	}, time.Second, stopCh)
	klog.Infof("Started Bhojpur DCP webhook-controller")

	<-stopCh
}

func (c *Controller) processNextWorkItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.sync()
	if err == nil {
		c.queue.Forget(key)
		return true
	}

	utilruntime.HandleError(fmt.Errorf("sync %q failed with %v", key, err))
	c.queue.AddRateLimited(key)

	return true
}

func (c *Controller) sync() error {
	klog.Infof("Starting to sync webhook certs and configurations")
	defer func() {
		klog.Infof("Finished to sync webhook certs and configurations")
	}()

	var dnsName string
	var certWriter writer.CertWriter
	var err error

	if dnsName = webhookutil.GetHost(); len(dnsName) > 0 {
		certWriter, err = writer.NewFSCertWriter(writer.FSCertWriterOptions{
			Path: webhookutil.GetCertDir(),
		})
		klog.Infof("Use Fs Cert Writer")
	} else {
		dnsName = generator.ServiceToCommonName(webhookutil.GetNamespace(), webhookutil.GetServiceName())
		certWriter, err = writer.NewSecretCertWriter(writer.SecretCertWriterOptions{
			Client: c.runtimeClient,
			Secret: &types.NamespacedName{Namespace: webhookutil.GetNamespace(), Name: webhookutil.GetSecretName()},
		})
		klog.Infof("Use Secret Cert Writer")
	}
	if err != nil {
		return fmt.Errorf("failed to create certs writer: %v", err)
	}

	certs, _, err := certWriter.EnsureCert(dnsName)
	if err != nil {
		return fmt.Errorf("failed to ensure certs: %v", err)
	}
	if err := writer.WriteCertsToDir(webhookutil.GetCertDir(), certs); err != nil {
		return fmt.Errorf("failed to write certs to dir: %v", err)
	}

	if err := configuration.Ensure(c.runtimeClient, c.handlers, certs.CACert); err != nil {
		return fmt.Errorf("failed to ensure configuration: %v", err)
	}

	onceInit.Do(func() {
		close(uninit)
	})
	return nil
}
