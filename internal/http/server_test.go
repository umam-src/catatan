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
	"time"
	"testing"

	"github.com/umam-src/buku-catatan/internal/db"
)

func serverUji(t *testing.T) http.Handler {
	t.Helper()
	handler, _ := serverUjiDenganServer(t)
	return handler
}

func serverUjiDenganServer(t *testing.T) (http.Handler, *Server) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "catatan.db")
	d, err := db.Open(context.Background(), path)
	if err != nil { t.Fatal(err) }
	t.Cleanup(func() { _ = d.Close() })

	server := &Server{db: d}
	handler := New(d)
	setup := requestUji(t, handler, http.MethodPost, "/api/auth/setup", map[string]string{
		"username": "pengguna",
		"email": "pengguna@lokal.invalid",
		"display_name": "Pengguna Uji",
		"password": "kata-sandi-uji-aman",
	})
	if setup.Code != http.StatusCreated { t.Fatalf("setup pengguna: status = %d, ingin %d", setup.Code, http.StatusCreated) }
	cookies := setup.Result().Cookies()
	if len(cookies) != 1 { t.Fatalf("cookie session = %d, ingin 1", len(cookies)) }
	sessionCookie := cookies[0]

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.Clone(r.Context())
		r.AddCookie(sessionCookie)
		handler.ServeHTTP(w, r)
	}), server
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

func TestNoteFlowWorksWithoutNetwork(t *testing.T) {
	handler := serverUji(t)
	createdBook := requestUji(t, handler, http.MethodPost, "/api/notebooks", map[string]string{"title": "Buku Luring"})
	if createdBook.Code != http.StatusCreated { t.Fatalf("buat buku: status = %d, ingin %d", createdBook.Code, http.StatusCreated) }
	var notebook Notebook
	if err := json.NewDecoder(createdBook.Body).Decode(&notebook); err != nil { t.Fatal(err) }
	createdNote := requestUji(t, handler, http.MethodPost, "/api/notebooks/"+notebook.ID+"/notes", map[string]string{"title": "Catatan Luring", "content": "Isi sebelum luring"})
	if createdNote.Code != http.StatusCreated { t.Fatalf("buat catatan: status = %d", createdNote.Code) }
	var note Note
	if err := json.NewDecoder(createdNote.Body).Decode(&note); err != nil { t.Fatal(err) }
	updated := requestUji(t, handler, http.MethodPut, "/api/notes/"+note.ID, map[string]string{"title": "Catatan Luring Diperbarui", "content": "Isi tetap tersimpan tanpa jaringan"})
	if updated.Code != http.StatusOK { t.Fatalf("ubah catatan: status = %d", updated.Code) }
	listed := requestUji(t, handler, http.MethodGet, "/api/notebooks/"+notebook.ID+"/notes", nil)
	if listed.Code != http.StatusOK { t.Fatalf("buka ulang daftar: status = %d", listed.Code) }
	var notes []Note
	if err := json.NewDecoder(listed.Body).Decode(&notes); err != nil { t.Fatal(err) }
	if len(notes) != 1 || notes[0].Content != "Isi tetap tersimpan tanpa jaringan" { t.Fatalf("isi catatan setelah buka ulang = %#v", notes) }
	deleted := requestUji(t, handler, http.MethodDelete, "/api/notes/"+note.ID, nil)
	if deleted.Code != http.StatusNoContent { t.Fatalf("hapus catatan: status = %d", deleted.Code) }
}

func TestCreateNotebookRejectsInvalidInput(t *testing.T) {
	handler := serverUji(t)
	tests := []struct { name string; body any }{{"kosong", map[string]string{"title": "   "}}, {"tidak dikenal", map[string]string{"title": "Uji", "lain": "rahasia"}}}
	for _, tt := range tests { t.Run(tt.name, func(t *testing.T) { res := requestUji(t, handler, http.MethodPost, "/api/notebooks", tt.body); if res.Code != http.StatusBadRequest { t.Fatalf("status = %d, ingin %d", res.Code, http.StatusBadRequest) } }) }
}

func TestCreateNotebookRejectsOversizedBody(t *testing.T) {
	handler := serverUji(t)
	body := `{"title":"` + strings.Repeat("a", maxRequestBody) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/notebooks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusRequestEntityTooLarge { t.Fatalf("status = %d, ingin %d", res.Code, http.StatusRequestEntityTooLarge) }
}

