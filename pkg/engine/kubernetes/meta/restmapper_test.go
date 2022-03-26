package meta

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
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bhojpur/dcp/pkg/engine/storage/disk"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var rootDir = "/tmp/restmapper"

func TestCreateRESTMapperManager(t *testing.T) {
	dStorage, err := disk.NewDiskStorage(rootDir)
	if err != nil {
		t.Errorf("failed to create disk storage, %v", err)
	}
	defer func() {
		if err := os.RemoveAll(rootDir); err != nil {
			t.Errorf("Unable to clean up test directory %q: %v", rootDir, err)
		}
	}()

	// initialize an empty DynamicRESTMapper
	engineRESTMapperManager := NewRESTMapperManager(dStorage)
	if engineRESTMapperManager.dynamicRESTMapper == nil || len(engineRESTMapperManager.dynamicRESTMapper) != 0 {
		t.Errorf("failed to initialize an empty dynamicRESTMapper, %v", err)
	}

	// reset engineRESTMapperManager
	if err := engineRESTMapperManager.ResetRESTMapper(); err != nil {
		t.Fatalf("failed to reset engineRESTMapperManager , %v", err)
	}

	// initialize an Non-empty DynamicRESTMapper
	// pre-cache the CRD information to the hard disk
	cachedDynamicRESTMapper := map[string]string{
		"samplecontroller.k8s.io/v1alpha1/foo":  "Foo",
		"samplecontroller.k8s.io/v1alpha1/foos": "Foo",
	}
	d, err := json.Marshal(cachedDynamicRESTMapper)
	if err != nil {
		t.Errorf("failed to serialize dynamicRESTMapper, %v", err)
	}
	err = dStorage.Update(CacheDynamicRESTMapperKey, d)
	if err != nil {
		t.Fatalf("failed to stored dynamicRESTMapper, %v", err)
	}

	// Determine whether the restmapper in the memory is the same as the information written to the disk
	engineRESTMapperManager = NewRESTMapperManager(dStorage)
	// get the CRD information in memory
	m := engineRESTMapperManager.dynamicRESTMapper
	gotMapper := dynamicRESTMapperToString(m)

	if !compareDynamicRESTMapper(gotMapper, cachedDynamicRESTMapper) {
		t.Errorf("Got mapper: %v, expect mapper: %v", gotMapper, cachedDynamicRESTMapper)
	}

	if err := engineRESTMapperManager.ResetRESTMapper(); err != nil {
		t.Fatalf("failed to reset engineRESTMapperManager , %v", err)
	}
}

