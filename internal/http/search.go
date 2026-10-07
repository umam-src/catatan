package http

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/umam-src/catatan/internal/db"
)

type searchResponse struct { Results []searchResult `json:"results"` }

type searchResult struct {
	ID string `json:"id"`
	Kind string `json:"kind"`
	NotebookID string `json:"notebook_id"`
	Title string `json:"title"`
	Relevance float64 `json:"relevance"`
}

func (s *Server) search(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" { writeJSON(w, http.StatusOK, searchResponse{Results: []searchResult{}}); return }
	notebookID := strings.TrimSpace(r.URL.Query().Get("notebook_id"))
	if notebookID != "" && !validNotebookID(notebookID) { http.Error(w, "ID buku tidak valid", http.StatusBadRequest); return }
	limit := 0
	if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed < 1 { http.Error(w, "Batas hasil tidak valid", http.StatusBadRequest); return }
		limit = parsed
	}
	results, err := s.db.Search(r.Context(), userID(r), notebookID, query, limit)
	if err != nil {
		if errors.Is(err, db.ErrInvalidSearchQuery) { http.Error(w, "Kueri pencarian tidak valid", http.StatusBadRequest); return }
		serverError(w, err); return
	}
	out := searchResponse{Results: make([]searchResult, 0, len(results))}
	for _, result := range results { out.Results = append(out.Results, searchResult{ID: result.ID, Kind: result.Kind, NotebookID: result.NotebookID, Title: result.Title, Relevance: result.Relevance}) }
	writeJSON(w, http.StatusOK, out)
}