func TestNoteEndpointsValidateIDsAndResources(t *testing.T) {
	handler := serverUji(t)
	res := requestUji(t, handler, http.MethodGet, "/api/notebooks/tidak-valid/notes", nil)
	if res.Code != http.StatusBadRequest { t.Fatalf("ID buku tidak valid: status = %d, ingin %d", res.Code, http.StatusBadRequest) }
	res = requestUji(t, handler, http.MethodGet, "/api/notebooks/00000000000000000000000000000000/notes", nil)
	if res.Code != http.StatusNotFound { t.Fatalf("buku tidak ditemukan: status = %d, ingin %d", res.Code, http.StatusNotFound) }
	res = requestUji(t, handler, http.MethodPut, "/api/notes/tidak-valid", map[string]string{"title": "Uji", "content": "Isi"})
	if res.Code != http.StatusBadRequest { t.Fatalf("PUT ID tidak valid: status = %d, ingin %d", res.Code, http.StatusBadRequest) }
	res = requestUji(t, handler, http.MethodPut, "/api/notes/00000000000000000000000000000000", map[string]string{"title": "Uji", "content": "Isi"})
	if res.Code != http.StatusNotFound { t.Fatalf("PUT catatan tidak ditemukan: status = %d, ingin %d", res.Code, http.StatusNotFound) }
}

func TestCreateAndUpdateNoteValidation(t *testing.T) {
	handler := serverUji(t)
	res := requestUji(t, handler, http.MethodPost, "/api/notebooks", map[string]string{"title": "Buku Uji"})
	if res.Code != http.StatusCreated { t.Fatalf("buat buku: status = %d", res.Code) }
	var notebook Notebook
	if err := json.NewDecoder(res.Body).Decode(&notebook); err != nil { t.Fatal(err) }
	res = requestUji(t, handler, http.MethodPost, "/api/notebooks/"+notebook.ID+"/notes", map[string]string{"title": "", "content": "Isi"})
	if res.Code != http.StatusBadRequest { t.Fatalf("judul catatan kosong: status = %d, ingin %d", res.Code, http.StatusBadRequest) }
	res = requestUji(t, handler, http.MethodPost, "/api/notebooks/"+notebook.ID+"/notes", map[string]string{"title": "Catatan Uji", "content": "Isi Uji"})
	if res.Code != http.StatusCreated { t.Fatalf("buat catatan: status = %d", res.Code) }
	var note Note
	if err := json.NewDecoder(res.Body).Decode(&note); err != nil { t.Fatal(err) }
	res = requestUji(t, handler, http.MethodPut, "/api/notes/"+note.ID, map[string]string{"title": "Catatan Diperbarui", "content": "Isi Diperbarui"})
	if res.Code != http.StatusOK { t.Fatalf("ubah catatan: status = %d", res.Code) }
}

func TestDeleteNote(t *testing.T) {
	handler := serverUji(t)
	res := requestUji(t, handler, http.MethodPost, "/api/notebooks", map[string]string{"title": "Buku Uji"})
	if res.Code != http.StatusCreated { t.Fatalf("buat buku: status = %d", res.Code) }
	var notebook Notebook
	if err := json.NewDecoder(res.Body).Decode(&notebook); err != nil { t.Fatal(err) }
	res = requestUji(t, handler, http.MethodPost, "/api/notebooks/"+notebook.ID+"/notes", map[string]string{"title": "Catatan", "content": ""})
	if res.Code != http.StatusCreated { t.Fatalf("buat catatan kosong: status = %d", res.Code) }
	var note Note
	if err := json.NewDecoder(res.Body).Decode(&note); err != nil { t.Fatal(err) }
	res = requestUji(t, handler, http.MethodDelete, "/api/notes/"+note.ID, nil)
	if res.Code != http.StatusNoContent { t.Fatalf("hapus: status = %d", res.Code) }
	res = requestUji(t, handler, http.MethodGet, "/api/notebooks/"+notebook.ID+"/notes", nil)
	if res.Code != http.StatusOK { t.Fatalf("daftar setelah hapus: status = %d", res.Code) }
	var notes []Note
	if err := json.NewDecoder(res.Body).Decode(&notes); err != nil { t.Fatal(err) }
	if len(notes) != 0 { t.Fatalf("catatan terhapus masih muncul: %d", len(notes)) }
}

func TestMethodsAreRestricted(t *testing.T) {
	handler := serverUji(t)
	res := requestUji(t, handler, http.MethodPatch, "/api/health", nil)
	if res.Code != http.StatusMethodNotAllowed { t.Fatalf("status = %d, ingin %d", res.Code, http.StatusMethodNotAllowed) }
}

func TestAuthRequiresSession(t *testing.T) {
	path := filepath.Join(t.TempDir(), "catatan.db")
	d, err := db.Open(context.Background(), path)
	if err != nil { t.Fatal(err) }
	t.Cleanup(func() { _ = d.Close() })
	handler := New(d)
	res := requestUji(t, handler, http.MethodGet, "/api/notebooks", nil)
	if res.Code != http.StatusUnauthorized { t.Fatalf("tanpa session: status = %d, ingin %d", res.Code, http.StatusUnauthorized) }
}

