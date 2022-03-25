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

	"go.etcd.io/etcd/api/v3/etcdserverpb"
)

func isUpdate(txn *etcdserverpb.TxnRequest) (int64, string, []byte, int64, bool) {
	if len(txn.Compare) == 1 &&
		txn.Compare[0].Target == etcdserverpb.Compare_MOD &&
		txn.Compare[0].Result == etcdserverpb.Compare_EQUAL &&
		len(txn.Success) == 1 &&
		txn.Success[0].GetRequestPut() != nil &&
		len(txn.Failure) == 1 &&
		txn.Failure[0].GetRequestRange() != nil {
		return txn.Compare[0].GetModRevision(),
			string(txn.Compare[0].Key),
			txn.Success[0].GetRequestPut().Value,
			txn.Success[0].GetRequestPut().Lease,
			true
	}
	return 0, "", nil, 0, false
}

func (l *LimitedServer) update(ctx context.Context, rev int64, key string, value []byte, lease int64) (*etcdserverpb.TxnResponse, error) {
	var (
		kv  *KeyValue
		ok  bool
		err error
	)

	if rev == 0 {
		rev, err = l.backend.Create(ctx, key, value, lease)
		ok = true
	} else {
		rev, kv, ok, err = l.backend.Update(ctx, key, value, rev, lease)
	}
	if err != nil {
		return nil, err
	}

	resp := &etcdserverpb.TxnResponse{
		Header:    txnHeader(rev),
		Succeeded: ok,
	}

	if ok {
		resp.Responses = []*etcdserverpb.ResponseOp{
			{
				Response: &etcdserverpb.ResponseOp_ResponsePut{
					ResponsePut: &etcdserverpb.PutResponse{
						Header: txnHeader(rev),
					},
				},
			},
		}
	} else {
		resp.Responses = []*etcdserverpb.ResponseOp{
			{
				Response: &etcdserverpb.ResponseOp_ResponseRange{
					ResponseRange: &etcdserverpb.RangeResponse{
						Header: txnHeader(rev),
						Kvs:    toKVs(kv),
					},
				},
			},
		}
	}

	return resp, nil
}
