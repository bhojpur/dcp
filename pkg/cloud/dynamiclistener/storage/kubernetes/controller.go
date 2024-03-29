package kubernetes

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
	"time"

	"github.com/bhojpur/dcp/pkg/cloud/dynamiclistener"
	"github.com/bhojpur/host/pkg/common/start"
	"github.com/bhojpur/host/pkg/generated/controllers/core"
	v1controller "github.com/bhojpur/host/pkg/generated/controllers/core/v1"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CoreGetter func() *core.Factory

func Load(ctx context.Context, secrets v1controller.SecretController, namespace, name string, backing dynamiclistener.TLSStorage) dynamiclistener.TLSStorage {
	storage := &storage{
		name:      name,
		namespace: namespace,
		storage:   backing,
		ctx:       ctx,
	}
	storage.init(secrets)
	return storage
}

func New(ctx context.Context, core CoreGetter, namespace, name string, backing dynamiclistener.TLSStorage) dynamiclistener.TLSStorage {
	storage := &storage{
		name:      name,
		namespace: namespace,
		storage:   backing,
		ctx:       ctx,
	}

	// lazy init
	go func() {
		for {
			core := core()
			if core != nil {
				storage.init(core.Core().V1().Secret())
				_ = start.All(ctx, 5, core)
				return
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
			}
		}
	}()

	return storage
}

type storage struct {
	sync.Mutex

	namespace, name string
	storage         dynamiclistener.TLSStorage
	secrets         v1controller.SecretClient
	ctx             context.Context
	tls             dynamiclistener.TLSFactory
}

func (s *storage) SetFactory(tls dynamiclistener.TLSFactory) {
	s.tls = tls
}

func (s *storage) init(secrets v1controller.SecretController) {
	s.Lock()
	defer s.Unlock()

	secrets.OnChange(s.ctx, "tls-storage", func(key string, secret *v1.Secret) (*v1.Secret, error) {
		if secret == nil {
			return nil, nil
		}
		if secret.Namespace == s.namespace && secret.Name == s.name {
			if err := s.Update(secret); err != nil {
				return nil, err
			}
		}

		return secret, nil
	})
	s.secrets = secrets

	secret, err := s.storage.Get()
	if err == nil && secret != nil && len(secret.Data) > 0 {
		// local storage had a cached secret, ensure that it exists in Kubernetes
		_, err := s.secrets.Create(&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:        s.name,
				Namespace:   s.namespace,
				Annotations: secret.Annotations,
			},
			Type: v1.SecretTypeTLS,
			Data: secret.Data,
		})
		if err != nil && !errors.IsAlreadyExists(err) {
			logrus.Warnf("Failed to create Kubernetes secret: %v", err)
		}
	} else {
		// local storage was empty, try to populate it
		secret, err := s.secrets.Get(s.namespace, s.name, metav1.GetOptions{})
		if err != nil {
			logrus.Warnf("Failed to init Kubernetes secret: %v", err)
			return
		}

		if err := s.storage.Update(secret); err != nil {
			logrus.Warnf("Failed to init backing storage secret: %v", err)
		}
	}
}

func (s *storage) Get() (*v1.Secret, error) {
	s.Lock()
	defer s.Unlock()

	return s.storage.Get()
}

func (s *storage) targetSecret() (*v1.Secret, error) {
	existingSecret, err := s.secrets.Get(s.namespace, s.name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      s.name,
				Namespace: s.namespace,
			},
		}, nil
	}
	return existingSecret, err
}

func (s *storage) saveInK8s(secret *v1.Secret) (*v1.Secret, error) {
	if s.secrets == nil {
		return secret, nil
	}

	targetSecret, err := s.targetSecret()
	if err != nil {
		return nil, err
	}

	// if we don't have a TLS factory we can't create certs, so don't bother trying to merge anything,
	// in favor of just blindly replacing the fields on the Kubernetes secret.
	if s.tls != nil {
		// merge new secret with secret from backing storage, if one exists
		if existing, err := s.storage.Get(); err == nil && existing != nil && len(existing.Data) > 0 {
			if newSecret, updated, err := s.tls.Merge(existing, secret); err == nil && updated {
				secret = newSecret
			}
		}

		// merge new secret with existing secret from Kubernetes, if one exists
		if len(targetSecret.Data) > 0 {
			if newSecret, updated, err := s.tls.Merge(targetSecret, secret); err != nil {
				return nil, err
			} else if !updated {
				return newSecret, nil
			} else {
				secret = newSecret
			}
		}
	}

	targetSecret.Annotations = secret.Annotations
	targetSecret.Type = v1.SecretTypeTLS
	targetSecret.Data = secret.Data

	if targetSecret.UID == "" {
		logrus.Infof("Creating new TLS secret for %v (count: %d): %v", targetSecret.Name, len(targetSecret.Annotations)-1, targetSecret.Annotations)
		return s.secrets.Create(targetSecret)
	}
	logrus.Infof("Updating TLS secret for %v (count: %d): %v", targetSecret.Name, len(targetSecret.Annotations)-1, targetSecret.Annotations)
	return s.secrets.Update(targetSecret)
}

func (s *storage) Update(secret *v1.Secret) (err error) {
	s.Lock()
	defer s.Unlock()

	for i := 0; i < 3; i++ {
		secret, err = s.saveInK8s(secret)
		if errors.IsConflict(err) {
			continue
		} else if err != nil {
			return err
		}
		break
	}

	if err != nil {
		return err
	}

	// update underlying storage
	return s.storage.Update(secret)
}
