package http

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/umam-src/catatan/internal/db"
	"github.com/umam-src/catatan/internal/ui"
)

const (
	maxNoteContent = 1 << 20
	maxRequestBody = maxNoteContent + 4096
	maxTitle       = 200
	maxDescription = 2000
)

type Server struct{ db *db.DB }

func New(d *db.DB) http.Handler {
	s := &Server{db: d}
	auth := newAuthServer(d)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/setup", auth.setup)
	mux.HandleFunc("POST /api/auth/login", auth.login)
	mux.HandleFunc("POST /api/auth/logout", auth.logout)
	mux.HandleFunc("GET /api/auth/me", auth.me)
	mux.HandleFunc("GET /api/notebooks", s.listNotebooks)
	mux.HandleFunc("POST /api/notebooks", s.createNotebook)
	mux.HandleFunc("GET /api/notebooks/{id}/notes", s.listNotes)
	mux.HandleFunc("POST /api/notebooks/{id}/notes", s.createNote)
	mux.HandleFunc("GET /api/notebooks/{id}/sources", s.listSources)
	mux.HandleFunc("POST /api/notebooks/{id}/sources", s.importSource)
	mux.HandleFunc("GET /api/notebooks/{id}/export", s.exportNotebook)
	mux.HandleFunc("GET /api/notebooks/{id}/backup", s.backupNotebook)
	mux.HandleFunc("POST /api/notebooks/{id}/backup", s.restoreNotebook)
	mux.HandleFunc("GET /api/sources/{id}", s.getSource)
	mux.HandleFunc("GET /api/sources/{id}/locations", s.listSourceLocations)
	mux.HandleFunc("POST /api/sources/{id}/locations", s.createSourceLocation)
	mux.HandleFunc("PUT /api/sources/{id}", s.updateSource)
	mux.HandleFunc("DELETE /api/sources/{id}", s.deleteSource)
	mux.HandleFunc("PUT /api/notes/{id}", s.updateNote)
	mux.HandleFunc("DELETE /api/notes/{id}", s.deleteNote)
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	assets, _ := fs.Sub(ui.Assets, "dist")
	fileServer := http.FileServer(http.FS(assets))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			if allow := allowedMethods(r.URL.Path); allow != "" && r.Method != methodFromAllow(allow) {
				if !methodAllowed(allow, r.Method) {
					w.Header().Set("Allow", allow)
					http.Error(w, "metode tidak diizinkan", http.StatusMethodNotAllowed)
					return
				}
			}
			http.NotFound(w, r)
			return
		}
		p := strings.TrimPrefix(r.URL.Path, "/")
		if p != "" && p != "." && fs.ValidPath(p) {
			if f, err := assets.Open(p); err == nil {
				_ = f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}
		}
		index, err := fs.ReadFile(assets, "index.html")
		if err != nil {
			http.Error(w, "antarmuka belum dibangun", http.StatusInternalServerError)
			return
		}
		http.ServeContent(w, r, "index.html", time.Time{}, bytes.NewReader(index))
	})
	return withHeaders(withAuth(auth, mux))
}

func (s *Server) listNotebooks(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.QueryContext(r.Context(), `SELECT id,title,description,archived,created_at,updated_at FROM notebooks WHERE archived=0 AND owner_id=? ORDER BY updated_at DESC`, userID(r))
	if err != nil { serverError(w, err); return }
	defer rows.Close()
	out := []Notebook{}
	for rows.Next() {
		var n Notebook
		var archived int
		if err := rows.Scan(&n.ID, &n.Title, &n.Description, &archived, &n.CreatedAt, &n.UpdatedAt); err != nil { serverError(w, err); return }
		n.Archived = archived != 0
		out = append(out, n)
	}
	if err := rows.Err(); err != nil { serverError(w, err); return }
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) createNotebook(w http.ResponseWriter, r *http.Request) {
	var in struct { Title string `json:"title"`; Description string `json:"description"` }
	if !decode(w, r, &in) { return }
	in.Title = strings.TrimSpace(in.Title)
	in.Description = strings.TrimSpace(in.Description)
	if err := validateLength(in.Title, 1, maxTitle, "Judul"); err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	if err := validateLength(in.Description, 0, maxDescription, "Deskripsi"); err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	now := time.Now().UTC().Format(time.RFC3339Nano)
	id := newID()
	_, err := s.db.ExecContext(r.Context(), `INSERT INTO notebooks(id,owner_id,title,description,created_at,updated_at) VALUES(?,?,?,?,?,?)`, id, userID(r), in.Title, in.Description, now, now)
	if err != nil { serverError(w, err); return }
	writeJSON(w, http.StatusCreated, Notebook{ID: id, Title: in.Title, Description: in.Description, CreatedAt: now, UpdatedAt: now})
}

