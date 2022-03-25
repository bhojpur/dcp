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

func isCreate(txn *etcdserverpb.TxnRequest) *etcdserverpb.PutRequest {
	if len(txn.Compare) == 1 &&
		txn.Compare[0].Target == etcdserverpb.Compare_MOD &&
		txn.Compare[0].Result == etcdserverpb.Compare_EQUAL &&
		txn.Compare[0].GetModRevision() == 0 &&
		len(txn.Failure) == 0 &&
		len(txn.Success) == 1 &&
		txn.Success[0].GetRequestPut() != nil {
		return txn.Success[0].GetRequestPut()
	}
	return nil
}

func (l *LimitedServer) create(ctx context.Context, put *etcdserverpb.PutRequest, txn *etcdserverpb.TxnRequest) (*etcdserverpb.TxnResponse, error) {
	if put.IgnoreLease {
		return nil, unsupported("ignoreLease")
	} else if put.IgnoreValue {
		return nil, unsupported("ignoreValue")
	} else if put.PrevKv {
		return nil, unsupported("prevKv")
	}

	rev, err := l.backend.Create(ctx, string(put.Key), put.Value, put.Lease)
	if err == ErrKeyExists {
		return &etcdserverpb.TxnResponse{
			Header:    txnHeader(rev),
			Succeeded: false,
		}, nil
	} else if err != nil {
		return nil, err
	}

	return &etcdserverpb.TxnResponse{
		Header: txnHeader(rev),
		Responses: []*etcdserverpb.ResponseOp{
			{
				Response: &etcdserverpb.ResponseOp_ResponsePut{
					ResponsePut: &etcdserverpb.PutResponse{
						Header: txnHeader(rev),
					},
				},
			},
		},
		Succeeded: true,
	}, nil
}
