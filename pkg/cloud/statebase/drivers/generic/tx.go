package generic

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

	"github.com/bhojpur/dcp/pkg/cloud/statebase/server"
	"github.com/sirupsen/logrus"
)

// explicit interface check
var _ server.Transaction = (*Tx)(nil)

type Tx struct {
	x *sql.Tx
	d *Generic
}

func (d *Generic) BeginTx(ctx context.Context, opts *sql.TxOptions) (server.Transaction, error) {
	logrus.Tracef("TX BEGIN")
	x, err := d.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{
		x: x,
		d: d,
	}, nil
}

func (t *Tx) Commit() error {
	logrus.Tracef("TX COMMIT")
	return t.x.Commit()
}

func (t *Tx) MustCommit() {
	if err := t.Commit(); err != nil {
		logrus.Fatalf("Transaction commit failed: %v", err)
	}
}

func (t *Tx) Rollback() error {
	logrus.Tracef("TX ROLLBACK")
	return t.x.Rollback()
}

func (t *Tx) MustRollback() {
	if err := t.Rollback(); err != nil {
		if err != sql.ErrTxDone {
			logrus.Fatalf("Transaction rollback failed: %v", err)
		}
	}
}

func (t *Tx) GetCompactRevision(ctx context.Context) (int64, error) {
	var id int64
	row := t.queryRow(ctx, compactRevSQL)
	err := row.Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return id, err
}

func (t *Tx) SetCompactRevision(ctx context.Context, revision int64) error {
	logrus.Tracef("TX SETCOMPACTREVISION %v", revision)
	_, err := t.execute(ctx, t.d.UpdateCompactSQL, revision)
	return err
}

func (t *Tx) Compact(ctx context.Context, revision int64) (int64, error) {
	logrus.Tracef("TX COMPACT %v", revision)
	res, err := t.execute(ctx, t.d.CompactSQL, revision, revision)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (t *Tx) GetRevision(ctx context.Context, revision int64) (*sql.Rows, error) {
	return t.query(ctx, t.d.GetRevisionSQL, revision)
}

func (t *Tx) DeleteRevision(ctx context.Context, revision int64) error {
	logrus.Tracef("TX DELETEREVISION %v", revision)
	_, err := t.execute(ctx, t.d.DeleteSQL, revision)
	return err
}

func (t *Tx) CurrentRevision(ctx context.Context) (int64, error) {
	var id int64
	row := t.queryRow(ctx, revSQL)
	err := row.Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return id, err
}

func (t *Tx) query(ctx context.Context, sql string, args ...interface{}) (*sql.Rows, error) {
	logrus.Tracef("TX QUERY %v : %s", args, Stripped(sql))
	return t.x.QueryContext(ctx, sql, args...)
}

func (t *Tx) queryRow(ctx context.Context, sql string, args ...interface{}) *sql.Row {
	logrus.Tracef("TX QUERY ROW %v : %s", args, Stripped(sql))
	return t.x.QueryRowContext(ctx, sql, args...)
}

func (t *Tx) execute(ctx context.Context, sql string, args ...interface{}) (result sql.Result, err error) {
	logrus.Tracef("TX EXEC %v : %s", args, Stripped(sql))
	return t.x.ExecContext(ctx, sql, args...)
}
