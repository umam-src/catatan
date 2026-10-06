CREATE INDEX IF NOT EXISTS idx_notes_notebook_deleted_updated
    ON notes(notebook_id, deleted_at, updated_at DESC);
