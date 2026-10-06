package http

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"
)

type SourceLocation struct {
	ID             string `json:"id"`
	SourceID       string `json:"source_id"`
	StartLine      int    `json:"start_line"`
	EndLine        int    `json:"end_line"`
	SourceChecksum string `json:"source_checksum"`
	CreatedAt      string `json:"created_at"`
}

func (s *Server) listSourceLocations(w http.ResponseWriter, r *http.Request) {
	sourceID := r.PathValue("id")
	if !validID(sourceID) {
		http.Error(w, "ID sumber tidak valid", http.StatusBadRequest)
		return
	}
	if !s.sourceOwnedBy(r, sourceID) {
		http.NotFound(w, r)
		return
	}

	rows, err := s.db.QueryContext(r.Context(), `SELECT id,source_id,start_line,end_line,source_checksum,created_at FROM source_locations WHERE source_id=? ORDER BY start_line,end_line,id`, sourceID)
	if err != nil {
		serverError(w, err)
		return
	}
	defer rows.Close()

	locations := []SourceLocation{}
	for rows.Next() {
		var location SourceLocation
		if err := rows.Scan(&location.ID, &location.SourceID, &location.StartLine, &location.EndLine, &location.SourceChecksum, &location.CreatedAt); err != nil {
			serverError(w, err)
			return
		}
		locations = append(locations, location)
	}
	if err := rows.Err(); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, locations)
}

func (s *Server) createSourceLocation(w http.ResponseWriter, r *http.Request) {
	sourceID := r.PathValue("id")
	if !validID(sourceID) {
		http.Error(w, "ID sumber tidak valid", http.StatusBadRequest)
		return
	}

	var in struct {
		StartLine int `json:"start_line"`
		EndLine   int `json:"end_line"`
	}
	if !decode(w, r, &in) {
		return
	}
	if in.StartLine < 1 || in.EndLine < in.StartLine {
		http.Error(w, "Rentang baris tidak valid", http.StatusBadRequest)
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
	if !verifySourceIntegrity(source) {
		http.Error(w, "Integritas isi sumber tidak dapat diverifikasi", http.StatusUnprocessableEntity)
		return
	}

	lineCount := strings.Count(source.Content, "\n") + 1
	if in.EndLine > lineCount {
		http.Error(w, "Rentang baris melebihi isi sumber", http.StatusBadRequest)
		return
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	location := SourceLocation{
		ID:             newID(),
		SourceID:       source.ID,
		StartLine:      in.StartLine,
		EndLine:        in.EndLine,
		SourceChecksum: source.Checksum,
		CreatedAt:      now,
	}
	_, err = s.db.ExecContext(r.Context(), `INSERT INTO source_locations(id,source_id,start_line,end_line,source_checksum,created_at) VALUES(?,?,?,?,?,?)`, location.ID, location.SourceID, location.StartLine, location.EndLine, location.SourceChecksum, location.CreatedAt)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, location)
}

func (s *Server) sourceOwnedBy(r *http.Request, sourceID string) bool {
	var exists int
	err := s.db.QueryRowContext(r.Context(), `SELECT 1 FROM sources WHERE id=? AND notebook_id IN (SELECT id FROM notebooks WHERE owner_id=?) LIMIT 1`, sourceID, userID(r)).Scan(&exists)
	return err == nil && exists == 1
}

func verifySourceLocation(source Source, location SourceLocation) bool {
	if location.SourceID != source.ID || location.StartLine < 1 || location.EndLine < location.StartLine {
		return false
	}
	if !verifySourceIntegrity(source) || !strings.EqualFold(location.SourceChecksum, source.Checksum) {
		return false
	}
	lineCount := strings.Count(source.Content, "\n") + 1
	return location.EndLine <= lineCount
}
