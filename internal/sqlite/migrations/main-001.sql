-- Copyright (C) Damien Dart, <damiendart@pobox.com>.
-- This file is distributed under the MIT licence. For more information,
-- please refer to the accompanying "LICENCE" file.

CREATE TABLE IF NOT EXISTS items (
  id TEXT PRIMARY KEY,
  alternative_text TEXT,
  source TEXT,
  description TEXT,
  media_type TEXT NOT NULL,
  filepath TEXT NOT NULL,
  original_filename TEXT NOT NULL,
  width INTEGER,
  height INTEGER,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
) STRICT;