func TestUpdateRESTMapper(t *testing.T) {
	dStorage, err := disk.NewDiskStorage(rootDir)
	if err != nil {
		t.Errorf("failed to create disk storage, %v", err)
	}
	defer func() {
		if err := os.RemoveAll(rootDir); err != nil {
			t.Errorf("Unable to clean up test directory %q: %v", rootDir, err)
		}
	}()
	engineRESTMapperManager := NewRESTMapperManager(dStorage)
	testcases := map[string]struct {
		cachedCRD        []schema.GroupVersionKind
		addCRD           schema.GroupVersionKind
		deleteCRD        schema.GroupVersionResource
		expectRESTMapper map[string]string
	}{
		"add the first CRD": {
			cachedCRD: []schema.GroupVersionKind{},
			addCRD:    schema.GroupVersionKind{Group: "samplecontroller.k8s.io", Version: "v1alpha1", Kind: "Foo"},
			expectRESTMapper: map[string]string{
				"samplecontroller.k8s.io/v1alpha1/foo":  "Foo",
				"samplecontroller.k8s.io/v1alpha1/foos": "Foo",
			},
		},

		"update with another CRD": {
			cachedCRD: []schema.GroupVersionKind{{Group: "samplecontroller.k8s.io", Version: "v1alpha1", Kind: "Foo"}},
			addCRD:    schema.GroupVersionKind{Group: "stable.example.com", Version: "v1", Kind: "CronTab"},
			expectRESTMapper: map[string]string{
				"samplecontroller.k8s.io/v1alpha1/foo":  "Foo",
				"samplecontroller.k8s.io/v1alpha1/foos": "Foo",
				"stable.example.com/v1/crontab":         "CronTab",
				"stable.example.com/v1/crontabs":        "CronTab",
			},
		},
		"delete one CRD": {
			cachedCRD: []schema.GroupVersionKind{
				{Group: "samplecontroller.k8s.io", Version: "v1alpha1", Kind: "Foo"},
				{Group: "stable.example.com", Version: "v1", Kind: "CronTab"},
			},
			deleteCRD: schema.GroupVersionResource{Group: "stable.example.com", Version: "v1", Resource: "crontabs"},
			expectRESTMapper: map[string]string{
				"samplecontroller.k8s.io/v1alpha1/foo":  "Foo",
				"samplecontroller.k8s.io/v1alpha1/foos": "Foo",
			},
		},
	}
	for k, tt := range testcases {
		t.Run(k, func(t *testing.T) {
			// initialize the cache CRD
			for _, gvk := range tt.cachedCRD {
				err := engineRESTMapperManager.UpdateKind(gvk)
				if err != nil {
					t.Errorf("failed to initialize the restmapper, %v", err)
				}
			}
			// add CRD information
			if !tt.addCRD.Empty() {
				err := engineRESTMapperManager.UpdateKind(tt.addCRD)
				if err != nil {
					t.Errorf("failed to add CRD information, %v", err)
				}
			} else {
				// delete CRD information
				err := engineRESTMapperManager.DeleteKindFor(tt.deleteCRD)
				if err != nil {
					t.Errorf("failed to delete CRD information, %v", err)
				}
			}

			// verify the CRD information in memory
			m := engineRESTMapperManager.dynamicRESTMapper
			memoryMapper := dynamicRESTMapperToString(m)

			if !compareDynamicRESTMapper(memoryMapper, tt.expectRESTMapper) {
				t.Errorf("Got mapper: %v, expect mapper: %v", memoryMapper, tt.expectRESTMapper)
			}

			// verify the CRD information in disk
			b, err := dStorage.Get(CacheDynamicRESTMapperKey)
			if err != nil {
				t.Fatalf("failed to get cached CRD information, %v", err)
			}
			cacheMapper := make(map[string]string)
			err = json.Unmarshal(b, &cacheMapper)
			if err != nil {
				t.Errorf("failed to decode the cached dynamicRESTMapper, %v", err)
			}

			if !compareDynamicRESTMapper(cacheMapper, tt.expectRESTMapper) {
				t.Errorf("cached mapper: %v, expect mapper: %v", cacheMapper, tt.expectRESTMapper)
			}
		})
	}
	if err := engineRESTMapperManager.ResetRESTMapper(); err != nil {
		t.Fatalf("failed to reset engineRESTMapperManager , %v", err)
	}
}

func TestResetRESTMapper(t *testing.T) {
	dStorage, err := disk.NewDiskStorage(rootDir)
	if err != nil {
		t.Errorf("failed to create disk storage, %v", err)
	}
	defer func() {
		if err := os.RemoveAll(rootDir); err != nil {
			t.Errorf("Unable to clean up test directory %q: %v", rootDir, err)
		}
	}()
	// initialize the RESTMapperManager
	engineRESTMapperManager := NewRESTMapperManager(dStorage)
	err = engineRESTMapperManager.UpdateKind(schema.GroupVersionKind{Group: "stable.example.com", Version: "v1", Kind: "CronTab"})
	if err != nil {
		t.Errorf("failed to initialize the restmapper, %v", err)
	}

	// reset the RESTMapperManager
	if err := engineRESTMapperManager.ResetRESTMapper(); err != nil {
		t.Errorf("failed to reset the restmapper, %v", err)
	}

	// Verify reset result
	if len(engineRESTMapperManager.dynamicRESTMapper) != 0 {
		t.Error("The cached GVR/GVK information in memory is not cleaned up.")
	} else if _, err := os.Stat(filepath.Join(rootDir, CacheDynamicRESTMapperKey)); !os.IsNotExist(err) {
		t.Error("The cached GVR/GVK information in disk is not deleted.")
	}
}

func compareDynamicRESTMapper(gotMapper map[string]string, expectedMapper map[string]string) bool {
	if len(gotMapper) != len(expectedMapper) {
		return false
	}

	for gvr, kind := range gotMapper {
		k, exists := expectedMapper[gvr]
		if !exists || k != kind {
			return false
		}
	}

	return true
}

func dynamicRESTMapperToString(m map[schema.GroupVersionResource]schema.GroupVersionKind) map[string]string {
	resultMapper := make(map[string]string, len(m))
	for currResource, currKind := range m {
		//key: Group/Version/Resource, value: Kind
		k := strings.Join([]string{currResource.Group, currResource.Version, currResource.Resource}, SepForGVR)
		resultMapper[k] = currKind.Kind
	}
	return resultMapper
}