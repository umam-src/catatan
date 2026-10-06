package http

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
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
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	NoteType  string `json:"note_type"`
	Metadata  string `json:"metadata_json"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	DeletedAt string `json:"deleted_at,omitempty"`
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

var errInvalidSourceIntegrity = errors.New("integritas sumber tidak valid")

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

	if err := s.validateExportSources(r, notebookID); err != nil {
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
	encoder := json.NewEncoder(w)
	if _, err := w.Write([]byte(`{"format":"catatan-export","version":1,"notebook":`)); err != nil {
		return
	}
	if err := encoder.Encode(notebook); err != nil {
		return
	}
	if _, err := w.Write([]byte(`,"notes":`)); err != nil {
		return
	}
	if err := s.encodeNotes(w, encoder, r, notebookID); err != nil {
		return
	}
	if _, err := w.Write([]byte(`,"sources":`)); err != nil {
		return
	}
	if err := s.encodeSources(w, encoder, r, notebookID); err != nil {
		return
	}
	_, _ = w.Write([]byte("}\n"))
}

func (s *Server) validateExportSources(r *http.Request, notebookID string) error {
	rows, err := s.db.QueryContext(r.Context(), `SELECT content,checksum FROM sources WHERE notebook_id=? ORDER BY id ASC`, notebookID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var content, checksum string
		if err := rows.Scan(&content, &checksum); err != nil {
			return err
		}
		if !verifySourceIntegrity(Source{Content: content, Checksum: checksum}) {
			return errInvalidSourceIntegrity
		}
	}
	return rows.Err()
}

func (s *Server) encodeNotes(w http.ResponseWriter, encoder *json.Encoder, r *http.Request, notebookID string) error {
	if _, err := w.Write([]byte("[")); err != nil {
		return err
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT id,title,content,note_type,metadata_json,created_at,updated_at,deleted_at FROM notes WHERE notebook_id=? ORDER BY id ASC`, notebookID)
	if err != nil {
		return err
	}
	defer rows.Close()

	first := true
	for rows.Next() {
		var note ExportNote
		var deletedAt sql.NullString
		if err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.NoteType, &note.Metadata, &note.CreatedAt, &note.UpdatedAt, &deletedAt); err != nil {
			return err
		}
		if deletedAt.Valid {
			note.DeletedAt = deletedAt.String
		}
		if !first {
			if _, err := w.Write([]byte(",")); err != nil {
				return err
			}
		}
		if err := encoder.Encode(note); err != nil {
			return err
		}
		first = false
	}
	if err := rows.Err(); err != nil {
		return err
	}
	_, err = w.Write([]byte("]"))
	return err
}

func (s *Server) encodeSources(w http.ResponseWriter, encoder *json.Encoder, r *http.Request, notebookID string) error {
	if _, err := w.Write([]byte("[")); err != nil {
		return err
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT id,title,kind,content,locator,checksum,metadata_json,created_at,updated_at FROM sources WHERE notebook_id=? ORDER BY id ASC`, notebookID)
	if err != nil {
		return err
	}
	defer rows.Close()

	first := true
	for rows.Next() {
		var source ExportSource
		if err := rows.Scan(&source.ID, &source.Title, &source.Kind, &source.Content, &source.Locator, &source.Checksum, &source.Metadata, &source.CreatedAt, &source.UpdatedAt); err != nil {
			return err
		}
		locations, err := s.exportSourceLocations(r, source.ID)
		if err != nil {
			return err
		}
		source.Locations = locations
		if !first {
			if _, err := w.Write([]byte(",")); err != nil {
				return err
			}
		}
		if err := encoder.Encode(source); err != nil {
			return err
		}
		first = false
	}
	if err := rows.Err(); err != nil {
		return err
	}
	_, err = w.Write([]byte("]"))
	return err
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
