package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/umam-src/buku-catatan/internal/db"
)

func serverUji(t *testing.T) http.Handler {
	t.Helper()
	path := filepath.Join(t.TempDir(), "catatan.db")
	d, err := db.Open(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = d.Close() })
	return New(d)
}

func requestUji(t *testing.T, handler http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(data)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	return res
}

func TestCreateNotebookRejectsInvalidInput(t *testing.T) {
	handler := serverUji(t)

	tests := []struct {
		name string
		body any
	}{
		{"kosong", map[string]string{"title": "   "}},
		{"tidak dikenal", map[string]string{"title": "Uji", "lain": "rahasia"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := requestUji(t, handler, http.MethodPost, "/api/notebooks", tt.body)
			if res.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, ingin %d", res.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestCreateNotebookRejectsOversizedBody(t *testing.T) {
	handler := serverUji(t)
	body := `{"title":"` + strings.Repeat("a", maxRequestBody) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/notebooks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, ingin %d", res.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestNoteEndpointsValidateIDsAndResources(t *testing.T) {
	handler := serverUji(t)

	res := requestUji(t, handler, http.MethodGet, "/api/notebooks/tidak-valid/notes", nil)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("ID buku tidak valid: status = %d, ingin %d", res.Code, http.StatusBadRequest)
	}

	res = requestUji(t, handler, http.MethodGet, "/api/notebooks/00000000000000000000000000000000/notes", nil)
	if res.Code != http.StatusNotFound {
		t.Fatalf("buku tidak ditemukan: status = %d, ingin %d", res.Code, http.StatusNotFound)
	}

	res = requestUji(t, handler, http.MethodPut, "/api/notes/tidak-valid", map[string]string{"title": "Uji", "content": "Isi"})
	if res.Code != http.StatusBadRequest {
		t.Fatalf("PUT ID tidak valid: status = %d, ingin %d", res.Code, http.StatusBadRequest)
	}

	res = requestUji(t, handler, http.MethodPut, "/api/notes/00000000000000000000000000000000", map[string]string{"title": "Uji", "content": "Isi"})
	if res.Code != http.StatusNotFound {
		t.Fatalf("PUT catatan tidak ditemukan: status = %d, ingin %d", res.Code, http.StatusNotFound)
	}
}

func TestCreateAndUpdateNoteValidation(t *testing.T) {
	handler := serverUji(t)

	res := requestUji(t, handler, http.MethodPost, "/api/notebooks", map[string]string{"title": "Buku Uji"})
	if res.Code != http.StatusCreated {
		t.Fatalf("buat buku: status = %d", res.Code)
	}
	var notebook Notebook
	if err := json.NewDecoder(res.Body).Decode(&notebook); err != nil {
		t.Fatal(err)
	}

	res = requestUji(t, handler, http.MethodPost, "/api/notebooks/"+notebook.ID+"/notes", map[string]string{"title": "", "content": "Isi"})
	if res.Code != http.StatusBadRequest {
		t.Fatalf("judul catatan kosong: status = %d, ingin %d", res.Code, http.StatusBadRequest)
	}

	res = requestUji(t, handler, http.MethodPost, "/api/notebooks/"+notebook.ID+"/notes", map[string]string{"title": "Catatan Uji", "content": "Isi Uji"})
	if res.Code != http.StatusCreated {
		t.Fatalf("buat catatan: status = %d", res.Code)
	}
	var note Note
	if err := json.NewDecoder(res.Body).Decode(&note); err != nil {
		t.Fatal(err)
	}

	res = requestUji(t, handler, http.MethodPut, "/api/notes/"+note.ID, map[string]string{"title": "Catatan Diperbarui", "content": "Isi Diperbarui"})
	if res.Code != http.StatusOK {
		t.Fatalf("ubah catatan: status = %d", res.Code)
	}
}

func TestMethodsAreRestricted(t *testing.T) {
	handler := serverUji(t)
	res := requestUji(t, handler, http.MethodPatch, "/api/health", nil)
	if res.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, ingin %d", res.Code, http.StatusMethodNotAllowed)
	}
}
