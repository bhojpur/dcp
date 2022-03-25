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
	"bytes"
	"context"
	"fmt"
	"strings"

	"go.etcd.io/etcd/api/v3/etcdserverpb"
)

func (l *LimitedServer) list(ctx context.Context, r *etcdserverpb.RangeRequest) (*RangeResponse, error) {
	if len(r.RangeEnd) == 0 {
		return nil, fmt.Errorf("invalid range end length of 0")
	}

	prefix := string(append(r.RangeEnd[:len(r.RangeEnd)-1], r.RangeEnd[len(r.RangeEnd)-1]-1))
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}
	start := string(bytes.TrimRight(r.Key, "\x00"))

	if r.CountOnly {
		rev, count, err := l.backend.Count(ctx, prefix)
		if err != nil {
			return nil, err
		}
		return &RangeResponse{
			Header: txnHeader(rev),
			Count:  count,
		}, nil
	}

	limit := r.Limit
	if limit > 0 {
		limit++
	}

	rev, kvs, err := l.backend.List(ctx, prefix, start, limit, r.Revision)
	if err != nil {
		return nil, err
	}

	resp := &RangeResponse{
		Header: txnHeader(rev),
		Count:  int64(len(kvs)),
		Kvs:    kvs,
	}

	if limit > 0 && resp.Count > r.Limit {
		resp.More = true
		resp.Kvs = kvs[0 : limit-1]
	}

	return resp, nil
}
