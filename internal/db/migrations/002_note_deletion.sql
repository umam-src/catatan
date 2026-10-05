ALTER TABLE notes ADD COLUMN deleted_at TEXT;

CREATE INDEX IF NOT EXISTS idx_notes_notebook_updated_active
ON notes(notebook_id, updated_at DESC)
WHERE deleted_at IS NULL;
