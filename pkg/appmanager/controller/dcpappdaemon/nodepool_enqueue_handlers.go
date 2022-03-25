package dcpappdaemon

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

	"github.com/bhojpur/dcp/pkg/appmanager/apis/apps/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type EnqueueAppDaemonForNodePool struct {
	client client.Client
}

func (e *EnqueueAppDaemonForNodePool) Create(event event.CreateEvent, limitingInterface workqueue.RateLimitingInterface) {
	e.addAllAppDaemonToWorkQueue(limitingInterface)
}

func (e *EnqueueAppDaemonForNodePool) Update(event event.UpdateEvent, limitingInterface workqueue.RateLimitingInterface) {
	e.addAllAppDaemonToWorkQueue(limitingInterface)
}

func (e *EnqueueAppDaemonForNodePool) Delete(event event.DeleteEvent, limitingInterface workqueue.RateLimitingInterface) {
	e.addAllAppDaemonToWorkQueue(limitingInterface)
}

func (e *EnqueueAppDaemonForNodePool) Generic(event event.GenericEvent, limitingInterface workqueue.RateLimitingInterface) {
	return
}

func (e *EnqueueAppDaemonForNodePool) addAllAppDaemonToWorkQueue(limitingInterface workqueue.RateLimitingInterface) {
	ydas := &v1alpha1.DcpAppDaemonList{}
	if err := e.client.List(context.TODO(), ydas); err != nil {
		return
	}

	for _, ud := range ydas.Items {
		addAppDaemonToWorkQueue(ud.GetNamespace(), ud.GetName(), limitingInterface)
	}
}

var _ handler.EventHandler = &EnqueueAppDaemonForNodePool{}

// addYAppDaemonToWorkQueue adds the DcpAppDaemon the reconciler's workqueue
func addAppDaemonToWorkQueue(namespace, name string,
	q workqueue.RateLimitingInterface) {
	q.Add(reconcile.Request{
		NamespacedName: types.NamespacedName{Name: name, Namespace: namespace},
	})
}
