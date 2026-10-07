package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBackupRestoreRoundTripPreservesSourceAndMetadata(t *testing.T) {
	handler, _ := serverUjiDenganServer(t)
	book := buatBukuUjiSumber(t, handler)

	created := requestUji(t, handler, http.MethodPost, "/api/notebooks/"+book.ID+"/notes", map[string]string{
		"title": "Catatan cadangan",
		"content": "Isi yang harus kembali",
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("buat catatan: status = %d", created.Code)
	}

	source := imporSumberUji(t, handler, book.ID)
	backup := requestUji(t, handler, http.MethodGet, "/api/notebooks/"+book.ID+"/backup", nil)
	if backup.Code != http.StatusOK {
		t.Fatalf("buat cadangan: status = %d, body = %q", backup.Code, backup.Body.String())
	}

	var document backupDocument
	if err := json.NewDecoder(backup.Body).Decode(&document); err != nil {
		t.Fatal(err)
	}
	if document.Format != "catatan-backup" || document.Version != backupFormatVersion {
		t.Fatalf("format cadangan = %#v", document)
	}
	if len(document.Notes) != 1 || len(document.Sources) != 1 {
		t.Fatalf("isi cadangan tidak lengkap: catatan=%d sumber=%d", len(document.Notes), len(document.Sources))
	}
	if document.Sources[0].ID != source.ID || document.Sources[0].Title != "Bahan" || document.Sources[0].Content != "isi sumber" || document.Sources[0].Checksum == "" {
		t.Fatalf("sumber cadangan tidak lengkap: %#v", document.Sources[0])
	}

	deleted := requestUji(t, handler, http.MethodDelete, "/api/sources/"+source.ID, nil)
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("hapus sumber sebelum pemulihan: status = %d", deleted.Code)
	}

	payload, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	restore := requestDenganBadan(t, handler, http.MethodPost, "/api/notebooks/"+book.ID+"/backup", payload)
	if restore.Code != http.StatusOK {
		t.Fatalf("pulihkan cadangan: status = %d, body = %q", restore.Code, restore.Body.String())
	}

	restored := detailSumberUji(t, handler, source.ID)
	if restored.Content != "isi sumber" || restored.Checksum != document.Sources[0].Checksum || restored.Title != "Bahan" {
		t.Fatalf("sumber setelah pemulihan tidak sesuai: %#v", restored)
	}
}

func TestRestoreRejectsCorruptBackupWithoutChangingData(t *testing.T) {
	handler := serverUji(t)
	book := buatBukuUjiSumber(t, handler)
	source := imporSumberUji(t, handler, book.ID)

	backup := requestUji(t, handler, http.MethodGet, "/api/notebooks/"+book.ID+"/backup", nil)
	if backup.Code != http.StatusOK {
		t.Fatalf("buat cadangan: status = %d", backup.Code)
	}
	var document backupDocument
	if err := json.NewDecoder(backup.Body).Decode(&document); err != nil {
		t.Fatal(err)
	}
	document.Checksum = "checksum-rusak"

	payload, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	restore := requestDenganBadan(t, handler, http.MethodPost, "/api/notebooks/"+book.ID+"/backup", payload)
	if restore.Code != http.StatusBadRequest {
		t.Fatalf("cadangan rusak: status = %d, ingin %d", restore.Code, http.StatusBadRequest)
	}

	restored := detailSumberUji(t, handler, source.ID)
	if restored.Content != "isi sumber" || restored.Checksum == "" {
		t.Fatalf("data berubah setelah cadangan rusak ditolak: %#v", restored)
	}
}

func requestDenganBadan(t *testing.T, handler http.Handler, method, path string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	return res
}
