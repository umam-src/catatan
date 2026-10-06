package http

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

func permintaanSumber(t *testing.T, handler http.Handler, notebookID, filename, title, content string) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	file, err := writer.CreateFormFile("file", filename)
	if err != nil { t.Fatal(err) }
	if _, err := file.Write([]byte(content)); err != nil { t.Fatal(err) }
	if title != "" {
		if err := writer.WriteField("title", title); err != nil { t.Fatal(err) }
	}
	if err := writer.Close(); err != nil { t.Fatal(err) }
	req := httptest.NewRequest(http.MethodPost, "/api/notebooks/"+notebookID+"/sources", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	return res
}

func buatBukuUjiSumber(t *testing.T, handler http.Handler) Notebook {
	t.Helper()
	res := requestUji(t, handler, http.MethodPost, "/api/notebooks", map[string]string{"title": "Buku Sumber"})
	if res.Code != http.StatusCreated { t.Fatalf("buat buku: status = %d", res.Code) }
	var notebook Notebook
	if err := json.NewDecoder(res.Body).Decode(&notebook); err != nil { t.Fatal(err) }
	return notebook
}

func imporSumberUji(t *testing.T, handler http.Handler, notebookID string) Source {
	t.Helper()
	res := permintaanSumber(t, handler, notebookID, "bahan.txt", "Bahan", "isi sumber")
	if res.Code != http.StatusCreated { t.Fatalf("impor sumber: status = %d, body = %q", res.Code, res.Body.String()) }
	var source Source
	if err := json.NewDecoder(res.Body).Decode(&source); err != nil { t.Fatal(err) }
	return source
}

func TestImportSourceStoresContentAndChecksum(t *testing.T) {
	handler := serverUji(t)
	notebook := buatBukuUjiSumber(t, handler)
	content := "Judul lokal\nIsi sumber tanpa jaringan."
	res := permintaanSumber(t, handler, notebook.ID, "bahan.txt", "Bahan lokal", content)
	if res.Code != http.StatusCreated { t.Fatalf("impor sumber: status = %d, body = %q", res.Code, res.Body.String()) }
	var source Source
	if err := json.NewDecoder(res.Body).Decode(&source); err != nil { t.Fatal(err) }
	wantDigest := sha256.Sum256([]byte(content))
	if source.Title != "Bahan lokal" || source.Checksum != hex.EncodeToString(wantDigest[:]) { t.Fatalf("sumber = %#v", source) }
	listed := requestUji(t, handler, http.MethodGet, "/api/notebooks/"+notebook.ID+"/sources", nil)
	if listed.Code != http.StatusOK { t.Fatalf("daftar sumber: status = %d", listed.Code) }
	var sources []Source
	if err := json.NewDecoder(listed.Body).Decode(&sources); err != nil { t.Fatal(err) }
	if len(sources) != 1 || sources[0].ID != source.ID { t.Fatalf("daftar sumber = %#v", sources) }
	detail := requestUji(t, handler, http.MethodGet, "/api/sources/"+source.ID, nil)
	if detail.Code != http.StatusOK { t.Fatalf("detail sumber: status = %d", detail.Code) }
	var loaded Source
	if err := json.NewDecoder(detail.Body).Decode(&loaded); err != nil { t.Fatal(err) }
	if loaded.Content != content || loaded.Checksum != source.Checksum { t.Fatalf("detail sumber = %#v", loaded) }
}

func TestUpdateSourceChangesOnlyMetadata(t *testing.T) {
	handler := serverUji(t)
	notebook := buatBukuUjiSumber(t, handler)
	source := imporSumberUji(t, handler, notebook.ID)
	res := requestUji(t, handler, http.MethodPut, "/api/sources/"+source.ID, map[string]string{"title": "Judul baru"})
	if res.Code != http.StatusOK { t.Fatalf("ubah sumber: status = %d, body = %q", res.Code, res.Body.String()) }
	var updated Source
	if err := json.NewDecoder(res.Body).Decode(&updated); err != nil { t.Fatal(err) }
	if updated.Title != "Judul baru" || updated.Content != source.Content || updated.Checksum != source.Checksum || updated.CreatedAt != source.CreatedAt { t.Fatalf("sumber berubah tidak semestinya: %#v", updated) }
	if updated.UpdatedAt == source.UpdatedAt { t.Fatalf("updated_at tidak berubah: %q", updated.UpdatedAt) }
}

