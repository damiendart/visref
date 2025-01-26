-- Copyright (C) Damien Dart, <damiendart@pobox.com>.
-- This file is distributed under the MIT licence. For more information,
-- please refer to the accompanying "LICENCE" file.

CREATE TABLE IF NOT EXISTS items (
  id TEXT PRIMARY KEY,
  alternative_text TEXT,
  description TEXT,
  mime_type TEXT NOT NULL,
  filepath TEXT NOT NULL,
  original_filename TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
) STRICT;

CREATE INDEX IF NOT EXISTS idx_items_alternative_text ON items(alternative_text);
