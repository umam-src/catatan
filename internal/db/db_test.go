package db

import (
	"context"
	"path/filepath"
	"testing"
)

func bukaDBUji(t *testing.T) (*DB, string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "catatan.db")
	d, err := Open(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = d.Close() })
	return d, path
}

func TestOpenRunsMigrationAndCreatesDefaultNotebook(t *testing.T) {
	d, path := bukaDBUji(t)

	var count int
	if err := d.QueryRow(`SELECT COUNT(*) FROM notebooks WHERE id='default'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("buku awal = %d, ingin 1", count)
	}

	var version int
	if err := d.QueryRow(`SELECT MAX(version) FROM schema_migrations`).Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version != 3 {
		t.Fatalf("versi migrasi = %d, ingin 3", version)
	}

	if err := d.Close(); err != nil {
		t.Fatal(err)
	}

	d, err := Open(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}

	if err := d.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatalf("jumlah migrasi setelah buka ulang = %d, ingin 3", count)
	}
}

func TestMigrationCreatesRequiredSchema(t *testing.T) {
	d, _ := bukaDBUji(t)

	tables := []string{
		"metadata",
		"notebooks",
		"sources",
		"notes",
		"conversations",
		"messages",
	}
	for _, table := range tables {
		t.Run(table, func(t *testing.T) {
			var count int
			if err := d.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&count); err != nil {
				t.Fatal(err)
			}
			if count != 1 {
				t.Fatalf("tabel %q tidak ditemukan", table)
			}
		})
	}

	indexes := []string{
		"idx_notebooks_updated",
		"idx_sources_notebook_updated",
		"idx_notes_notebook_updated",
		"idx_conversations_notebook_updated",
		"idx_messages_conversation_created",
	}
	for _, index := range indexes {
		t.Run("indeks_"+index, func(t *testing.T) {
			var count int
			if err := d.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?`, index).Scan(&count); err != nil {
				t.Fatal(err)
			}
			if count != 1 {
				t.Fatalf("indeks %q tidak ditemukan", index)
			}
		})
	}
}

func TestNoteSupportsSoftDeletion(t *testing.T) {
	d, _ := bukaDBUji(t)

	var count int
	if err := d.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('notes') WHERE name='deleted_at'`).Scan(&count); err != nil { t.Fatal(err) }
	if count != 1 { t.Fatalf("kolom deleted_at = %d, ingin 1", count) }
}

func TestDatabasePragmas(t *testing.T) {
	d, _ := bukaDBUji(t)

	var journalMode string
	if err := d.QueryRow(`PRAGMA journal_mode`).Scan(&journalMode); err != nil {
		t.Fatal(err)
	}
	if journalMode != "wal" {
		t.Fatalf("journal_mode = %q, ingin wal", journalMode)
	}

	var foreignKeys int
	if err := d.QueryRow(`PRAGMA foreign_keys`).Scan(&foreignKeys); err != nil {
		t.Fatal(err)
	}
	if foreignKeys != 1 {
		t.Fatalf("foreign_keys = %d, ingin 1", foreignKeys)
	}
}

func TestNotebookAndNoteReadWrite(t *testing.T) {
	d, _ := bukaDBUji(t)

	_, err := d.Exec(`INSERT INTO notebooks(id, title, description, created_at, updated_at) VALUES ('uji', 'Buku Uji', '', datetime('now'), datetime('now'))`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.Exec(`INSERT INTO notes(id, notebook_id, title, content, note_type, metadata_json, created_at, updated_at) VALUES ('catatan-uji', 'uji', 'Catatan Uji', 'Isi uji', 'human', '{}', datetime('now'), datetime('now'))`)
	if err != nil {
		t.Fatal(err)
	}

	var title, content string
	if err := d.QueryRow(`SELECT title, content FROM notes WHERE id='catatan-uji'`).Scan(&title, &content); err != nil {
		t.Fatal(err)
	}
	if title != "Catatan Uji" || content != "Isi uji" {
		t.Fatalf("catatan = %q / %q, tidak sesuai", title, content)
	}
}

func TestTransactionRollsBackOnFailure(t *testing.T) {
	d, _ := bukaDBUji(t)

	before := hitungBaris(t, d, `SELECT COUNT(*) FROM notebooks`)

	tx, err := d.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`INSERT INTO notebooks(id, title, description, created_at, updated_at) VALUES ('default', 'Bentrok', '', datetime('now'), datetime('now'))`); err == nil {
		t.Fatal("duplikasi id seharusnya gagal")
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}

	after := hitungBaris(t, d, `SELECT COUNT(*) FROM notebooks`)
	if after != before {
		t.Fatalf("jumlah buku setelah rollback = %d, sebelum = %d", after, before)
	}
}

func hitungBaris(t *testing.T, d *DB, query string, args ...any) int {
	t.Helper()

	var count int
	if err := d.QueryRow(query, args...).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func TestAuthSchemaAndOwnership(t *testing.T) {
	d, _ := bukaDBUji(t)

	var userCount int
	if err := d.QueryRow(`SELECT COUNT(*) FROM users WHERE id='local'`).Scan(&userCount); err != nil {
		t.Fatal(err)
	}
	if userCount != 1 {
		t.Fatalf("pengguna lokal = %d, ingin 1", userCount)
	}

	var owner string
	if err := d.QueryRow(`SELECT owner_id FROM notebooks WHERE id='default'`).Scan(&owner); err != nil {
		t.Fatal(err)
	}
	if owner != "local" {
		t.Fatalf("pemilik buku awal = %q, ingin local", owner)
	}
}
