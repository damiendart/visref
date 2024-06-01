CREATE TABLE IF NOT EXISTS media (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  description TEXT
);

CREATE INDEX IF NOT EXISTS idx_media_title ON media(title);
