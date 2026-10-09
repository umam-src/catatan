package db

import (
	"context"
	"testing"
)

func TestSearchFindsNotesAndSourcesAndKeepsOwnership(t *testing.T) {
	d, _ := bukaDBUji(t)
	if _, err := d.Exec(`INSERT INTO users(id, username, display_name, created_at, updated_at) VALUES (?, ?, ?, datetime("now"), datetime("now"))`, "user-2", "pengguna2", "Pengguna 2"); err != nil { t.Fatal(err) }
	if _, err := d.Exec(`INSERT INTO notebooks(id, owner_id, title, description, created_at, updated_at) VALUES (?, ?, ?, "", datetime("now"), datetime("now"))`, "buku-2", "user-2", "Buku Dua"); err != nil { t.Fatal(err) }
	if _, err := d.Exec(`INSERT INTO notes(id, notebook_id, title, content, created_at, updated_at) VALUES (?, "default", ?, ?, datetime("now"), datetime("now"))`, "note-1", "Belajar", "SQLite lokal tanpa jaringan"); err != nil { t.Fatal(err) }
	if _, err := d.Exec(`INSERT INTO sources(id, notebook_id, title, content, checksum, created_at, updated_at) VALUES (?, "default", ?, ?, "uji", datetime("now"), datetime("now"))`, "source-1", "Dokumen", "Referensi SQLite lokal"); err != nil { t.Fatal(err) }
	if _, err := d.Exec(`INSERT INTO notes(id, notebook_id, title, content, created_at, updated_at) VALUES (?, "buku-2", ?, ?, datetime("now"), datetime("now"))`, "note-2", "Rahasia", "SQLite lokal pengguna lain"); err != nil { t.Fatal(err) }

	results, err := d.Search(context.Background(), "local", "", "SQLite lokal", 20)
	if err != nil { t.Fatal(err) }
	if len(results) != 2 { t.Fatalf("hasil = %d, ingin 2: %#v", len(results), results) }
	for _, result := range results { if result.NotebookID != "default" { t.Fatalf("hasil lintas pemilik: %#v", result) } }
	if results[0].Relevance > results[1].Relevance { t.Fatalf("hasil tidak terurut relevansi: %#v", results) }
}

func TestSearchRespectsNotebookAndLimit(t *testing.T) {
	d, _ := bukaDBUji(t)
	for i, id := range []string{"note-a", "note-b", "note-c"} {
		if _, err := d.Exec(`INSERT INTO notes(id, notebook_id, title, content, created_at, updated_at) VALUES (?, "default", ?, ?, datetime("now"), datetime("now"))`, id, "Judul "+id, "kata kunci pencarian"); err != nil { t.Fatal(err) }
		_ = i
	}
	results, err := d.Search(context.Background(), "local", "default", "pencarian", 2)
	if err != nil { t.Fatal(err) }
	if len(results) != 2 { t.Fatalf("jumlah hasil = %d, ingin 2", len(results)) }
}

func TestSearchIndexFollowsUpdatesAndSoftDeletion(t *testing.T) {
	d, _ := bukaDBUji(t)
	if _, err := d.Exec(`INSERT INTO notes(id, notebook_id, title, content, created_at, updated_at) VALUES ("note-sync", "default", "Lama", "isi lama", datetime("now"), datetime("now"))`); err != nil { t.Fatal(err) }
	results, err := d.Search(context.Background(), "local", "", "lama", 20); if err != nil { t.Fatal(err) }; if len(results) != 1 { t.Fatalf("hasil awal = %#v", results) }
	if _, err := d.Exec(`UPDATE notes SET title="Baru", content="isi baru" WHERE id="note-sync"`); err != nil { t.Fatal(err) }
	results, err = d.Search(context.Background(), "local", "", "lama", 20); if err != nil { t.Fatal(err) }; if len(results) != 0 { t.Fatalf("hasil lama masih ada = %#v", results) }
	results, err = d.Search(context.Background(), "local", "", "baru", 20); if err != nil { t.Fatal(err) }; if len(results) != 1 { t.Fatalf("hasil baru = %#v", results) }
	if _, err := d.Exec(`UPDATE notes SET deleted_at=datetime("now") WHERE id="note-sync"`); err != nil { t.Fatal(err) }
	results, err = d.Search(context.Background(), "local", "", "baru", 20); if err != nil { t.Fatal(err) }; if len(results) != 0 { t.Fatalf("catatan terhapus masih ditemukan = %#v", results) }
}

func TestSearchTreatsQueryAsPlainText(t *testing.T) {
	d, _ := bukaDBUji(t)
	if _, err := d.Exec(`INSERT INTO notes(id, notebook_id, title, content, created_at, updated_at) VALUES (?, "default", ?, ?, datetime("now"), datetime("now"))`, "note-code", "Paket KJ-7319", `Kode pengiriman: KJ-7319 status "menunggu" di gudang`); err != nil { t.Fatal(err) }

	for _, query := range []string{"KJ-7319", "KJ 7319", `"menunggu"`} {
		results, err := d.Search(context.Background(), "local", "default", query, 20)
		if err != nil { t.Fatalf("query %q menghasilkan galat: %v", query, err) }
		if len(results) != 1 || results[0].ID != "note:note-code" { t.Fatalf("query %q: hasil = %#v", query, results) }
	}

	results, err := d.Search(context.Background(), "local", "", "   ", 20)
	if err != nil { t.Fatal(err) }
	if len(results) != 0 { t.Fatalf("query kosong = %#v", results) }

	results, err = d.Search(context.Background(), "local", "", "---", 20)
	if err != nil { t.Fatalf("query tanda baca saja menghasilkan galat: %v", err) }
	if len(results) != 0 { t.Fatalf("query tanda baca saja = %#v", results) }
}

func BenchmarkSearch(b *testing.B) {
	d, _ := bukaDBUji(b)
	for i := 0; i < 1000; i++ {
		if _, err := d.Exec(`INSERT INTO notes(id, notebook_id, title, content, created_at, updated_at) VALUES (?, "default", ?, ?, datetime("now"), datetime("now"))`, newIDForBenchmark(i), "Catatan lokal", "isi catatan dengan kata kunci benchmark dan pencarian offline"); err != nil { b.Fatal(err) }
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ { if _, err := d.Search(context.Background(), "local", "", "benchmark pencarian", 20); err != nil { b.Fatal(err) } }
}

func newIDForBenchmark(i int) string { return "benchmark-" + string(rune("a"[0]+byte(i/26))) + string(rune("a"[0]+byte(i%26))) }
