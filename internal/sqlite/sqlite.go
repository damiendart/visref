// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package sqlite

import (
	"context"
	"crypto/rand"
	"database/sql"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"time"

	// This package is only imported for its side effect of registering
	// the "sqlite3" driver for use with the "database/sql" package.
	_ "github.com/mattn/go-sqlite3"
)

// DB represents an SQLite database.
type DB struct {
	logger        *slog.Logger
	migrateFunc   func(*DB) error
	path          string
	readOnlyPool  *sql.DB
	readWritePool *sql.DB
	Now           func() time.Time
}

// Tx provides a sql.Tx and a transaction start timestamp.
type Tx struct {
	*sql.Tx
	Now time.Time
}

// BeginTx starts a transaction using the read-write database connection
// and returns a Tx.
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.readWritePool.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &Tx{tx, db.Now().UTC().Truncate(time.Second)}, nil
}

// Open opens reading and writing database connections and executes any
// database migrations.
func (db *DB) Open() (err error) {
	dsnParams := url.Values{}

	// Allow two or more distinct but shareable in-memory databases to
	// run in a single process. For more information, please see
	// <https://www.sqlite.org/inmemorydb.html#sharedmemdb>.
	if db.path == "" || db.path == ":memory:" {
		db.path = "memory_database_" + rand.Text()

		dsnParams.Add("cache", "shared")
		dsnParams.Add("mode", "memory")
	} else {
		if err := os.MkdirAll(filepath.Dir(db.path), 0700); err != nil {
			return err
		}
	}

	dsnParams.Add("_busy_timeout", "5000")
	dsnParams.Add("_cache_size", "1000000000")
	dsnParams.Add("_foreign_keys", "true")
	dsnParams.Add("_journal_mode", "WAL")
	dsnParams.Add("_synchronous", "NORMAL")
	dsnParams.Add("_txlock", "immediate")

	dsn := "file:" + db.path + "?" + dsnParams.Encode()

	db.readWritePool, err = sql.Open("sqlite3", dsn)
	if err != nil {
		return err
	}

	db.readWritePool.SetMaxOpenConns(1)
	_, err = db.readWritePool.Exec("PRAGMA temp_store = memory")
	if err != nil {
		return err
	}

	dsnParams.Add("mode", "ro")

	db.readOnlyPool, err = sql.Open("sqlite3", dsn)
	if err != nil {
		return err
	}

	db.readOnlyPool.SetMaxOpenConns(max(4, runtime.NumCPU()))
	_, err = db.readOnlyPool.Exec("PRAGMA temp_store = memory")
	if err != nil {
		return err
	}

	err = db.migrateFunc(db)
	if err != nil {
		return err
	}

	return nil
}

// QueryContext executes a query that returns rows using the read-only
// database connection.
func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.readOnlyPool.QueryContext(ctx, query, args...)
}

// QueryRowContext executes a query that is expected to return at most
// one row using the read-only database connection.
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return db.readOnlyPool.QueryRowContext(ctx, query, args...)
}
