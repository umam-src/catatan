package http

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

const maxSourceContent = 10 << 20

type Source struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Kind      string `json:"kind"`
	Content   string `json:"content,omitempty"`
	Locator   string `json:"locator"`
	Checksum  string `json:"checksum"`
	Metadata  string `json:"metadata_json"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (s *Server) listSources(w http.ResponseWriter, r *http.Request) {
	notebookID := r.PathValue("id")
	if !validID(notebookID) {
		http.Error(w, "ID buku tidak valid", http.StatusBadRequest)
		return
	}
	owned, err := s.notebookOwnedBy(r, notebookID)
	if err != nil {
		serverError(w, err)
		return
	}
	if !owned {
		http.NotFound(w, r)
		return
	}

	rows, err := s.db.QueryContext(r.Context(), `SELECT id,title,kind,locator,checksum,metadata_json,created_at,updated_at FROM sources WHERE notebook_id=? ORDER BY updated_at DESC`, notebookID)
	if err != nil {
		serverError(w, err)
		return
	}
	defer rows.Close()

	out := []Source{}
	for rows.Next() {
		var source Source
		if err := rows.Scan(&source.ID, &source.Title, &source.Kind, &source.Locator, &source.Checksum, &source.Metadata, &source.CreatedAt, &source.UpdatedAt); err != nil {
			serverError(w, err)
			return
		}
		out = append(out, source)
	}
	if err := rows.Err(); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) getSource(w http.ResponseWriter, r *http.Request) {
	sourceID := r.PathValue("id")
	if !validID(sourceID) {
		http.Error(w, "ID sumber tidak valid", http.StatusBadRequest)
		return
	}

	var source Source
	err := s.db.QueryRowContext(r.Context(), `SELECT id,title,kind,content,locator,checksum,metadata_json,created_at,updated_at FROM sources WHERE id=? AND notebook_id IN (SELECT id FROM notebooks WHERE owner_id=?)`, sourceID, userID(r)).Scan(
		&source.ID, &source.Title, &source.Kind, &source.Content, &source.Locator, &source.Checksum, &source.Metadata, &source.CreatedAt, &source.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, source)
}

func (s *Server) importSource(w http.ResponseWriter, r *http.Request) {
	notebookID := r.PathValue("id")
	if !validID(notebookID) {
		http.Error(w, "ID buku tidak valid", http.StatusBadRequest)
		return
	}
	owned, err := s.notebookOwnedBy(r, notebookID)
	if err != nil {
		serverError(w, err)
		return
	}
	if !owned {
		http.NotFound(w, r)
		return
	}
	if r.ContentLength > maxSourceContent+1<<20 {
		http.Error(w, "Berkas sumber terlalu besar", http.StatusRequestEntityTooLarge)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxSourceContent+1<<20)
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		http.Error(w, "Formulir impor tidak valid", http.StatusBadRequest)
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	files := r.MultipartForm.File["file"]
	if len(files) != 1 {
		http.Error(w, "Berkas sumber wajib dipilih", http.StatusBadRequest)
		return
	}
	fileHeader := files[0]
	file, err := fileHeader.Open()
	if err != nil {
		http.Error(w, "Berkas sumber tidak dapat dibuka", http.StatusBadRequest)
		return
	}
	defer file.Close()

	content, err := io.ReadAll(io.LimitReader(file, maxSourceContent+1))
	if err != nil {
		serverError(w, err)
		return
	}
	if len(content) > maxSourceContent {
		http.Error(w, "Berkas sumber terlalu besar", http.StatusRequestEntityTooLarge)
		return
	}
	if !utf8.Valid(content) {
		http.Error(w, "Berkas sumber harus menggunakan UTF-8", http.StatusBadRequest)
		return
	}

	filename := filepath.Base(strings.TrimSpace(fileHeader.Filename))
	if filename == "." || filename == "" {
		http.Error(w, "Nama berkas sumber tidak valid", http.StatusBadRequest)
		return
	}
	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		title = filename
	}
	if err := validateLength(title, 1, maxTitle, "Judul sumber"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	digest := sha256.Sum256(content)
	checksum := hex.EncodeToString(digest[:])
	metadata, err := json.Marshal(map[string]any{
		"filename":   filename,
		"media_type": "text/plain; charset=utf-8",
		"size":       len(content),
	})
	if err != nil {
		serverError(w, err)
		return
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	source := Source{ID: newID(), Title: title, Kind: "text", Locator: "", Checksum: checksum, Metadata: string(metadata), CreatedAt: now, UpdatedAt: now}
	_, err = s.db.ExecContext(r.Context(), `INSERT INTO sources(id,notebook_id,title,kind,content,locator,checksum,metadata_json,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?)`, source.ID, notebookID, source.Title, source.Kind, string(content), source.Locator, source.Checksum, source.Metadata, now, now)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, source)
}
