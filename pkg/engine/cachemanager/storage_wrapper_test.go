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
	"fmt"
	"os"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/bhojpur/dcp/pkg/engine/storage"
	"github.com/bhojpur/dcp/pkg/engine/storage/disk"
)

func clearDir(dir string) error {
	return os.RemoveAll(dir)
}

var testPod = runtime.Object(&v1.Pod{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Pod",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:            "mypod1",
		Namespace:       "default",
		ResourceVersion: "1",
	},
})

func TestStorageWrapper(t *testing.T) {
	dir := fmt.Sprintf("%s-%d", rootDir, time.Now().Unix())

	defer clearDir(dir)

	dStorage, err := disk.NewDiskStorage(dir)
	if err != nil {
		t.Errorf("failed to create disk storage, %v", err)
	}
	sWrapper := NewStorageWrapper(dStorage)

	t.Run("Test create storage", func(t *testing.T) {
		err = sWrapper.Create("kubelet/pods/default/mypod1", testPod)
		if err != nil {
			t.Errorf("failed to create obj, %v", err)
		}
		obj, err := sWrapper.Get("kubelet/pods/default/mypod1")
		if err != nil {
			t.Errorf("failed to create obj, %v", err)
		}
		accessor := meta.NewAccessor()
		name, _ := accessor.Name(obj)
		if name != "mypod1" {
			t.Errorf("the name is not expected, expect mypod1, get %s", name)
		}
	})

	t.Run("Test update storage", func(t *testing.T) {
		updatePod := runtime.Object(&v1.Pod{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Pod",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:            "mypod1",
				Namespace:       "default",
				ResourceVersion: "1",
				Labels: map[string]string{
					"tag": "test",
				},
			},
		})
		err = sWrapper.Update("kubelet/pods/default/mypod1", updatePod)
		if err != nil {
			t.Errorf("failed to update obj, %v", err)
		}
		obj, err := sWrapper.Get("kubelet/pods/default/mypod1")
		if err != nil {
			t.Errorf("unexpected error, %v", err)
		}
		accessor := meta.NewAccessor()
		labels, _ := accessor.Labels(obj)
		if vaule, ok := labels["tag"]; ok {
			if vaule != "test" {
				t.Errorf("failed to get label, expect test, get %s", vaule)
			}
		} else {
			t.Errorf("unexpected error, the label `tag` is not existed")
		}

	})

	t.Run("Test list keys and obj", func(t *testing.T) {
		// test an exist key
		keys, err := sWrapper.ListKeys("kubelet/pods/default")
		if err != nil {
			t.Errorf("failed to list keys, %v", err)
		}
		if len(keys) != 1 {
			t.Errorf("the length of keys is not expected, expect 1, get %d", len(keys))
		}

		// test a not exist key
		_, err = sWrapper.ListKeys("kubelet/pods/test")
		if err != nil {
			t.Errorf("failed to list keys, %v", err)
		}

		// test list obj
		_, err = sWrapper.List("kubelet/pods/default")
		if err != nil {
			t.Errorf("failed to list obj, %v", err)
		}
	})

	t.Run("Test replace obj", func(t *testing.T) {
		err = sWrapper.Replace("kubelet/pods/default", map[string]runtime.Object{
			"kubelet/pods/default/mypod1": runtime.Object(&v1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:            "mypod1",
					Namespace:       "default",
					ResourceVersion: "1",
					Labels: map[string]string{
						"tag": "test",
					},
				},
			}),
		})
		if err != nil {
			t.Errorf("failed to replace objs, %v", err)
		}
	})

	t.Run("Test delete storage", func(t *testing.T) {
		err = sWrapper.Delete("kubelet/pods/default/mypod1")
		if err != nil {
			t.Errorf("failed to delete obj, %v", err)
		}
		_, err = sWrapper.Get("kubelet/pods/default/mypod1")
		if err != storage.ErrStorageNotFound {
			t.Errorf("unexpected error, %v", err)
		}
	})

	t.Run("Test list obj in empty path", func(t *testing.T) {
		_, err = sWrapper.List("kubelet/pods/default")
		if err != storage.ErrStorageNotFound {
			t.Errorf("failed to list obj, %v", err)
		}
	})
}
