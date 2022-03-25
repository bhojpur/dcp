package fieldindex

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
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	IndexNameForPodNodeName = "spec.nodeName"
	IndexNameForOwnerRefUID = "ownerRefUID"
)

var registerOnce sync.Once

func RegisterFieldIndexes(c cache.Cache) error {
	var err error
	registerOnce.Do(func() {
		// pod nodeName
		err = c.IndexField(context.TODO(), &v1.Pod{}, IndexNameForPodNodeName, func(obj client.Object) []string {
			pod, ok := obj.(*v1.Pod)
			if !ok {
				return []string{}
			}
			if len(pod.Spec.NodeName) == 0 {
				return []string{}
			}
			return []string{pod.Spec.NodeName}
		})
		if err != nil {
			return
		}

		ownerIndexFunc := func(obj client.Object) []string {
			metaObj, ok := obj.(metav1.Object)
			if !ok {
				return []string{}
			}
			var owners []string
			for _, ref := range metaObj.GetOwnerReferences() {
				owners = append(owners, string(ref.UID))
			}
			return owners
		}

		// pod ownerReference
		if err = c.IndexField(context.TODO(), &v1.Pod{}, IndexNameForOwnerRefUID, ownerIndexFunc); err != nil {
			return
		}
		// pvc ownerReference
		if err = c.IndexField(context.TODO(), &v1.PersistentVolumeClaim{}, IndexNameForOwnerRefUID, ownerIndexFunc); err != nil {
			return
		}
	})
	return err
}
