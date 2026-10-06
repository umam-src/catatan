package http

import (
	"net/http"
	"strings"
	"testing"
)

func TestNoteValidationRejectsOversizedTitleAndContent(t *testing.T) {
	handler := serverUji(t)

	res := requestUji(t, handler, http.MethodPost, "/api/notebooks", map[string]string{"title": "Buku Uji"})
	if res.Code != http.StatusCreated {
		t.Fatalf("buat buku: status = %d", res.Code)
	}
	var notebook Notebook
	if err := decodeJSON(res, &notebook); err != nil {
		t.Fatal(err)
	}

	res = requestUji(t, handler, http.MethodPost, "/api/notebooks/"+notebook.ID+"/notes", map[string]string{
		"title":   strings.Repeat("a", maxTitle+1),
		"content": "Isi",
	})
	if res.Code != http.StatusBadRequest {
		t.Fatalf("judul terlalu panjang: status = %d, ingin %d", res.Code, http.StatusBadRequest)
	}

	res = requestUji(t, handler, http.MethodPost, "/api/notebooks/"+notebook.ID+"/notes", map[string]string{
		"title":   "Catatan Uji",
		"content": strings.Repeat("a", maxNoteContent+1),
	})
	if res.Code != http.StatusBadRequest {
		t.Fatalf("isi terlalu panjang: status = %d, ingin %d", res.Code, http.StatusBadRequest)
	}
}

func TestUpdateNoteRejectsInvalidContent(t *testing.T) {
	handler := serverUji(t)

	res := requestUji(t, handler, http.MethodPost, "/api/notebooks", map[string]string{"title": "Buku Uji"})
	if res.Code != http.StatusCreated {
		t.Fatalf("buat buku: status = %d", res.Code)
	}
	var notebook Notebook
	if err := decodeJSON(res, &notebook); err != nil {
		t.Fatal(err)
	}

	res = requestUji(t, handler, http.MethodPost, "/api/notebooks/"+notebook.ID+"/notes", map[string]string{
		"title":   "Catatan Uji",
		"content": "Isi awal",
	})
	if res.Code != http.StatusCreated {
		t.Fatalf("buat catatan: status = %d", res.Code)
	}
	var note Note
	if err := decodeJSON(res, &note); err != nil {
		t.Fatal(err)
	}

	res = requestUji(t, handler, http.MethodPut, "/api/notes/"+note.ID, map[string]string{
		"title":   "Catatan Diperbarui",
		"content": "",
	})
	if res.Code != http.StatusBadRequest {
		t.Fatalf("isi kosong saat ubah: status = %d, ingin %d", res.Code, http.StatusBadRequest)
	}

	res = requestUji(t, handler, http.MethodPut, "/api/notes/"+note.ID, map[string]string{
		"title":   "Catatan Diperbarui",
		"content": strings.Repeat("a", maxNoteContent+1),
	})
	if res.Code != http.StatusBadRequest {
		t.Fatalf("isi terlalu panjang saat ubah: status = %d, ingin %d", res.Code, http.StatusBadRequest)
	}
}
