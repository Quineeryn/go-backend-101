-- Users: tambah kolom password & timestamps jika belum ada
ALTER TABLE users ADD COLUMN password_hash TEXT;
ALTER TABLE users ADD COLUMN created_at TEXT DEFAULT (datetime('now')) NOT NULL;
ALTER TABLE users ADD COLUMN updated_at TEXT DEFAULT (datetime('now')) NOT NULL;

-- Refresh tokens (rotating refresh)
CREATE TABLE IF NOT EXISTS refresh_tokens (
  id TEXT PRIMARY KEY,         -- uuid
  user_id TEXT NOT NULL,
  jti TEXT NOT NULL,           -- token id (uuid) tertanam di JWT
  expires_at TEXT NOT NULL,    -- datetime ISO
  revoked_at TEXT,             -- null = aktif
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  FOREIGN KEY (user_id) REFERENCES users(id)
);
CREATE INDEX IF NOT EXISTS idx_refresh_user ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_jti ON refresh_tokens(jti);
