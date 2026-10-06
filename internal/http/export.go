package http

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

const exportFormatVersion = 1

type ExportDocument struct {
	Format   string         `json:"format"`
	Version  int            `json:"version"`
	Notebook ExportNotebook `json:"notebook"`
	Notes    []ExportNote   `json:"notes"`
	Sources  []ExportSource `json:"sources"`
}

type ExportNotebook struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Archived    bool   `json:"archived"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type ExportNote struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	NoteType   string `json:"note_type"`
	Metadata   string `json:"metadata_json"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	DeletedAt  string `json:"deleted_at,omitempty"`
}

type ExportSource struct {
	ID        string           `json:"id"`
	Title     string           `json:"title"`
	Kind      string           `json:"kind"`
	Content   string           `json:"content"`
	Locator   string           `json:"locator"`
	Checksum  string           `json:"checksum"`
	Metadata  string           `json:"metadata_json"`
	CreatedAt string           `json:"created_at"`
	UpdatedAt string           `json:"updated_at"`
	Locations []ExportLocation `json:"locations"`
}

type ExportLocation struct {
	ID             string `json:"id"`
	StartLine      int    `json:"start_line"`
	EndLine        int    `json:"end_line"`
	SourceChecksum string `json:"source_checksum"`
	CreatedAt      string `json:"created_at"`
}

func (s *Server) exportNotebook(w http.ResponseWriter, r *http.Request) {
	notebookID := r.PathValue("id")
	if !validID(notebookID) {
		http.Error(w, "ID buku tidak valid", http.StatusBadRequest)
		return
	}

	var notebook ExportNotebook
	var archived int
	err := s.db.QueryRowContext(r.Context(), `SELECT id,title,description,archived,created_at,updated_at FROM notebooks WHERE id=? AND owner_id=?`, notebookID, userID(r)).Scan(
		&notebook.ID, &notebook.Title, &notebook.Description, &archived, &notebook.CreatedAt, &notebook.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}
	notebook.Archived = archived != 0

	notes, err := s.exportNotes(r, notebookID)
	if err != nil {
		serverError(w, err)
		return
	}
	sources, err := s.exportSources(r, notebookID)
	if err != nil {
		if errors.Is(err, errInvalidSourceIntegrity) {
			http.Error(w, "Integritas sumber tidak dapat diverifikasi", http.StatusUnprocessableEntity)
			return
		}
		serverError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="catatan-`+notebook.ID+`.json"`)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(ExportDocument{
		Format:   "catatan-export",
		Version:  exportFormatVersion,
		Notebook: notebook,
		Notes:    notes,
		Sources:  sources,
	})
}

var errInvalidSourceIntegrity = errors.New("integritas sumber tidak valid")

func (s *Server) exportNotes(r *http.Request, notebookID string) ([]ExportNote, error) {
	rows, err := s.db.QueryContext(r.Context(), `SELECT id,title,content,note_type,metadata_json,created_at,updated_at,deleted_at FROM notes WHERE notebook_id=? ORDER BY id ASC`, notebookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notes := make([]ExportNote, 0)
	for rows.Next() {
		var note ExportNote
		var deletedAt sql.NullString
		if err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.NoteType, &note.Metadata, &note.CreatedAt, &note.UpdatedAt, &deletedAt); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			note.DeletedAt = deletedAt.String
		}
		notes = append(notes, note)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return notes, nil
}

func (s *Server) exportSources(r *http.Request, notebookID string) ([]ExportSource, error) {
	rows, err := s.db.QueryContext(r.Context(), `SELECT id,title,kind,content,locator,checksum,metadata_json,created_at,updated_at FROM sources WHERE notebook_id=? ORDER BY id ASC`, notebookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sources := make([]ExportSource, 0)
	for rows.Next() {
		var source ExportSource
		if err := rows.Scan(&source.ID, &source.Title, &source.Kind, &source.Content, &source.Locator, &source.Checksum, &source.Metadata, &source.CreatedAt, &source.UpdatedAt); err != nil {
			return nil, err
		}
		if !verifySourceIntegrity(Source{Content: source.Content, Checksum: source.Checksum}) {
			return nil, errInvalidSourceIntegrity
		}
		locations, err := s.exportSourceLocations(r, source.ID)
		if err != nil {
			return nil, err
		}
		source.Locations = locations
		sources = append(sources, source)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return sources, nil
}

func (s *Server) exportSourceLocations(r *http.Request, sourceID string) ([]ExportLocation, error) {
	rows, err := s.db.QueryContext(r.Context(), `SELECT id,start_line,end_line,source_checksum,created_at FROM source_locations WHERE source_id=? ORDER BY start_line ASC,end_line ASC,id ASC`, sourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	locations := make([]ExportLocation, 0)
	for rows.Next() {
		var location ExportLocation
		if err := rows.Scan(&location.ID, &location.StartLine, &location.EndLine, &location.SourceChecksum, &location.CreatedAt); err != nil {
			return nil, err
		}
		locations = append(locations, location)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return locations, nil
}

func exportTimestamp(now time.Time) string {
	return now.UTC().Format(time.RFC3339Nano)
}