func TestAuthLoginLogoutAndSessionCookie(t *testing.T) {
	path := filepath.Join(t.TempDir(), "catatan.db")
	d, err := db.Open(context.Background(), path)
	if err != nil { t.Fatal(err) }
	t.Cleanup(func() { _ = d.Close() })
	handler := New(d)
	setup := requestUji(t, handler, http.MethodPost, "/api/auth/setup", map[string]string{"username": "pengguna", "password": "kata-sandi-uji-aman"})
	if setup.Code != http.StatusCreated { t.Fatalf("setup: status = %d", setup.Code) }
	cookies := setup.Result().Cookies()
	if len(cookies) != 1 || !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteStrictMode { t.Fatal("cookie session tidak memiliki atribut keamanan yang diharapkan") }
	logout := requestDenganCookie(t, handler, http.MethodPost, "/api/auth/logout", nil, cookies[0])
	if logout.Code != http.StatusNoContent { t.Fatalf("logout: status = %d", logout.Code) }
	login := requestUji(t, handler, http.MethodPost, "/api/auth/login", map[string]string{"username": "pengguna", "password": "kata-sandi-uji-aman"})
	if login.Code != http.StatusOK { t.Fatalf("login: status = %d", login.Code) }
	if len(login.Result().Cookies()) != 1 { t.Fatal("login tidak menerbitkan session baru") }
	bad := requestUji(t, handler, http.MethodPost, "/api/auth/login", map[string]string{"username": "pengguna", "password": "salah"})
	if bad.Code != http.StatusUnauthorized { t.Fatalf("login gagal: status = %d", bad.Code) }
}

func requestDenganCookie(t *testing.T, handler http.Handler, method, path string, body any, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != nil { data, err := json.Marshal(body); if err != nil { t.Fatal(err) }; reader = bytes.NewReader(data) }
	req := httptest.NewRequest(method, path, reader)
	if body != nil { req.Header.Set("Content-Type", "application/json") }
	req.AddCookie(cookie)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	return res
}

func TestSessionExpirationAndUserIsolation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "catatan.db")
	d, err := db.Open(context.Background(), path)
	if err != nil { t.Fatal(err) }
	t.Cleanup(func() { _ = d.Close() })
	handler := New(d)
	setup := requestUji(t, handler, http.MethodPost, "/api/auth/setup", map[string]string{"username": "pengguna1", "password": "kata-sandi-uji-aman"})
	if setup.Code != http.StatusCreated { t.Fatalf("setup: status = %d", setup.Code) }
	cookie1 := setup.Result().Cookies()[0]
	res := requestDenganCookie(t, handler, http.MethodPost, "/api/notebooks", map[string]string{"title": "Buku pengguna 1"}, cookie1)
	if res.Code != http.StatusCreated { t.Fatalf("buat buku: status = %d", res.Code) }
	hash, err := hashPassword("kata-sandi-pengguna-2")
	if err != nil { t.Fatal(err) }
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := d.Exec(`INSERT INTO users(id,username,email,display_name,created_at,updated_at) VALUES('user2','pengguna2','','Pengguna 2',?,?)`, now, now); err != nil { t.Fatal(err) }
	if _, err := d.Exec(`INSERT INTO auth_credentials(user_id,password_hash,updated_at) VALUES('user2',?,?)`, hash, now); err != nil { t.Fatal(err) }
	login := requestUji(t, handler, http.MethodPost, "/api/auth/login", map[string]string{"username": "pengguna2", "password": "kata-sandi-pengguna-2"})
	if login.Code != http.StatusOK { t.Fatalf("login pengguna 2: status = %d", login.Code) }
	cookie2 := login.Result().Cookies()[0]
	res = requestDenganCookie(t, handler, http.MethodGet, "/api/notebooks", nil, cookie2)
	if res.Code != http.StatusOK { t.Fatalf("daftar pengguna 2: status = %d", res.Code) }
	var notebooks []Notebook
	if err := json.NewDecoder(res.Body).Decode(&notebooks); err != nil { t.Fatal(err) }
	if len(notebooks) != 0 { t.Fatalf("buku pengguna 1 bocor ke pengguna 2: %#v", notebooks) }
	res = requestDenganCookie(t, handler, http.MethodGet, "/api/notebooks", nil, cookie1)
	if res.Code != http.StatusOK { t.Fatalf("daftar pengguna 1: status = %d", res.Code) }
	if err := json.NewDecoder(res.Body).Decode(&notebooks); err != nil { t.Fatal(err) }
	if len(notebooks) != 1 { t.Fatalf("buku pengguna 1 hilang: %#v", notebooks) }
	if _, err := d.Exec(`UPDATE sessions SET expires_at=?`, time.Now().UTC().Add(-time.Hour).Format(time.RFC3339Nano)); err != nil { t.Fatal(err) }
	res = requestDenganCookie(t, handler, http.MethodGet, "/api/notebooks", nil, cookie1)
	if res.Code != http.StatusUnauthorized { t.Fatalf("session kedaluwarsa: status = %d", res.Code) }
}
