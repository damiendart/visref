package sqlite

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"
	"time"

	"github.com/damiendart/visref/internal/sqlite/migrations"
)

// MainDB represents the SQLite database where permanent data is stored.
type MainDB struct {
	DB
}

// NewMainDB returns a new instance of MainDB. If the given path is an
// empty string, an in-memory database is used.
func NewMainDB(path string, logger *slog.Logger) *MainDB {
	return &MainDB{
		DB{
			logger:      logger,
			migrateFunc: migrateMainDB,
			path:        path,
			Now:         time.Now,
		},
	}
}

func migrateMainDB(db *DB) error {
	var version int

	files, err := fs.Glob(migrations.MainDBMigrations, "*.sql")
	if err != nil {
		return err
	}

	sort.Strings(files)

	if err = db.readOnlyPool.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return err
	}

	if version >= len(files) {
		return nil
	}

	for i, name := range files[version:] {
		contents, err := fs.ReadFile(migrations.MainDBMigrations, name)
		if err != nil {
			return err
		}

		tx, err := db.readWritePool.Begin()
		if err != nil {
			return err
		}

		if _, err = tx.Exec(string(contents)); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return rollbackErr
			}

			return err
		}

		if _, err = tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", version+i+1)); err != nil {
			return err
		}

		if err = tx.Commit(); err != nil {
			return err
		}

		db.logger.LogAttrs(
			context.TODO(),
			slog.LevelInfo,
			"database migration completed",
			slog.GroupAttrs("migration", slog.String("file", name)),
		)
	}

	return nil
}
