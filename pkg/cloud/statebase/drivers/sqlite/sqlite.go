//go:build cgo
// +build cgo

package sqlite

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
	"os"
	"time"

	"github.com/bhojpur/dcp/pkg/cloud/statebase/drivers/generic"
	"github.com/bhojpur/dcp/pkg/cloud/statebase/logstructured"
	"github.com/bhojpur/dcp/pkg/cloud/statebase/logstructured/sqllog"
	"github.com/bhojpur/dcp/pkg/cloud/statebase/server"
	"github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	// sqlite db driver
	_ "github.com/mattn/go-sqlite3"
)

var (
	schema = []string{
		`CREATE TABLE IF NOT EXISTS statebase
			(
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name INTEGER,
				created INTEGER,
				deleted INTEGER,
				create_revision INTEGER,
				prev_revision INTEGER,
				lease INTEGER,
				value BLOB,
				old_value BLOB
			)`,
		`CREATE INDEX IF NOT EXISTS statebase_name_index ON statebase (name)`,
		`CREATE INDEX IF NOT EXISTS statebase_name_id_index ON statebase (name,id)`,
		`CREATE INDEX IF NOT EXISTS statebase_id_deleted_index ON statebase (id,deleted)`,
		`CREATE INDEX IF NOT EXISTS statebase_prev_revision_index ON statebase (prev_revision)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS statebase_name_prev_revision_uindex ON statebase (name, prev_revision)`,
		`PRAGMA wal_checkpoint(TRUNCATE)`,
	}
)

func New(ctx context.Context, dataSourceName string, connPoolConfig generic.ConnectionPoolConfig) (server.Backend, error) {
	backend, _, err := NewVariant(ctx, "sqlite3", dataSourceName, connPoolConfig)
	return backend, err
}

func NewVariant(ctx context.Context, driverName, dataSourceName string, connPoolConfig generic.ConnectionPoolConfig) (server.Backend, *generic.Generic, error) {
	if dataSourceName == "" {
		if err := os.MkdirAll("./db", 0700); err != nil {
			return nil, nil, err
		}
		dataSourceName = "./db/state.db?_journal=WAL&cache=shared"
	}

	dialect, err := generic.Open(ctx, driverName, dataSourceName, connPoolConfig, "?", false)
	if err != nil {
		return nil, nil, err
	}

	dialect.LastInsertID = true
	dialect.GetSizeSQL = `SELECT SUM(pgsize) FROM dbstat`
	dialect.CompactSQL = `
		DELETE FROM statebase AS kv
		WHERE
			kv.id IN (
				SELECT kp.prev_revision AS id
				FROM statebase AS kp
				WHERE
					kp.name != 'compact_rev_key' AND
					kp.prev_revision != 0 AND
					kp.id <= ?
				UNION
				SELECT kd.id AS id
				FROM statebase AS kd
				WHERE
					kd.deleted != 0 AND
					kd.id <= ?
			)`
	dialect.PostCompactSQL = `PRAGMA wal_checkpoint(FULL)`
	dialect.TranslateErr = func(err error) error {
		if err, ok := err.(sqlite3.Error); ok && err.ExtendedCode == sqlite3.ErrConstraintUnique {
			return server.ErrKeyExists
		}
		return err
	}

	// this is the first SQL that will be executed on a new DB conn so
	// loop on failure here because in the case of dqlite it could still be initializing
	for i := 0; i < 300; i++ {
		err = setup(dialect.DB)
		if err == nil {
			break
		}
		logrus.Errorf("failed to setup db: %v", err)
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-time.After(time.Second):
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		return nil, nil, errors.Wrap(err, "setup db")
	}

	dialect.Migrate(context.Background())
	return logstructured.New(sqllog.New(dialect)), dialect, nil
}

func setup(db *sql.DB) error {
	logrus.Infof("Configuring database table schema and indexes, this may take a moment...")

	for _, stmt := range schema {
		logrus.Tracef("SETUP EXEC : %v", generic.Stripped(stmt))
		_, err := db.Exec(stmt)
		if err != nil {
			return err
		}
	}

	logrus.Infof("Database tables and indexes are up to date")
	return nil
}
