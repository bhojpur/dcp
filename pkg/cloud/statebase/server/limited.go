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

type LimitedServer struct {
	backend Backend
	scheme  string
}

func (l *LimitedServer) Range(ctx context.Context, r *etcdserverpb.RangeRequest) (*RangeResponse, error) {
	if len(r.RangeEnd) == 0 {
		return l.get(ctx, r)
	}
	return l.list(ctx, r)
}

func txnHeader(rev int64) *etcdserverpb.ResponseHeader {
	return &etcdserverpb.ResponseHeader{
		Revision: rev,
	}
}

func (l *LimitedServer) Txn(ctx context.Context, txn *etcdserverpb.TxnRequest) (*etcdserverpb.TxnResponse, error) {
	if put := isCreate(txn); put != nil {
		return l.create(ctx, put, txn)
	}
	if rev, key, ok := isDelete(txn); ok {
		return l.delete(ctx, key, rev)
	}
	if rev, key, value, lease, ok := isUpdate(txn); ok {
		return l.update(ctx, rev, key, value, lease)
	}
	if isCompact(txn) {
		return l.compact(ctx)
	}
	return nil, fmt.Errorf("unsupported transaction: %v", txn)
}

type ResponseHeader struct {
	Revision int64
}

type RangeResponse struct {
	Header *etcdserverpb.ResponseHeader
	Kvs    []*KeyValue
	More   bool
	Count  int64
}
