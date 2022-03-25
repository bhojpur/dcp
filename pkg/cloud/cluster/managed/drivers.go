package managed

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
	"net/http"

	"github.com/bhojpur/dcp/pkg/cloud/clientaccess"
	"github.com/bhojpur/dcp/pkg/cloud/daemons/config"
)

var (
	defaultDriver string
	drivers       []Driver
)

type Driver interface {
	IsInitialized(ctx context.Context, config *config.Control) (bool, error)
	Register(ctx context.Context, config *config.Control, handler http.Handler) (http.Handler, error)
	Reset(ctx context.Context, reboostrap func() error) error
	Start(ctx context.Context, clientAccessInfo *clientaccess.Info) error
	Test(ctx context.Context) error
	Restore(ctx context.Context) error
	EndpointName() string
	Snapshot(ctx context.Context, config *config.Control) error
	ReconcileSnapshotData(ctx context.Context) error
	GetMembersClientURLs(ctx context.Context) ([]string, error)
	RemoveSelf(ctx context.Context) error
}

func RegisterDriver(d Driver) {
	drivers = append(drivers, d)
}

func Registered() []Driver {
	return drivers
}

func Default() string {
	if defaultDriver == "" && len(drivers) == 1 {
		return drivers[0].EndpointName()
	}
	return defaultDriver
}
