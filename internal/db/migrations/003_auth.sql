PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL DEFAULT '',
    display_name TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS auth_credentials (
    user_id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    password_hash TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TEXT NOT NULL,
    created_at TEXT NOT NULL,
    revoked_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_sessions_token_hash ON sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_sessions_user_active ON sessions(user_id, expires_at);

INSERT OR IGNORE INTO users(id, username, display_name, created_at, updated_at)
VALUES ('local', 'local', 'Pengguna lokal', datetime('now'), datetime('now'));

ALTER TABLE notebooks ADD COLUMN owner_id TEXT REFERENCES users(id) ON DELETE CASCADE;

UPDATE notebooks SET owner_id = 'local' WHERE owner_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_notebooks_owner_updated ON notebooks(owner_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_notes_notebook_deleted ON notes(notebook_id, deleted_at);
