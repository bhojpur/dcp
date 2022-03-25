package server

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
	"fmt"

	"go.etcd.io/etcd/api/v3/etcdserverpb"
)

// explicit interface check
var _ etcdserverpb.MaintenanceServer = (*KVServerBridge)(nil)

func (s *KVServerBridge) Alarm(context.Context, *etcdserverpb.AlarmRequest) (*etcdserverpb.AlarmResponse, error) {
	return nil, fmt.Errorf("alarm is not supported")
}

func (s *KVServerBridge) Status(ctx context.Context, r *etcdserverpb.StatusRequest) (*etcdserverpb.StatusResponse, error) {
	size, err := s.limited.dbSize(ctx)
	if err != nil {
		return nil, err
	}
	return &etcdserverpb.StatusResponse{
		Header: &etcdserverpb.ResponseHeader{},
		DbSize: size,
	}, nil
}

func (s *KVServerBridge) Defragment(context.Context, *etcdserverpb.DefragmentRequest) (*etcdserverpb.DefragmentResponse, error) {
	return nil, fmt.Errorf("defragment is not supported")
}

func (s *KVServerBridge) Hash(context.Context, *etcdserverpb.HashRequest) (*etcdserverpb.HashResponse, error) {
	return nil, fmt.Errorf("hash is not supported")
}

func (s *KVServerBridge) HashKV(context.Context, *etcdserverpb.HashKVRequest) (*etcdserverpb.HashKVResponse, error) {
	return nil, fmt.Errorf("hash kv is not supported")
}

func (s *KVServerBridge) Snapshot(*etcdserverpb.SnapshotRequest, etcdserverpb.Maintenance_SnapshotServer) error {
	return fmt.Errorf("snapshot is not supported")
}

func (s *KVServerBridge) MoveLeader(context.Context, *etcdserverpb.MoveLeaderRequest) (*etcdserverpb.MoveLeaderResponse, error) {
	return nil, fmt.Errorf("move leader is not supported")
}

func (s *KVServerBridge) Downgrade(context.Context, *etcdserverpb.DowngradeRequest) (*etcdserverpb.DowngradeResponse, error) {
	return nil, fmt.Errorf("downgrade is not supported")
}
