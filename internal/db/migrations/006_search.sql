PRAGMA foreign_keys = ON;

CREATE TABLE search_documents (
    id TEXT PRIMARY KEY,
    notebook_id TEXT NOT NULL REFERENCES notebooks(id) ON DELETE CASCADE,
    kind TEXT NOT NULL CHECK (kind IN ('note', 'source')),
    title TEXT NOT NULL,
    content TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_search_documents_notebook ON search_documents(notebook_id);

CREATE VIRTUAL TABLE search_index USING fts5(
    title,
    content,
    content='search_documents',
    content_rowid='rowid',
    tokenize='unicode61'
);

INSERT INTO search_documents(id, notebook_id, kind, title, content)
SELECT 'note:' || id, notebook_id, 'note', title, content
FROM notes
WHERE deleted_at IS NULL;

INSERT INTO search_documents(id, notebook_id, kind, title, content)
SELECT 'source:' || id, notebook_id, 'source', title, content
FROM sources;

INSERT INTO search_index(rowid, title, content)
SELECT rowid, title, content FROM search_documents;

CREATE TRIGGER search_documents_ai AFTER INSERT ON search_documents BEGIN
    INSERT INTO search_index(rowid, title, content)
    VALUES (new.rowid, new.title, new.content);
END;

CREATE TRIGGER search_documents_ad AFTER DELETE ON search_documents BEGIN
    INSERT INTO search_index(search_index, rowid, title, content)
    VALUES ('delete', old.rowid, old.title, old.content);
END;

CREATE TRIGGER search_documents_au AFTER UPDATE OF title, content ON search_documents BEGIN
    INSERT INTO search_index(search_index, rowid, title, content)
    VALUES ('delete', old.rowid, old.title, old.content);
    INSERT INTO search_index(rowid, title, content)
    VALUES (new.rowid, new.title, new.content);
END;

CREATE TRIGGER notes_search_ai AFTER INSERT ON notes
WHEN new.deleted_at IS NULL
BEGIN
    INSERT INTO search_documents(id, notebook_id, kind, title, content)
    VALUES ('note:' || new.id, new.notebook_id, 'note', new.title, new.content);
END;

CREATE TRIGGER notes_search_au AFTER UPDATE OF notebook_id, title, content, deleted_at ON notes
BEGIN
    DELETE FROM search_documents WHERE id = 'note:' || old.id;
    INSERT INTO search_documents(id, notebook_id, kind, title, content)
    SELECT 'note:' || new.id, new.notebook_id, 'note', new.title, new.content
    WHERE new.deleted_at IS NULL;
END;

CREATE TRIGGER notes_search_ad AFTER DELETE ON notes BEGIN
    DELETE FROM search_documents WHERE id = 'note:' || old.id;
END;

CREATE TRIGGER sources_search_ai AFTER INSERT ON sources BEGIN
    INSERT INTO search_documents(id, notebook_id, kind, title, content)
    VALUES ('source:' || new.id, new.notebook_id, 'source', new.title, new.content);
END;

CREATE TRIGGER sources_search_au AFTER UPDATE OF notebook_id, title, content ON sources BEGIN
    DELETE FROM search_documents WHERE id = 'source:' || old.id;
    INSERT INTO search_documents(id, notebook_id, kind, title, content)
    VALUES ('source:' || new.id, new.notebook_id, 'source', new.title, new.content);
END;

CREATE TRIGGER sources_search_ad AFTER DELETE ON sources BEGIN
    DELETE FROM search_documents WHERE id = 'source:' || old.id;
END;