func (s *Server) listNotes(w http.ResponseWriter, r *http.Request) {
	notebookID := r.PathValue("id")
	if !validID(notebookID) { http.Error(w, "ID buku tidak valid", http.StatusBadRequest); return }
	exists, err := s.notebookOwnedBy(r, notebookID)
	if err != nil { serverError(w, err); return }
	if !exists { http.NotFound(w, r); return }
	rows, err := s.db.QueryContext(r.Context(), `SELECT id,title,content,note_type,created_at,updated_at FROM notes WHERE notebook_id=? AND deleted_at IS NULL ORDER BY updated_at DESC`, notebookID)
	if err != nil { serverError(w, err); return }
	defer rows.Close()
	out := []Note{}
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.NoteType, &n.CreatedAt, &n.UpdatedAt); err != nil { serverError(w, err); return }
		out = append(out, n)
	}
	if err := rows.Err(); err != nil { serverError(w, err); return }
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) createNote(w http.ResponseWriter, r *http.Request) {
	notebookID := r.PathValue("id")
	if !validID(notebookID) { http.Error(w, "ID buku tidak valid", http.StatusBadRequest); return }
	exists, err := s.notebookOwnedBy(r, notebookID)
	if err != nil { serverError(w, err); return }
	if !exists { http.NotFound(w, r); return }
	var in struct { Title string `json:"title"`; Content string `json:"content"` }
	if !decode(w, r, &in) { return }
	in.Title = strings.TrimSpace(in.Title)
	if err := validateLength(in.Title, 1, maxTitle, "Judul"); err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	if err := validateLength(in.Content, 0, maxNoteContent, "Isi"); err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	id := newID()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = s.db.ExecContext(r.Context(), `INSERT INTO notes(id,notebook_id,title,content,created_at,updated_at) VALUES(?,?,?,?,?,?)`, id, notebookID, in.Title, in.Content, now, now)
	if err != nil { serverError(w, err); return }
	writeJSON(w, http.StatusCreated, Note{ID: id, Title: in.Title, Content: in.Content, NoteType: "human", CreatedAt: now, UpdatedAt: now})
}

func (s *Server) deleteNote(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !validID(id) { http.Error(w, "ID catatan tidak valid", http.StatusBadRequest); return }
	result, err := s.db.ExecContext(r.Context(), `UPDATE notes SET deleted_at=? WHERE id=? AND deleted_at IS NULL AND notebook_id IN (SELECT id FROM notebooks WHERE owner_id=?)`, time.Now().UTC().Format(time.RFC3339Nano), id, userID(r))
	if err != nil { serverError(w, err); return }
	if affected, _ := result.RowsAffected(); affected == 0 { http.NotFound(w, r); return }
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) updateNote(w http.ResponseWriter, r *http.Request) {
	noteID := r.PathValue("id")
	if !validID(noteID) { http.Error(w, "ID catatan tidak valid", http.StatusBadRequest); return }
	exists, err := s.noteOwnedBy(r, noteID)
	if err != nil { serverError(w, err); return }
	if !exists { http.NotFound(w, r); return }
	var in struct { Title string `json:"title"`; Content string `json:"content"` }
	if !decode(w, r, &in) { return }
	in.Title = strings.TrimSpace(in.Title)
	if err := validateLength(in.Title, 1, maxTitle, "Judul"); err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	if err := validateLength(in.Content, 1, maxNoteContent, "Isi"); err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = s.db.ExecContext(r.Context(), `UPDATE notes SET title=?,content=?,updated_at=? WHERE id=?`, in.Title, in.Content, now, noteID)
	if err != nil { serverError(w, err); return }
	writeJSON(w, http.StatusOK, map[string]string{"updated_at": now})
}

type Notebook struct { ID string `json:"id"`; Title string `json:"title"`; Description string `json:"description"`; Archived bool `json:"archived"`; CreatedAt string `json:"created_at"`; UpdatedAt string `json:"updated_at"` }
type Note struct { ID string `json:"id"`; Title string `json:"title"`; Content string `json:"content"`; NoteType string `json:"note_type"`; CreatedAt string `json:"created_at"`; UpdatedAt string `json:"updated_at"` }

