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

	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/api/v3/etcdserverpb"
	"go.etcd.io/etcd/api/v3/mvccpb"
)

// explicit interface check
var _ etcdserverpb.KVServer = (*KVServerBridge)(nil)

func (k *KVServerBridge) Range(ctx context.Context, r *etcdserverpb.RangeRequest) (*etcdserverpb.RangeResponse, error) {
	if r.KeysOnly {
		return nil, unsupported("keysOnly")
	}

	if r.MaxCreateRevision != 0 {
		return nil, unsupported("maxCreateRevision")
	}

	if r.SortOrder != 0 {
		return nil, unsupported("sortOrder")
	}

	if r.SortTarget != 0 {
		return nil, unsupported("sortTarget")
	}

	if r.Serializable {
		return nil, unsupported("serializable")
	}

	if r.KeysOnly {
		return nil, unsupported("keysOnly")
	}

	if r.MinModRevision != 0 {
		return nil, unsupported("minModRevision")
	}

	if r.MinCreateRevision != 0 {
		return nil, unsupported("minCreateRevision")
	}

	if r.MaxCreateRevision != 0 {
		return nil, unsupported("maxCreateRevision")
	}

	if r.MaxModRevision != 0 {
		return nil, unsupported("maxModRevision")
	}

	resp, err := k.limited.Range(ctx, r)
	if err != nil {
		logrus.Errorf("error while range on %s %s: %v", r.Key, r.RangeEnd, err)
		return nil, err
	}

	rangeResponse := &etcdserverpb.RangeResponse{
		More:   resp.More,
		Count:  resp.Count,
		Header: resp.Header,
		Kvs:    toKVs(resp.Kvs...),
	}

	return rangeResponse, nil
}

func toKVs(kvs ...*KeyValue) []*mvccpb.KeyValue {
	if len(kvs) == 0 || kvs[0] == nil {
		return nil
	}

	ret := make([]*mvccpb.KeyValue, 0, len(kvs))
	for _, kv := range kvs {
		newKV := toKV(kv)
		if newKV != nil {
			ret = append(ret, newKV)
		}
	}
	return ret
}

func toKV(kv *KeyValue) *mvccpb.KeyValue {
	if kv == nil {
		return nil
	}
	return &mvccpb.KeyValue{
		Key:            []byte(kv.Key),
		Value:          kv.Value,
		Lease:          kv.Lease,
		CreateRevision: kv.CreateRevision,
		ModRevision:    kv.ModRevision,
	}
}

func (k *KVServerBridge) Put(ctx context.Context, r *etcdserverpb.PutRequest) (*etcdserverpb.PutResponse, error) {
	return nil, fmt.Errorf("put is not supported")
}

func (k *KVServerBridge) DeleteRange(ctx context.Context, r *etcdserverpb.DeleteRangeRequest) (*etcdserverpb.DeleteRangeResponse, error) {
	return nil, fmt.Errorf("delete is not supported")
}

func (k *KVServerBridge) Txn(ctx context.Context, r *etcdserverpb.TxnRequest) (*etcdserverpb.TxnResponse, error) {
	res, err := k.limited.Txn(ctx, r)
	if err != nil {
		logrus.Errorf("error in txn: %v", err)
	}
	return res, err
}

func (k *KVServerBridge) Compact(ctx context.Context, r *etcdserverpb.CompactionRequest) (*etcdserverpb.CompactionResponse, error) {
	return &etcdserverpb.CompactionResponse{
		Header: &etcdserverpb.ResponseHeader{
			Revision: r.Revision,
		},
	}, nil
}

func unsupported(field string) error {
	return fmt.Errorf("%s is unsupported", field)
}
