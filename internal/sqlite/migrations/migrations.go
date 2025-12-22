// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package migrations

import "embed"

// MainDBMigrations is a read-only collection of SQL files used to
// perform database migrations for the main database.
//
//go:embed main*.sql
var MainDBMigrations embed.FS
