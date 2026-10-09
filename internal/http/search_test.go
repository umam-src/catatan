package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/umam-src/catatan/internal/db"
)

func decodeJSON(t *testing.T, res *httptest.ResponseRecorder, target any) { t.Helper(); if err := json.NewDecoder(res.Body).Decode(target); err != nil { t.Fatal(err) } }

func TestSearchEndpointRequiresSession(t *testing.T) {
	path := filepath.Join(t.TempDir(), "catatan.db")
	d, err := db.Open(context.Background(), path)
	if err != nil { t.Fatal(err) }
	t.Cleanup(func() { _ = d.Close() })
	res := requestUji(t, New(d), http.MethodGet, "/api/search?q=catatan", nil)
	if res.Code != http.StatusUnauthorized { t.Fatalf("status = %d, ingin %d", res.Code, http.StatusUnauthorized) }
}

func TestSearchEndpointReturnsOwnedResultsOnly(t *testing.T) {
	handler, server := serverUjiDenganServer(t)
	if _, err := server.db.Exec(`INSERT INTO notes(id, notebook_id, title, content, created_at, updated_at) VALUES (?, "default", ?, ?, datetime("now"), datetime("now"))`, "note-search", "Catatan Rahasia", "isi pencarian"); err != nil { t.Fatal(err) }
	res := requestUji(t, handler, http.MethodGet, "/api/search?q=pencarian", nil)
	if res.Code != http.StatusOK { t.Fatalf("status = %d, ingin %d; body = %s", res.Code, http.StatusOK, res.Body.String()) }
	var payload searchResponse; decodeJSON(t, res, &payload)
	if len(payload.Results) != 1 || payload.Results[0].ID != "note:note-search" { t.Fatalf("hasil = %#v", payload.Results) }
	if payload.Results[0].Title != "Catatan Rahasia" { t.Fatalf("judul hasil = %q", payload.Results[0].Title) }
}

func TestSearchEndpointFiltersNotebookAndLimit(t *testing.T) {
	handler, server := serverUjiDenganServer(t)
	res := requestUji(t, handler, http.MethodPost, "/api/notebooks", map[string]string{"title": "Buku Pencarian"})
	if res.Code != http.StatusCreated { t.Fatalf("buat buku: status = %d", res.Code) }
	var notebook Notebook; decodeJSON(t, res, &notebook)
	for i, title := range []string{"Pertama", "Kedua", "Ketiga"} {
		id := string(rune("a"[0] + byte(i)))
		if _, err := server.db.Exec(`INSERT INTO notes(id, notebook_id, title, content, created_at, updated_at) VALUES (?, ?, ?, "kata pencarian", datetime("now"), datetime("now"))`, "note-"+id, notebook.ID, title); err != nil { t.Fatal(err) }
	}
	res = requestUji(t, handler, http.MethodGet, "/api/search?q=pencarian&notebook_id="+notebook.ID+"&limit=2", nil)
	if res.Code != http.StatusOK { t.Fatalf("status = %d", res.Code) }
	var payload searchResponse; decodeJSON(t, res, &payload)
	if len(payload.Results) != 2 { t.Fatalf("jumlah hasil = %d, ingin 2", len(payload.Results)) }
	for _, result := range payload.Results { if result.NotebookID != notebook.ID { t.Fatalf("hasil buku lain: %#v", result) } }
}

func TestSearchEndpointHandlesEmptyAndSpecialCharacterQueries(t *testing.T) {
	handler, server := serverUjiDenganServer(t)
	if _, err := server.db.Exec(`INSERT INTO notes(id, notebook_id, title, content, created_at, updated_at) VALUES ("note-code", "default", "Paket KJ-7319", "Kode pengiriman KJ-7319", datetime("now"), datetime("now"))`); err != nil { t.Fatal(err) }

	res := requestUji(t, handler, http.MethodGet, "/api/search?q=%20%20", nil)
	if res.Code != http.StatusOK { t.Fatalf("query kosong: status = %d", res.Code) }

	res = requestUji(t, handler, http.MethodGet, "/api/search?q=KJ-7319", nil)
	if res.Code != http.StatusOK { t.Fatalf("query bertanda hubung: status = %d; body = %s", res.Code, res.Body.String()) }
	var payload searchResponse
	decodeJSON(t, res, &payload)
	if len(payload.Results) != 1 || payload.Results[0].ID != "note:note-code" { t.Fatalf("hasil kode bertanda hubung = %#v", payload.Results) }

	res = requestUji(t, handler, http.MethodGet, "/api/search?q=%22", nil)
	if res.Code != http.StatusOK { t.Fatalf("query tanda kutip sebagai teks: status = %d; body = %s", res.Code, res.Body.String()) }
}

func TestSearchEndpointDoesNotExposeContent(t *testing.T) {
	handler, server := serverUjiDenganServer(t)
	if _, err := server.db.Exec(`INSERT INTO notes(id, notebook_id, title, content, created_at, updated_at) VALUES ("note-content", "default", "Hasil", "isi rahasia", datetime("now"), datetime("now"))`); err != nil { t.Fatal(err) }
	res := requestUji(t, handler, http.MethodGet, "/api/search?q=rahasia", nil)
	if res.Code != http.StatusOK { t.Fatalf("status = %d", res.Code) }
	if body := res.Body.String(); strings.Contains(body, "isi rahasia") { t.Fatalf("isi catatan bocor pada hasil pencarian: %s", body) }
}
