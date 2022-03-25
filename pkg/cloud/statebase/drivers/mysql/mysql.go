package mysql

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
	cryptotls "crypto/tls"
	"database/sql"

	"github.com/bhojpur/dcp/pkg/cloud/statebase/drivers/generic"
	"github.com/bhojpur/dcp/pkg/cloud/statebase/logstructured"
	"github.com/bhojpur/dcp/pkg/cloud/statebase/logstructured/sqllog"
	"github.com/bhojpur/dcp/pkg/cloud/statebase/server"
	"github.com/bhojpur/dcp/pkg/cloud/statebase/tls"
	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

const (
	defaultUnixDSN = "root@unix(/var/run/mysqld/mysqld.sock)/"
	defaultHostDSN = "root@tcp(127.0.0.1)/"
)

var (
	schema = []string{
		`CREATE TABLE IF NOT EXISTS statebase
			(
				id INTEGER AUTO_INCREMENT,
				name VARCHAR(630),
				created INTEGER,
				deleted INTEGER,
				create_revision INTEGER,
 				prev_revision INTEGER,
				lease INTEGER,
				value MEDIUMBLOB,
				old_value MEDIUMBLOB,
				PRIMARY KEY (id)
			);`,
		`CREATE INDEX statebase_name_index ON statebase (name)`,
		`CREATE INDEX statebase_name_id_index ON statebase (name,id)`,
		`CREATE INDEX statebase_id_deleted_index ON statebase (id,deleted)`,
		`CREATE INDEX statebase_prev_revision_index ON statebase (prev_revision)`,
		`CREATE UNIQUE INDEX statebase_name_prev_revision_uindex ON statebase (name, prev_revision)`,
	}
	createDB = "CREATE DATABASE IF NOT EXISTS "
)

func New(ctx context.Context, dataSourceName string, tlsInfo tls.Config, connPoolConfig generic.ConnectionPoolConfig) (server.Backend, error) {
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		return nil, err
	}

	if tlsConfig != nil {
		tlsConfig.MinVersion = cryptotls.VersionTLS11
	}

	parsedDSN, err := prepareDSN(dataSourceName, tlsConfig)
	if err != nil {
		return nil, err
	}

	if err := createDBIfNotExist(parsedDSN); err != nil {
		return nil, err
	}

	dialect, err := generic.Open(ctx, "mysql", parsedDSN, connPoolConfig, "?", false)
	if err != nil {
		return nil, err
	}

	dialect.LastInsertID = true
	dialect.GetSizeSQL = `
		SELECT SUM(data_length + index_length)
		FROM information_schema.TABLES
		WHERE table_schema = DATABASE() AND table_name = 'statebase'`
	dialect.CompactSQL = `
		DELETE kv FROM statebase AS kv
		INNER JOIN (
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
		) AS ks
		ON kv.id = ks.id`
	dialect.TranslateErr = func(err error) error {
		if err, ok := err.(*mysql.MySQLError); ok && err.Number == 1062 {
			return server.ErrKeyExists
		}
		return err
	}
	if err := setup(dialect.DB); err != nil {
		return nil, err
	}

	dialect.Migrate(context.Background())
	return logstructured.New(sqllog.New(dialect)), nil
}

func setup(db *sql.DB) error {
	logrus.Infof("Configuring database table schema and indexes, this may take a moment...")

	for _, stmt := range schema {
		logrus.Tracef("SETUP EXEC : %v", generic.Stripped(stmt))
		_, err := db.Exec(stmt)
		if err != nil {
			if mysqlError, ok := err.(*mysql.MySQLError); !ok || mysqlError.Number != 1061 {
				return err
			}
		}
	}

	logrus.Infof("Database tables and indexes are up to date")
	return nil
}

func createDBIfNotExist(dataSourceName string) error {
	config, err := mysql.ParseDSN(dataSourceName)
	if err != nil {
		return err
	}
	dbName := config.DBName

	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return err
	}
	_, err = db.Exec(createDB + dbName)
	if err != nil {
		if mysqlError, ok := err.(*mysql.MySQLError); !ok || mysqlError.Number != 1049 {
			return err
		}
		config.DBName = ""
		db, err = sql.Open("mysql", config.FormatDSN())
		if err != nil {
			return err
		}
		_, err = db.Exec(createDB + dbName)
		if err != nil {
			return err
		}
	}
	return nil
}

func prepareDSN(dataSourceName string, tlsConfig *cryptotls.Config) (string, error) {
	if len(dataSourceName) == 0 {
		dataSourceName = defaultUnixDSN
		if tlsConfig != nil {
			dataSourceName = defaultHostDSN
		}
	}
	config, err := mysql.ParseDSN(dataSourceName)
	if err != nil {
		return "", err
	}
	// setting up tlsConfig
	if tlsConfig != nil {
		if err := mysql.RegisterTLSConfig("statebase", tlsConfig); err != nil {
			return "", err
		}
		config.TLSConfig = "statebase"
	}
	dbName := "kubernetes"
	if len(config.DBName) > 0 {
		dbName = config.DBName
	}
	config.DBName = dbName
	parsedDSN := config.FormatDSN()

	return parsedDSN, nil
}
