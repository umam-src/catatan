PRAGMA foreign_keys = ON;

-- Potongan teks Catatan dan Sumber untuk pengambilan konteks chat.
-- Pemecahan dilakukan di Go; trigger hanya menandai dokumen yang perlu dipecah ulang.
CREATE TABLE potongan (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    dokumen_id TEXT NOT NULL,
    buku_id TEXT NOT NULL REFERENCES notebooks(id) ON DELETE CASCADE,
    urutan INTEGER NOT NULL,
    baris_awal INTEGER NOT NULL,
    baris_akhir INTEGER NOT NULL,
    isi TEXT NOT NULL
);

CREATE INDEX idx_potongan_dokumen ON potongan(dokumen_id, urutan);
CREATE INDEX idx_potongan_buku ON potongan(buku_id);

CREATE TABLE potongan_tertunda (
    dokumen_id TEXT PRIMARY KEY
);

CREATE VIRTUAL TABLE indeks_potongan USING fts5(
    isi,
    content='potongan',
    content_rowid='id',
    tokenize='unicode61'
);

CREATE TRIGGER potongan_ai AFTER INSERT ON potongan BEGIN
    INSERT INTO indeks_potongan(rowid, isi) VALUES (new.id, new.isi);
END;

CREATE TRIGGER potongan_ad AFTER DELETE ON potongan BEGIN
    INSERT INTO indeks_potongan(indeks_potongan, rowid, isi) VALUES ('delete', old.id, old.isi);
END;

-- Dokumen pencarian dibuat ulang oleh trigger Catatan dan Sumber (hapus lalu sisip),
-- sehingga cukup mengikuti sisip dan hapus pada search_documents.
CREATE TRIGGER search_documents_potongan_ai AFTER INSERT ON search_documents BEGIN
    INSERT OR IGNORE INTO potongan_tertunda(dokumen_id) VALUES (new.id);
END;

CREATE TRIGGER search_documents_potongan_au AFTER UPDATE OF content ON search_documents BEGIN
    DELETE FROM potongan WHERE dokumen_id = old.id;
    INSERT OR IGNORE INTO potongan_tertunda(dokumen_id) VALUES (new.id);
END;

CREATE TRIGGER search_documents_potongan_ad AFTER DELETE ON search_documents BEGIN
    DELETE FROM potongan WHERE dokumen_id = old.id;
    DELETE FROM potongan_tertunda WHERE dokumen_id = old.id;
END;

INSERT OR IGNORE INTO potongan_tertunda(dokumen_id) SELECT id FROM search_documents;