func newID() string { b := make([]byte, 16); if _, err := rand.Read(b); err != nil { panic("sumber acak sistem tidak tersedia") }; return hex.EncodeToString(b) }
func validID(id string) bool { if len(id) != 32 { return false }; for _, c := range id { if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) { return false } }; return true }
func validateLength(value string, min, max int, field string) error { length := len([]rune(value)); if length < min { if min == 1 { return errors.New(field + " wajib diisi") }; return errors.New(field + " terlalu pendek") }; if length > max { return errors.New(field + " terlalu panjang") }; return nil }

func (s *Server) notebookOwnedBy(r *http.Request, id string) (bool, error) { var exists int; err := s.db.QueryRowContext(r.Context(), `SELECT 1 FROM notebooks WHERE id=? AND owner_id=? LIMIT 1`, id, userID(r)).Scan(&exists); if errors.Is(err, sql.ErrNoRows) { return false, nil }; if err != nil { return false, err }; return exists == 1, nil }
func (s *Server) noteOwnedBy(r *http.Request, id string) (bool, error) { var exists int; err := s.db.QueryRowContext(r.Context(), `SELECT 1 FROM notes WHERE id=? AND deleted_at IS NULL AND notebook_id IN (SELECT id FROM notebooks WHERE owner_id=?) LIMIT 1`, id, userID(r)).Scan(&exists); if errors.Is(err, sql.ErrNoRows) { return false, nil }; if err != nil { return false, err }; return exists == 1, nil }

func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	if r.ContentLength > maxRequestBody { http.Error(w, "Badan permintaan terlalu besar", http.StatusRequestEntityTooLarge); return false }
	body := http.MaxBytesReader(w, r.Body, maxRequestBody); defer body.Close()
	decoder := json.NewDecoder(body); decoder.DisallowUnknownFields()
	if err := decoder.Decode(v); err != nil { var maxBytesErr *http.MaxBytesError; switch { case errors.As(err, &maxBytesErr): http.Error(w, "Badan permintaan terlalu besar", http.StatusRequestEntityTooLarge); case errors.Is(err, io.EOF): http.Error(w, "Badan permintaan wajib diisi", http.StatusBadRequest); default: http.Error(w, "JSON tidak valid", http.StatusBadRequest) }; return false }
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF { http.Error(w, "Badan permintaan harus berisi satu objek JSON", http.StatusBadRequest); return false }
	return true
}

func allowedMethods(path string) string {
	switch {
	case path == "/api/auth/setup" || path == "/api/auth/login" || path == "/api/auth/logout": return http.MethodPost
	case path == "/api/auth/me": return http.MethodGet
	case path == "/api/health": return http.MethodGet
	case path == "/api/notebooks": return http.MethodGet + ", " + http.MethodPost
	case strings.HasPrefix(path, "/api/notebooks/") && strings.HasSuffix(path, "/notes"): return http.MethodGet + ", " + http.MethodPost
	case strings.HasPrefix(path, "/api/notebooks/") && strings.HasSuffix(path, "/sources"): return http.MethodGet + ", " + http.MethodPost
	case strings.HasPrefix(path, "/api/notebooks/") && strings.HasSuffix(path, "/export"): return http.MethodGet
	case strings.HasPrefix(path, "/api/notebooks/") && strings.HasSuffix(path, "/backup"): return http.MethodGet + ", " + http.MethodPost
	case strings.HasPrefix(path, "/api/sources/") && strings.HasSuffix(path, "/locations"): return http.MethodGet + ", " + http.MethodPost
	case strings.HasPrefix(path, "/api/sources/"): return http.MethodGet + ", " + http.MethodPut + ", " + http.MethodDelete
	case strings.HasPrefix(path, "/api/notes/"): return http.MethodPut + ", " + http.MethodDelete
	default: return ""
	}
}

func methodFromAllow(allow string) string { if i := strings.IndexByte(allow, ','); i >= 0 { return strings.TrimSpace(allow[:i]) }; return allow }
func methodAllowed(allow, method string) bool { for _, allowed := range strings.Split(allow, ",") { if strings.TrimSpace(allowed) == method { return true } }; return false }
func writeJSON(w http.ResponseWriter, status int, v any) { w.Header().Set("Content-Type", "application/json; charset=utf-8"); w.WriteHeader(status); _ = json.NewEncoder(w).Encode(v) }
func serverError(w http.ResponseWriter, _ error) { http.Error(w, "kesalahan internal", http.StatusInternalServerError) }
func withHeaders(next http.Handler) http.Handler { return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Header().Set("X-Content-Type-Options", "nosniff"); w.Header().Set("Referrer-Policy", "no-referrer"); next.ServeHTTP(w, r) }) }
