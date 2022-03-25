package executor

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
	"errors"
	"io/ioutil"
	"path/filepath"

	daemonconfig "github.com/bhojpur/dcp/pkg/cloud/daemons/config"
	"github.com/bhojpur/dcp/pkg/cloud/version"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/server/v3/embed"
	"go.etcd.io/etcd/server/v3/etcdserver/api/rafthttp"
)

type Embedded struct {
	nodeConfig *daemonconfig.Node
}

func (e *Embedded) ETCD(ctx context.Context, args ETCDConfig, extraArgs []string) error {
	configFile, err := args.ToConfigFile(extraArgs)
	if err != nil {
		return err
	}
	cfg, err := embed.ConfigFromFile(configFile)
	if err != nil {
		return err
	}

	etcd, err := embed.StartEtcd(cfg)
	if err != nil {
		return err
	}

	go func() {
		select {
		case err := <-etcd.Server.ErrNotify():
			if errors.Is(err, rafthttp.ErrMemberRemoved) {
				tombstoneFile := filepath.Join(args.DataDir, "tombstone")
				if err := ioutil.WriteFile(tombstoneFile, []byte{}, 0600); err != nil {
					logrus.Fatalf("failed to write tombstone file to %s", tombstoneFile)
				}
				logrus.Infof("this node has been removed from the cluster please restart %s to rejoin the cluster", version.Program)
				return
			}
		case <-ctx.Done():
			logrus.Infof("stopping etcd")
			etcd.Close()
		case <-etcd.Server.StopNotify():
			logrus.Fatalf("etcd stopped")
		case err := <-etcd.Err():
			logrus.Fatalf("etcd exited: %v", err)
		}
	}()
	return nil
}
