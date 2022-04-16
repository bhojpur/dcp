package cluster

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
	"crypto/tls"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bhojpur/dcp/pkg/cloud/daemons/config"
	"github.com/bhojpur/dcp/pkg/cloud/dynamiclistener"
	"github.com/bhojpur/dcp/pkg/cloud/dynamiclistener/factory"
	"github.com/bhojpur/dcp/pkg/cloud/dynamiclistener/storage/file"
	"github.com/bhojpur/dcp/pkg/cloud/dynamiclistener/storage/kubernetes"
	"github.com/bhojpur/dcp/pkg/cloud/dynamiclistener/storage/memory"
	"github.com/bhojpur/dcp/pkg/cloud/etcd"
	"github.com/bhojpur/dcp/pkg/cloud/version"
	"github.com/bhojpur/host/pkg/generated/controllers/core"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// newListener returns a new TCP listener and HTTP request handler using dynamiclistener.
// dynamiclistener will use the cluster's Server CA to sign the dynamically generate certificate,
// and will sync the certs into the Kubernetes datastore, with a local disk cache.
func (c *Cluster) newListener(ctx context.Context) (net.Listener, http.Handler, error) {
	if c.managedDB != nil {
		if _, err := os.Stat(etcd.ResetFile(c.config)); err == nil {
			// delete the dynamic listener file if it exists after restoration to fix restoration
			// on fresh nodes
			os.Remove(filepath.Join(c.config.DataDir, "tls/dynamic-cert.json"))
		}
	}
	tcp, err := dynamiclistener.NewTCPListener(c.config.BindAddress, c.config.SupervisorPort)
	if err != nil {
		return nil, nil, err
	}
	cert, key, err := factory.LoadCerts(c.config.Runtime.ServerCA, c.config.Runtime.ServerCAKey)
	if err != nil {
		return nil, nil, err
	}
	storage := tlsStorage(ctx, c.config.DataDir, c.config.Runtime)
	return dynamiclistener.NewListener(tcp, storage, cert, key, dynamiclistener.Config{
		ExpirationDaysCheck: config.CertificateRenewDays,
		Organization:        []string{version.Program},
		SANs:                append(c.config.SANs, "kubernetes", "kubernetes.default", "kubernetes.default.svc", "kubernetes.default.svc."+c.config.ClusterDomain),
		CN:                  version.Program,
		TLSConfig: &tls.Config{
			ClientAuth:   tls.RequestClientCert,
			MinVersion:   c.config.TLSMinVersion,
			CipherSuites: c.config.TLSCipherSuites,
			NextProtos:   []string{"h2", "http/1.1"},
		},
		RegenerateCerts: func() bool {
			const regenerateDynamicListenerFile = "dynamic-cert-regenerate"
			dynamicListenerRegenFilePath := filepath.Join(c.config.DataDir, "tls", regenerateDynamicListenerFile)
			if _, err := os.Stat(dynamicListenerRegenFilePath); err == nil {
				os.Remove(dynamicListenerRegenFilePath)
				return true
			}
			return false
		},
	})
}

// initClusterAndHTTPS sets up the dynamic tls listener, request router,
// and cluster database. Once the database is up, it starts the supervisor http server.
func (c *Cluster) initClusterAndHTTPS(ctx context.Context) error {
	// Set up dynamiclistener TLS listener and request handler
	listener, handler, err := c.newListener(ctx)
	if err != nil {
		return err
	}

	// Get the base request handler
	handler, err = c.getHandler(handler)
	if err != nil {
		return err
	}

	// Config the cluster database and allow it to add additional request handlers
	handler, err = c.initClusterDB(ctx, handler)
	if err != nil {
		return err
	}

	// Create a HTTP server with the registered request handlers, using logrus for logging
	server := http.Server{
		Handler: handler,
	}

	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		server.ErrorLog = log.New(logrus.StandardLogger().Writer(), "Cluster-Http-Server ", log.LstdFlags)
	} else {
		server.ErrorLog = log.New(ioutil.Discard, "Cluster-Http-Server", 0)
	}

	// Start the supervisor http server on the tls listener
	go func() {
		err := server.Serve(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.Fatalf("server stopped: %v", err)
		}
	}()

	// Shutdown the http server when the context is closed
	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background())
	}()

	return nil
}

// tlsStorage creates an in-memory cache for dynamiclistener's certificate, backed by a file on disk
// and the Kubernetes datastore.
func tlsStorage(ctx context.Context, dataDir string, runtime *config.ControlRuntime) dynamiclistener.TLSStorage {
	fileStorage := file.New(filepath.Join(dataDir, "tls/dynamic-cert.json"))
	cache := memory.NewBacked(fileStorage)
	return kubernetes.New(ctx, func() *core.Factory {
		return runtime.Core
	}, metav1.NamespaceSystem, version.Program+"-serving", cache)
}
