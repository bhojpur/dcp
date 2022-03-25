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
	"bytes"
	"sync"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/pkg/engine/storage"
	"github.com/bhojpur/dcp/pkg/engine/util"
)

// StorageWrapper is wrapper for storage.Store interface
// in order to handle serialize runtime object
type StorageWrapper interface {
	Create(key string, obj runtime.Object) error
	Delete(key string) error
	Get(key string) (runtime.Object, error)
	ListKeys(key string) ([]string, error)
	List(key string) ([]runtime.Object, error)
	Update(key string, obj runtime.Object) error
	Replace(rootKey string, objs map[string]runtime.Object) error
	DeleteCollection(rootKey string) error
	GetRaw(key string) ([]byte, error)
	UpdateRaw(key string, contents []byte) error
}

type storageWrapper struct {
	sync.RWMutex
	store             storage.Store
	backendSerializer runtime.Serializer
	cache             map[string]runtime.Object
}

// NewStorageWrapper create a StorageWrapper object
func NewStorageWrapper(storage storage.Store) StorageWrapper {
	return &storageWrapper{
		store:             storage,
		backendSerializer: json.NewSerializerWithOptions(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme, json.SerializerOptions{}),
		cache:             make(map[string]runtime.Object),
	}
}

// Create store runtime object into backend storage
// if obj is nil, the storage used to represent the key
// will be created. for example: for disk storage,
// a directory that indicates the key will be created.
func (sw *storageWrapper) Create(key string, obj runtime.Object) error {
	var buf bytes.Buffer
	if obj != nil {
		if err := sw.backendSerializer.Encode(obj, &buf); err != nil {
			klog.Errorf("failed to encode object in create for %s, %v", key, err)
			return err
		}
	}

	if err := sw.store.Create(key, buf.Bytes()); err != nil {
		return err
	}

	if obj != nil && isCacheKey(key) {
		sw.Lock()
		sw.cache[key] = obj
		sw.Unlock()
	}

	return nil
}

// Delete remove runtime object that by specified key from backend storage
func (sw *storageWrapper) Delete(key string) error {
	if err := sw.store.Delete(key); err != nil {
		return err
	}

	if isCacheKey(key) {
		sw.Lock()
		delete(sw.cache, key)
		sw.Unlock()
	}

	return nil
}

// Get get the runtime object that specified by key from backend storage
func (sw *storageWrapper) Get(key string) (runtime.Object, error) {
	cachedKey := isCacheKey(key)
	if cachedKey {
		sw.RLock()
		cachedObject, ok := sw.cache[key]
		sw.RUnlock()
		if ok && cachedObject != nil {
			return cachedObject, nil
		}
	}

	b, err := sw.GetRaw(key)
	if err != nil {
		return nil, err
	} else if len(b) == 0 {
		return nil, nil
	}
	//get the gvk from json data
	gvk, err := json.DefaultMetaFactory.Interpret(b)
	if err != nil {
		return nil, err
	}
	var UnstructuredObj runtime.Object
	if scheme.Scheme.Recognizes(*gvk) {
		UnstructuredObj = nil
	} else {
		UnstructuredObj = new(unstructured.Unstructured)
	}
	obj, gvk, err := sw.backendSerializer.Decode(b, nil, UnstructuredObj)
	if err != nil {
		klog.Errorf("could not decode %v for %s, %v", gvk, key, err)
		return nil, err
	}

	if cachedKey {
		sw.Lock()
		sw.cache[key] = obj
		sw.Unlock()
	}
	return obj, nil
}

// ListKeys list all keys with key as prefix
func (sw *storageWrapper) ListKeys(key string) ([]string, error) {
	return sw.store.ListKeys(key)
}

// List get all of runtime objects that specified by key as prefix
func (sw *storageWrapper) List(key string) ([]runtime.Object, error) {
	bb, err := sw.store.List(key)
	objects := make([]runtime.Object, 0, len(bb))
	if err != nil {
		klog.Errorf("could not list objects for %s, %v", key, err)
		return nil, err
	} else if len(bb) == 0 {
		if isPodKey(key) {
			// because at least there will be Bhojpur DCP server engine pod on the node.
			// if no pods in cache, maybe all of pods have been deleted by accident,
			// if empty object is returned, pods on node will be deleted by kubelet.
			// in order to prevent the influence to business, return error here so pods
			// will be kept on node.
			return objects, storage.ErrStorageNotFound
		}
		return objects, nil
	}
	//get the gvk from json data
	gvk, err := json.DefaultMetaFactory.Interpret(bb[0])
	if err != nil {
		return nil, err
	}
	var UnstructuredObj runtime.Object
	var recognized bool
	if scheme.Scheme.Recognizes(*gvk) {
		recognized = true
	}

	for i := range bb {
		if !recognized {
			UnstructuredObj = new(unstructured.Unstructured)
		}

		obj, gvk, err := sw.backendSerializer.Decode(bb[i], nil, UnstructuredObj)
		if err != nil {
			klog.Errorf("could not decode %v for %s, %v", gvk, key, err)
			continue
		}
		objects = append(objects, obj)
	}

	return objects, nil
}

// Update update runtime object in backend storage
func (sw *storageWrapper) Update(key string, obj runtime.Object) error {
	var buf bytes.Buffer
	if err := sw.backendSerializer.Encode(obj, &buf); err != nil {
		klog.Errorf("failed to encode object in update for %s, %v", key, err)
		return err
	}

	if err := sw.UpdateRaw(key, buf.Bytes()); err != nil {
		return err
	}

	if isCacheKey(key) {
		sw.Lock()
		sw.cache[key] = obj
		sw.Unlock()
	}

	return nil
}

// Replace will delete the old objects, and use the given objs instead.
func (sw *storageWrapper) Replace(rootKey string, objs map[string]runtime.Object) error {
	var buf bytes.Buffer
	contents := make(map[string][]byte, len(objs))
	for key, obj := range objs {
		if err := sw.backendSerializer.Encode(obj, &buf); err != nil {
			klog.Errorf("failed to encode object in update for %s, %v", key, err)
			return err
		}
		contents[key] = make([]byte, len(buf.Bytes()))
		copy(contents[key], buf.Bytes())
		buf.Reset()
	}

	return sw.store.Replace(rootKey, contents)
}

// DeleteCollection will delete all objects under rootKey
func (sw *storageWrapper) DeleteCollection(rootKey string) error {
	return sw.store.DeleteCollection(rootKey)
}

// GetRaw get byte data for specified key
func (sw *storageWrapper) GetRaw(key string) ([]byte, error) {
	return sw.store.Get(key)
}

// UpdateRaw update contents(byte date) for specified key
func (sw *storageWrapper) UpdateRaw(key string, contents []byte) error {
	return sw.store.Update(key, contents)
}

// isCacheKey verify runtime object is cached for specified key.
// in order to accelerate kubelet get node and lease object, we cache them
func isCacheKey(key string) bool {
	comp, resource, _, _ := util.SplitKey(key)
	switch {
	case comp == "kubelet" && resource == "nodes":
		return true
	case comp == "kubelet" && resource == "leases":
		return true
	}

	return false
}

// isPodKey verify the key is kubelet/pods or not
func isPodKey(key string) bool {
	comp, resource, _, _ := util.SplitKey(key)
	return comp == "kubelet" && resource == "pods"
}
