package db

import (
	"context"
	"errors"
	"strings"
)

var ErrInvalidSearchQuery = errors.New("kueri pencarian tidak valid")

const (
	defaultSearchLimit = 20
	maxSearchLimit     = 100
)

type SearchResult struct {
	ID string `json:"id"`
	Kind string `json:"kind"`
	NotebookID string `json:"notebook_id"`
	Title string `json:"title"`
	Content string `json:"content"`
	Relevance float64 `json:"relevance"`
}

func (d *DB) Search(ctx context.Context, ownerID, notebookID, query string, limit int) ([]SearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" { return []SearchResult{}, nil }
	if limit <= 0 { limit = defaultSearchLimit }
	if limit > maxSearchLimit { limit = maxSearchLimit }
	if ownerID == "" { return nil, errors.New("pemilik pencarian wajib diisi") }
	if strings.Count(query, `"`)%2 != 0 { return nil, ErrInvalidSearchQuery }

	args := []any{query, ownerID}
	filter := ""
	if notebookID != "" { filter = " AND sd.notebook_id = ?"; args = append(args, notebookID) }
	args = append(args, limit)
	rows, err := d.QueryContext(ctx, `SELECT sd.id, sd.kind, sd.notebook_id, sd.title, sd.content,
		bm25(search_index, 5.0, 1.0) AS relevance
		FROM search_index
		JOIN search_documents sd ON sd.rowid = search_index.rowid
		JOIN notebooks n ON n.id = sd.notebook_id
		WHERE search_index MATCH ? AND n.owner_id = ?`+filter+`
		ORDER BY relevance ASC, sd.rowid ASC
		LIMIT ?`, args...)
	if err != nil { return nil, err }
	defer rows.Close()
	results := make([]SearchResult, 0)
	for rows.Next() {
		var result SearchResult
		if err := rows.Scan(&result.ID, &result.Kind, &result.NotebookID, &result.Title, &result.Content, &result.Relevance); err != nil { return nil, err }
		results = append(results, result)
	}
	if err := rows.Err(); err != nil { return nil, err }
	return results, nil
}
