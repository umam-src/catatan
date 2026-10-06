package http

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	backupFormatVersion = 1
	maxBackupBody       = 64 << 20
)

type backupDocument struct {
	Format   string         `json:"format"`
	Version  int            `json:"version"`
	Checksum string         `json:"checksum"`
	Notebook ExportNotebook `json:"notebook"`
	Notes    []ExportNote   `json:"notes"`
	Sources  []ExportSource `json:"sources"`
}

var errInvalidBackup = errors.New("cadangan tidak valid")

func (s *Server) backupNotebook(w http.ResponseWriter, r *http.Request) {
	notebookID := r.PathValue("id")
	if !validNotebookID(notebookID) {
		http.Error(w, "ID buku tidak valid", http.StatusBadRequest)
		return
	}
	doc, err := s.buildBackup(r, notebookID)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if errors.Is(err, errInvalidSourceIntegrity) {
		http.Error(w, "Integritas sumber tidak dapat diverifikasi", http.StatusUnprocessableEntity)
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}
	payload, err := json.Marshal(doc)
	if err != nil {
		serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="catatan-backup-`+notebookID+`.json"`)
	_, _ = w.Write(append(payload, '\n'))
}

func (s *Server) restoreNotebook(w http.ResponseWriter, r *http.Request) {
	notebookID := r.PathValue("id")
	if !validNotebookID(notebookID) {
		http.Error(w, "ID buku tidak valid", http.StatusBadRequest)
		return
	}
	exists, err := s.notebookOwnedBy(r, notebookID)
	if err != nil {
		serverError(w, err)
		return
	}
	if !exists {
		http.NotFound(w, r)
		return
	}

	body := http.MaxBytesReader(w, r.Body, maxBackupBody)
	defer body.Close()
	var doc backupDocument
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&doc); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "Badan cadangan terlalu besar", http.StatusRequestEntityTooLarge)
		} else {
			http.Error(w, "Cadangan tidak valid", http.StatusBadRequest)
		}
		return
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		http.Error(w, "Cadangan harus berisi satu objek JSON", http.StatusBadRequest)
		return
	}
	if err := validateBackup(doc, notebookID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validateBackupSources(doc); err != nil {
		if errors.Is(err, errInvalidSourceIntegrity) {
			http.Error(w, "Integritas sumber pada cadangan tidak valid", http.StatusUnprocessableEntity)
			return
		}
		http.Error(w, "Cadangan tidak valid", http.StatusBadRequest)
		return
	}
	if err := s.replaceNotebook(r, doc); err != nil {
		if errors.Is(err, errInvalidBackup) {
			http.Error(w, "Cadangan tidak dapat dipulihkan", http.StatusBadRequest)
			return
		}
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "dipulihkan"})
}

func validateBackup(doc backupDocument, notebookID string) error {
	if doc.Format != "catatan-backup" || doc.Version != backupFormatVersion {
		return errors.New("format cadangan tidak didukung")
	}
	if doc.Notebook.ID != notebookID || !validNotebookID(doc.Notebook.ID) {
		return errors.New("buku pada cadangan tidak cocok")
	}
	if err := validateLength(doc.Notebook.Title, 1, maxTitle, "Judul"); err != nil {
		return err
	}
	if err := validateLength(doc.Notebook.Description, 0, maxDescription, "Deskripsi"); err != nil {
		return err
	}
	if doc.Checksum == "" {
		return errors.New("checksum cadangan wajib diisi")
	}
	expected, err := backupChecksum(doc)
	if err != nil || !strings.EqualFold(expected, doc.Checksum) {
		return errors.New("checksum cadangan tidak cocok")
	}
	return nil
}

func validateBackupSources(doc backupDocument) error {
	for _, source := range doc.Sources {
		if !validID(source.ID) || source.Kind == "" || len([]byte(source.Content)) > maxSourceContent || !verifySourceIntegrity(Source{Content: source.Content, Checksum: source.Checksum}) {
			return errInvalidSourceIntegrity
		}
		if err := validateLength(source.Title, 1, maxTitle, "Judul sumber"); err != nil {
			return err
		}
		for _, location := range source.Locations {
			if !validID(location.ID) || location.StartLine < 1 || location.EndLine < location.StartLine || location.SourceChecksum != source.Checksum {
				return errInvalidBackup
			}
		}
	}
	for _, note := range doc.Notes {
		if !validID(note.ID) || len([]byte(note.Content)) > maxNoteContent {
			return errInvalidBackup
		}
		if err := validateLength(note.Title, 1, maxTitle, "Judul"); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) replaceNotebook(r *http.Request, doc backupDocument) error {
	tx, err := s.db.BeginTx(r.Context(), nil)
	if err != nil {
		return err
	}
	rollback := func(err error) error {
		_ = tx.Rollback()
		return err
	}
	if _, err := tx.ExecContext(r.Context(), `UPDATE notebooks SET title=?,description=?,archived=?,created_at=?,updated_at=? WHERE id=? AND owner_id=?`,
		doc.Notebook.Title, doc.Notebook.Description, boolInt(doc.Notebook.Archived), doc.Notebook.CreatedAt, doc.Notebook.UpdatedAt, doc.Notebook.ID, userID(r)); err != nil {
		return rollback(err)
	}
	if _, err := tx.ExecContext(r.Context(), `DELETE FROM notes WHERE notebook_id=?`, doc.Notebook.ID); err != nil {
		return rollback(err)
	}
	if _, err := tx.ExecContext(r.Context(), `DELETE FROM sources WHERE notebook_id=?`, doc.Notebook.ID); err != nil {
		return rollback(err)
	}
	for _, note := range doc.Notes {
		deletedAt := any(nil)
		if note.DeletedAt != "" {
			deletedAt = note.DeletedAt
		}
		if _, err := tx.ExecContext(r.Context(), `INSERT INTO notes(id,notebook_id,title,content,note_type,metadata_json,created_at,updated_at,deleted_at) VALUES(?,?,?,?,?,?,?,?,?)`,
			note.ID, doc.Notebook.ID, note.Title, note.Content, note.NoteType, note.Metadata, note.CreatedAt, note.UpdatedAt, deletedAt); err != nil {
			return rollback(err)
		}
	}
	for _, source := range doc.Sources {
		if _, err := tx.ExecContext(r.Context(), `INSERT INTO sources(id,notebook_id,title,kind,content,locator,checksum,metadata_json,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?)`,
			source.ID, doc.Notebook.ID, source.Title, source.Kind, source.Content, source.Locator, source.Checksum, source.Metadata, source.CreatedAt, source.UpdatedAt); err != nil {
			return rollback(err)
		}
		for _, location := range source.Locations {
			if _, err := tx.ExecContext(r.Context(), `INSERT INTO source_locations(id,source_id,start_line,end_line,source_checksum,created_at) VALUES(?,?,?,?,?,?)`,
				location.ID, source.ID, location.StartLine, location.EndLine, location.SourceChecksum, location.CreatedAt); err != nil {
				return rollback(err)
			}
		}
	}
	return tx.Commit()
}

func boolInt(value bool) int {
	if value { return 1 }
	return 0
}

func (s *Server) buildBackup(r *http.Request, notebookID string) (backupDocument, error) {
	var doc backupDocument
	var archived int
	err := s.db.QueryRowContext(r.Context(), `SELECT id,title,description,archived,created_at,updated_at FROM notebooks WHERE id=? AND owner_id=?`, notebookID, userID(r)).Scan(
		&doc.Notebook.ID, &doc.Notebook.Title, &doc.Notebook.Description, &archived, &doc.Notebook.CreatedAt, &doc.Notebook.UpdatedAt)
	if err != nil { return doc, err }
	doc.Notebook.Archived = archived != 0
	doc.Format = "catatan-backup"
	doc.Version = backupFormatVersion
	doc.Notes = []ExportNote{}
	doc.Sources = []ExportSource{}

	rows, err := s.db.QueryContext(r.Context(), `SELECT id,title,content,note_type,metadata_json,created_at,updated_at,deleted_at FROM notes WHERE notebook_id=? ORDER BY id ASC`, notebookID)
	if err != nil { return doc, err }
	for rows.Next() {
		var note ExportNote
		var deletedAt sql.NullString
		if err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.NoteType, &note.Metadata, &note.CreatedAt, &note.UpdatedAt, &deletedAt); err != nil {
			rows.Close(); return doc, err
		}
		if deletedAt.Valid { note.DeletedAt = deletedAt.String }
		doc.Notes = append(doc.Notes, note)
	}
	if err := rows.Err(); err != nil { rows.Close(); return doc, err }
	rows.Close()

	rows, err = s.db.QueryContext(r.Context(), `
		SELECT s.id,s.title,s.kind,s.content,s.locator,s.checksum,s.metadata_json,s.created_at,s.updated_at,
		       l.id,l.start_line,l.end_line,l.source_checksum,l.created_at
		FROM sources s
		LEFT JOIN source_locations l ON l.source_id=s.id
		WHERE s.notebook_id=?
		ORDER BY s.id ASC,l.start_line ASC,l.end_line ASC,l.id ASC`, notebookID)
	if err != nil { return doc, err }
	var current *ExportSource
	for rows.Next() {
		var source ExportSource
		var locationID sql.NullString
		var startLine, endLine sql.NullInt64
		var locationChecksum, locationCreatedAt sql.NullString
		if err := rows.Scan(&source.ID, &source.Title, &source.Kind, &source.Content, &source.Locator, &source.Checksum, &source.Metadata, &source.CreatedAt, &source.UpdatedAt,
			&locationID, &startLine, &endLine, &locationChecksum, &locationCreatedAt); err != nil {
			rows.Close(); return doc, err
		}
		if current == nil || current.ID != source.ID {
			if current != nil {
				if !verifySourceIntegrity(Source{Content: current.Content, Checksum: current.Checksum}) {
					rows.Close(); return doc, errInvalidSourceIntegrity
				}
				doc.Sources = append(doc.Sources, *current)
			}
			current = &source
		}
		if locationID.Valid {
			current.Locations = append(current.Locations, ExportLocation{
				ID: locationID.String, StartLine: int(startLine.Int64), EndLine: int(endLine.Int64),
				SourceChecksum: locationChecksum.String, CreatedAt: locationCreatedAt.String})
		}
	}
	if err := rows.Err(); err != nil { rows.Close(); return doc, err }
	rows.Close()
	if current != nil {
		if !verifySourceIntegrity(Source{Content: current.Content, Checksum: current.Checksum}) {
			return doc, errInvalidSourceIntegrity
		}
		doc.Sources = append(doc.Sources, *current)
	}
	checksum, err := backupChecksum(doc)
	if err != nil { return doc, err }
	doc.Checksum = checksum
	return doc, nil
}

func backupChecksum(doc backupDocument) (string, error) {
	payload, err := json.Marshal(struct {
		Format string `json:"format"`
		Version int `json:"version"`
		Notebook ExportNotebook `json:"notebook"`
		Notes []ExportNote `json:"notes"`
		Sources []ExportSource `json:"sources"`
	}{doc.Format, doc.Version, doc.Notebook, doc.Notes, doc.Sources})
	if err != nil { return "", fmt.Errorf("checksum cadangan: %w", err) }
	digest := sha256.Sum256(payload)
	return fmt.Sprintf("%x", digest[:]), nil
}
