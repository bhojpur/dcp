package initializer

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
	"k8s.io/client-go/informers"

	dcpinformers "github.com/bhojpur/dcp/pkg/appmanager/client/informers/externalversions"
	"github.com/bhojpur/dcp/pkg/engine/cachemanager"
	"github.com/bhojpur/dcp/pkg/engine/filter"
	"github.com/bhojpur/dcp/pkg/engine/kubernetes/serializer"
	"github.com/bhojpur/dcp/pkg/engine/util"
)

// WantsSharedInformerFactory is an interface for setting SharedInformerFactory
type WantsSharedInformerFactory interface {
	SetSharedInformerFactory(factory informers.SharedInformerFactory) error
}

// WantsDcpSharedInformerFactory is an interface for setting App-Manager SharedInformerFactory
type WantsDcpSharedInformerFactory interface {
	SetDcpSharedInformerFactory(dcpFactory dcpinformers.SharedInformerFactory) error
}

// WantsNodeName is an interface for setting node name
type WantsNodeName interface {
	SetNodeName(nodeName string) error
}

// WantsSerializerManager is an interface for setting serializer manager
type WantsSerializerManager interface {
	SetSerializerManager(s *serializer.SerializerManager) error
}

// WantsStorageWrapper is an interface for setting StorageWrapper
type WantsStorageWrapper interface {
	SetStorageWrapper(s cachemanager.StorageWrapper) error
}

// WantsMasterServiceAddr is an interface for setting mutated master service address
type WantsMasterServiceAddr interface {
	SetMasterServiceAddr(addr string) error
}

// WantsWorkingMode is an interface for setting working mode
type WantsWorkingMode interface {
	SetWorkingMode(mode util.WorkingMode) error
}

// genericFilterInitializer is responsible for initializing generic filter
type genericFilterInitializer struct {
	factory           informers.SharedInformerFactory
	dcpFactory        dcpinformers.SharedInformerFactory
	serializerManager *serializer.SerializerManager
	storageWrapper    cachemanager.StorageWrapper
	nodeName          string
	masterServiceAddr string
	workingMode       util.WorkingMode
}

// New creates an filterInitializer object
func New(factory informers.SharedInformerFactory,
	dcpFactory dcpinformers.SharedInformerFactory,
	sm *serializer.SerializerManager,
	sw cachemanager.StorageWrapper,
	nodeName string,
	masterServiceAddr string,
	workingMode util.WorkingMode) *genericFilterInitializer {
	return &genericFilterInitializer{
		factory:           factory,
		dcpFactory:        dcpFactory,
		serializerManager: sm,
		storageWrapper:    sw,
		nodeName:          nodeName,
		masterServiceAddr: masterServiceAddr,
		workingMode:       workingMode,
	}
}

// Initialize used for executing filter initialization
func (fi *genericFilterInitializer) Initialize(ins filter.Interface) error {
	if wants, ok := ins.(WantsWorkingMode); ok {
		if err := wants.SetWorkingMode(fi.workingMode); err != nil {
			return err
		}
	}

	if wants, ok := ins.(WantsNodeName); ok {
		if err := wants.SetNodeName(fi.nodeName); err != nil {
			return err
		}
	}

	if wants, ok := ins.(WantsMasterServiceAddr); ok {
		if err := wants.SetMasterServiceAddr(fi.masterServiceAddr); err != nil {
			return err
		}
	}

	if wants, ok := ins.(WantsSharedInformerFactory); ok {
		if err := wants.SetSharedInformerFactory(fi.factory); err != nil {
			return err
		}
	}

	if wants, ok := ins.(WantsDcpSharedInformerFactory); ok {
		if err := wants.SetDcpSharedInformerFactory(fi.dcpFactory); err != nil {
			return err
		}
	}

	if wants, ok := ins.(WantsSerializerManager); ok {
		if err := wants.SetSerializerManager(fi.serializerManager); err != nil {
			return err
		}
	}

	if wants, ok := ins.(WantsStorageWrapper); ok {
		if err := wants.SetStorageWrapper(fi.storageWrapper); err != nil {
			return err
		}
	}

	return nil
}
