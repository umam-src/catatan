package http

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"testing"
)

func TestExportNotebookPreservesDataAndIsDeterministic(t *testing.T) {
	handler, server := serverUjiDenganServer(t)
	createdBook := requestUji(t, handler, http.MethodPost, "/api/notebooks", map[string]string{
		"title":       "Buku Ekspor",
		"description": "Buku untuk uji ekspor",
	})
	if createdBook.Code != http.StatusCreated {
		t.Fatalf("buat buku: status = %d", createdBook.Code)
	}
	var notebook Notebook
	if err := json.NewDecoder(createdBook.Body).Decode(&notebook); err != nil {
		t.Fatal(err)
	}

	createdNote := requestUji(t, handler, http.MethodPost, "/api/notebooks/"+notebook.ID+"/notes", map[string]string{
		"title":   "Catatan Ekspor",
		"content": "Isi catatan yang harus tetap sama",
	})
	if createdNote.Code != http.StatusCreated {
		t.Fatalf("buat catatan: status = %d", createdNote.Code)
	}

	content := "baris satu\nbaris dua\nbaris tiga"
	digest := sha256.Sum256([]byte(content))
	checksum := hex.EncodeToString(digest[:])
	_, err := server.db.Exec(`INSERT INTO sources(id,notebook_id,title,kind,content,locator,checksum,metadata_json,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?)`,
		"11111111111111111111111111111111", notebook.ID, "Sumber Ekspor", "text", content, "", checksum, `{"filename":"sumber.txt","size":31}`, "2026-01-01T00:00:00Z", "2026-01-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	_, err = server.db.Exec(`INSERT INTO source_locations(id,source_id,start_line,end_line,source_checksum,created_at) VALUES(?,?,?,?,?,?)`,
		"22222222222222222222222222222222", "11111111111111111111111111111111", 1, 2, checksum, "2026-01-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}

	res := requestUji(t, handler, http.MethodGet, "/api/notebooks/"+notebook.ID+"/export", nil)
	if res.Code != http.StatusOK {
		t.Fatalf("ekspor: status = %d", res.Code)
	}
	if got := res.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q", got)
	}
	if got := res.Header().Get("Content-Disposition"); got == "" {
		t.Fatal("Content-Disposition tidak ada")
	}

	var exported ExportDocument
	if err := json.NewDecoder(res.Body).Decode(&exported); err != nil {
		t.Fatal(err)
	}
	if exported.Format != "catatan-export" || exported.Version != exportFormatVersion {
		t.Fatalf("format ekspor = %#v", exported)
	}
	if exported.Notebook.ID != notebook.ID || exported.Notebook.Title != "Buku Ekspor" {
		t.Fatalf("buku hasil ekspor = %#v", exported.Notebook)
	}
	if len(exported.Notes) != 1 || exported.Notes[0].Content != "Isi catatan yang harus tetap sama" {
		t.Fatalf("catatan hasil ekspor = %#v", exported.Notes)
	}
	if len(exported.Sources) != 1 {
		t.Fatalf("jumlah sumber = %d", len(exported.Sources))
	}
	source := exported.Sources[0]
	if source.Content != content || source.Checksum != checksum || source.Metadata != `{"filename":"sumber.txt","size":31}` {
		t.Fatalf("sumber hasil ekspor = %#v", source)
	}
	if len(source.Locations) != 1 || source.Locations[0].StartLine != 1 || source.Locations[0].EndLine != 2 {
		t.Fatalf("lokasi hasil ekspor = %#v", source.Locations)
	}

	encoded, err := json.Marshal(exported)
	if err != nil {
		t.Fatal(err)
	}
	var roundTrip ExportDocument
	if err := json.Unmarshal(encoded, &roundTrip); err != nil {
		t.Fatal(err)
	}
	if roundTrip.Sources[0].Content != source.Content || roundTrip.Sources[0].Checksum != source.Checksum {
		t.Fatal("round-trip mengubah isi atau checksum sumber")
	}
}

func TestExportRejectsCorruptSource(t *testing.T) {
	handler, server := serverUjiDenganServer(t)
	createdBook := requestUji(t, handler, http.MethodPost, "/api/notebooks", map[string]string{"title": "Buku Rusak"})
	if createdBook.Code != http.StatusCreated {
		t.Fatalf("buat buku: status = %d", createdBook.Code)
	}
	var notebook Notebook
	if err := json.NewDecoder(createdBook.Body).Decode(&notebook); err != nil {
		t.Fatal(err)
	}
	_, err := server.db.Exec(`INSERT INTO sources(id,notebook_id,title,kind,content,locator,checksum,metadata_json,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?)`,
		"33333333333333333333333333333333", notebook.ID, "Sumber Rusak", "text", "isi berubah", "", "checksum-salah", `{}`, "2026-01-01T00:00:00Z", "2026-01-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}

	res := requestUji(t, handler, http.MethodGet, "/api/notebooks/"+notebook.ID+"/export", nil)
	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status sumber rusak = %d, ingin %d", res.Code, http.StatusUnprocessableEntity)
	}
}

func TestExportDoesNotExposeAuthenticationData(t *testing.T) {
	handler, _ := serverUjiDenganServer(t)
	createdBook := requestUji(t, handler, http.MethodPost, "/api/notebooks", map[string]string{"title": "Buku Aman"})
	if createdBook.Code != http.StatusCreated {
		t.Fatalf("buat buku: status = %d", createdBook.Code)
	}
	var notebook Notebook
	if err := json.NewDecoder(createdBook.Body).Decode(&notebook); err != nil {
		t.Fatal(err)
	}

	res := requestUji(t, handler, http.MethodGet, "/api/notebooks/"+notebook.ID+"/export", nil)
	if res.Code != http.StatusOK {
		t.Fatalf("ekspor: status = %d", res.Code)
	}
	var raw map[string]any
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["password_hash"]; ok {
		t.Fatal("password_hash ikut diekspor")
	}
	if _, ok := raw["token_hash"]; ok {
		t.Fatal("token_hash ikut diekspor")
	}
}