func TestUpdateSourceValidatesTitleAndMissingSource(t *testing.T) {
	handler := serverUji(t)
	notebook := buatBukuUjiSumber(t, handler)
	source := imporSumberUji(t, handler, notebook.ID)
	res := requestUji(t, handler, http.MethodPut, "/api/sources/"+source.ID, map[string]string{"title": "   "})
	if res.Code != http.StatusBadRequest { t.Fatalf("judul kosong: status = %d", res.Code) }
	tooLong := bytes.Repeat([]byte("x"), maxTitle+1)
	res = requestUji(t, handler, http.MethodPut, "/api/sources/"+source.ID, map[string]string{"title": string(tooLong)})
	if res.Code != http.StatusBadRequest { t.Fatalf("judul terlalu panjang: status = %d", res.Code) }
	res = requestUji(t, handler, http.MethodPut, "/api/sources/00000000000000000000000000000000", map[string]string{"title": "Tidak ada"})
	if res.Code != http.StatusNotFound { t.Fatalf("sumber tidak ditemukan: status = %d", res.Code) }
}

func TestDeleteSourceRemovesSourceAndRejectsMissing(t *testing.T) {
	handler := serverUji(t)
	notebook := buatBukuUjiSumber(t, handler)
	source := imporSumberUji(t, handler, notebook.ID)
	res := requestUji(t, handler, http.MethodDelete, "/api/sources/"+source.ID, nil)
	if res.Code != http.StatusNoContent { t.Fatalf("hapus sumber: status = %d", res.Code) }
	res = requestUji(t, handler, http.MethodGet, "/api/sources/"+source.ID, nil)
	if res.Code != http.StatusNotFound { t.Fatalf("detail setelah hapus: status = %d", res.Code) }
	res = requestUji(t, handler, http.MethodDelete, "/api/sources/"+source.ID, nil)
	if res.Code != http.StatusNotFound { t.Fatalf("hapus sumber yang sudah tidak ada: status = %d", res.Code) }
}

func TestImportSourceRejectsInvalidContentAndSize(t *testing.T) {
	handler := serverUji(t)
	notebook := buatBukuUjiSumber(t, handler)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	file, err := writer.CreateFormFile("file", "biner.txt")
	if err != nil { t.Fatal(err) }
	if _, err := file.Write([]byte{0xff, 0xfe, 0xfd}); err != nil { t.Fatal(err) }
	if err := writer.Close(); err != nil { t.Fatal(err) }
	req := httptest.NewRequest(http.MethodPost, "/api/notebooks/"+notebook.ID+"/sources", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest { t.Fatalf("UTF-8 tidak valid: status = %d", res.Code) }
	res = permintaanSumber(t, handler, "tidak-valid", "bahan.txt", "", "isi")
	if res.Code != http.StatusBadRequest { t.Fatalf("ID buku tidak valid: status = %d", res.Code) }
	tooLarge := bytes.Repeat([]byte("x"), maxSourceContent+1)
	var oversized bytes.Buffer
	overWriter := multipart.NewWriter(&oversized)
	overFile, err := overWriter.CreateFormFile("file", "besar.txt")
	if err != nil { t.Fatal(err) }
	if _, err := overFile.Write(tooLarge); err != nil { t.Fatal(err) }
	if err := overWriter.Close(); err != nil { t.Fatal(err) }
	req = httptest.NewRequest(http.MethodPost, "/api/notebooks/"+notebook.ID+"/sources", &oversized)
	req.Header.Set("Content-Type", overWriter.FormDataContentType())
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusRequestEntityTooLarge { t.Fatalf("sumber terlalu besar: status = %d", res.Code) }
}

func TestSourceDetailDoesNotCrossNotebookOwnership(t *testing.T) {
	handler := serverUji(t)
	notebook := buatBukuUjiSumber(t, handler)
	source := imporSumberUji(t, handler, notebook.ID)
	res := requestUji(t, handler, http.MethodGet, "/api/sources/00000000000000000000000000000000", nil)
	if res.Code != http.StatusNotFound { t.Fatalf("sumber tidak ditemukan: status = %d", res.Code) }
	_ = source
}
