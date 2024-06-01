package sqlite

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"

	// This package is only imported for its side effect of registering
	// the "sqlite3" driver for use with the "database/sql" package.
	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrations embed.FS

// DB represents a SQLite database connection.
type DB struct {
	logger    *slog.Logger
	path      string
	readPool  *sql.DB
	writePool *sql.DB
}

// NewDB returns a new instance of DB.
func NewDB(path string, logger *slog.Logger) *DB {
	if path == "" {
		path = ":memory:"
	}

	db := &DB{logger: logger, path: path}

	return db
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

	db.writePool, err = sql.Open("sqlite3", dsn)
	if err != nil {
		return err
	}

	db.writePool.SetMaxOpenConns(1)
	_, err = db.writePool.Exec("PRAGMA temp_store = memory")
	if err != nil {
		return err
	}

	db.readPool, err = sql.Open("sqlite3", dsn)
	if err != nil {
		return err
	}

	db.readPool.SetMaxOpenConns(max(4, runtime.NumCPU()))
	_, err = db.readPool.Exec("PRAGMA temp_store = memory")
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

	files, err := fs.Glob(migrations, "migrations/*.sql")
	if err != nil {
		return err
	}
	sort.Strings(files)

	err = db.readPool.QueryRow("PRAGMA user_version").Scan(&version)
	if err != nil {
		return err
	}

	if version >= len(files) {
		return nil
	}

	for i, name := range files[version:] {
		contents, err := fs.ReadFile(migrations, name)
		if err != nil {
			return err
		}

		tx, err := db.writePool.Begin()
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
