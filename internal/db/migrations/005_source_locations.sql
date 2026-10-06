CREATE TABLE IF NOT EXISTS source_locations (
    id TEXT PRIMARY KEY,
    source_id TEXT NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
    start_line INTEGER NOT NULL CHECK (start_line >= 1),
    end_line INTEGER NOT NULL CHECK (end_line >= start_line),
    source_checksum TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_source_locations_source ON source_locations(source_id, start_line, end_line);
