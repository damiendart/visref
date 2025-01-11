package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	// This package is only imported for its side effect of registering
	// the "sqlite3" driver for use with the "database/sql" package.
	_ "github.com/mattn/go-sqlite3"

	"github.com/damiendart/visref/internal/sqlite/migrations"
)

// DB represents a SQLite database connection.
type DB struct {
	logger        *slog.Logger
	migrations    fs.FS
	path          string
	readOnlyPool  *sql.DB
	readWritePool *sql.DB
	Now           func() time.Time
}

// MainDB represents the main database where permanent data is stored.
type MainDB struct {
	DB
}

// Tx provides a sql.Tx and a transaction start timestamp.
type Tx struct {
	*sql.Tx
	now time.Time
}

// NewMainDB returns a new instance of DB for the main database.
func NewMainDB(path string, logger *slog.Logger) *MainDB {
	if path == "" {
		path = ":memory:"
	}

	return &MainDB{
		DB{
			logger:     logger,
			migrations: migrations.MainDBMigrations,
			path:       path,
			Now:        time.Now,
		},
	}
}

// BeginTx starts a transaction.
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.readWritePool.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &Tx{tx, db.Now().UTC().Truncate(time.Second)}, nil
}

// Open opens reading and writing database connections and executes any
// outstanding database migrations.
func (db *DB) Open() (err error) {
	if db.path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(db.path), 0700); err != nil {
			return err
		}
	}

	dsnParams := url.Values{}
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

	db.readOnlyPool, err = sql.Open("sqlite3", dsn)
	if err != nil {
		return err
	}

	db.readOnlyPool.SetMaxOpenConns(max(4, runtime.NumCPU()))
	_, err = db.readOnlyPool.Exec("PRAGMA temp_store = memory")
	if err != nil {
		return err
	}

	err = db.migrate()
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) migrate() error {
	var version int

	files, err := fs.Glob(db.migrations, "*.sql")
	if err != nil {
		return err
	}
	sort.Strings(files)

	err = db.readOnlyPool.QueryRow("PRAGMA user_version").Scan(&version)
	if err != nil {
		return err
	}

	if version >= len(files) {
		return nil
	}

	for i, name := range files[version:] {
		contents, err := fs.ReadFile(db.migrations, name)
		if err != nil {
			return err
		}

		tx, err := db.readWritePool.Begin()
		if err != nil {
			return err
		}

		if _, err = tx.Exec(string(contents)); err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				return rollbackErr
			}

			return err
		}

		_, err = tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", version+i+1))
		if err != nil {
			return err
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

		db.logger.Info("database migration completed", slog.Group("migration", "file", name))
	}

	return nil
}
