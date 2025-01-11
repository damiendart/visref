package migrations

import "embed"

// MainDBMigrations is a read-only collection of SQL files used to
// perform database migrations for the main database.
//
//go:embed main*.sql
var MainDBMigrations embed.FS
