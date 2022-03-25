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
	"database/sql"

	"go.etcd.io/etcd/etcdserver/api/v3rpc/rpctypes"
)

var (
	ErrKeyExists = rpctypes.ErrGRPCDuplicateKey
	ErrCompacted = rpctypes.ErrGRPCCompacted
)

type Backend interface {
	Start(ctx context.Context) error
	Get(ctx context.Context, key string, revision int64) (int64, *KeyValue, error)
	Create(ctx context.Context, key string, value []byte, lease int64) (int64, error)
	Delete(ctx context.Context, key string, revision int64) (int64, *KeyValue, bool, error)
	List(ctx context.Context, prefix, startKey string, limit, revision int64) (int64, []*KeyValue, error)
	Count(ctx context.Context, prefix string) (int64, int64, error)
	Update(ctx context.Context, key string, value []byte, revision, lease int64) (int64, *KeyValue, bool, error)
	Watch(ctx context.Context, key string, revision int64) <-chan []*Event
	DbSize(ctx context.Context) (int64, error)
}

type Dialect interface {
	ListCurrent(ctx context.Context, prefix string, limit int64, includeDeleted bool) (*sql.Rows, error)
	List(ctx context.Context, prefix, startKey string, limit, revision int64, includeDeleted bool) (*sql.Rows, error)
	Count(ctx context.Context, prefix string) (int64, int64, error)
	CurrentRevision(ctx context.Context) (int64, error)
	After(ctx context.Context, prefix string, rev, limit int64) (*sql.Rows, error)
	Insert(ctx context.Context, key string, create, delete bool, createRevision, previousRevision int64, ttl int64, value, prevValue []byte) (int64, error)
	GetRevision(ctx context.Context, revision int64) (*sql.Rows, error)
	DeleteRevision(ctx context.Context, revision int64) error
	GetCompactRevision(ctx context.Context) (int64, error)
	SetCompactRevision(ctx context.Context, revision int64) error
	Compact(ctx context.Context, revision int64) (int64, error)
	PostCompact(ctx context.Context) error
	Fill(ctx context.Context, revision int64) error
	IsFill(key string) bool
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Transaction, error)
	GetSize(ctx context.Context) (int64, error)
}

type Transaction interface {
	Commit() error
	MustCommit()
	Rollback() error
	MustRollback()
	GetCompactRevision(ctx context.Context) (int64, error)
	SetCompactRevision(ctx context.Context, revision int64) error
	Compact(ctx context.Context, revision int64) (int64, error)
	GetRevision(ctx context.Context, revision int64) (*sql.Rows, error)
	DeleteRevision(ctx context.Context, revision int64) error
	CurrentRevision(ctx context.Context) (int64, error)
}

type KeyValue struct {
	Key            string
	CreateRevision int64
	ModRevision    int64
	Value          []byte
	Lease          int64
}

type Event struct {
	Delete bool
	Create bool
	KV     *KeyValue
	PrevKV *KeyValue
}
