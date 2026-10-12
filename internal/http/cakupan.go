package http

import (
	"database/sql"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/umam-src/catatan/internal/db"
)

// Cakupan percakapan menentukan sumber pengetahuan AI dan disimpan pada percakapan.
const (
	cakupanSemuaCatatan = "semua_catatan" // seluruh Catatan pada buku, diambil per potongan
	cakupanSatuCatatan  = "satu_catatan"  // hanya satu Catatan tertentu
	cakupanDokumen      = "dokumen"       // hanya Sumber (dokumen) yang dipilih pada setiap pesan
)

func validCakupan(c string) bool {
	return c == cakupanSemuaCatatan || c == cakupanSatuCatatan || c == cakupanDokumen
}

func (s *Server) listConversations(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !validNotebookID(id) {
		http.Error(w, "ID buku tidak valid", 400)
		return
	}
	ok, err := s.notebookOwnedBy(r, id)
	if err != nil {
		serverError(w, err)
		return
	}
	if !ok {
		http.NotFound(w, r)
		return
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT id,notebook_id,title,cakupan,COALESCE(catatan_id,''),created_at,updated_at FROM conversations WHERE notebook_id=? ORDER BY updated_at DESC`, id)
	if err != nil {
		serverError(w, err)
		return
	}
	defer rows.Close()
	out := []Conversation{}
	for rows.Next() {
		var c Conversation
		if err := rows.Scan(&c.ID, &c.NotebookID, &c.Title, &c.Scope, &c.NoteID, &c.CreatedAt, &c.UpdatedAt); err != nil {
			serverError(w, err)
			return
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, 200, out)
}

func (s *Server) createConversation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !validNotebookID(id) {
		http.Error(w, "ID buku tidak valid", 400)
		return
	}
	ok, err := s.notebookOwnedBy(r, id)
	if err != nil {
		serverError(w, err)
		return
	}
	if !ok {
		http.NotFound(w, r)
		return
	}
	var in struct {
		Title  string `json:"title"`
		Scope  string `json:"scope"`
		NoteID string `json:"note_id"`
	}
	if !decode(w, r, &in) {
		return
	}
	in.Title = strings.TrimSpace(in.Title)
	in.Scope = strings.TrimSpace(in.Scope)
	in.NoteID = strings.TrimSpace(in.NoteID)
	if err := validateLength(in.Title, 0, maxConversationTitle, "Judul percakapan"); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if in.Scope == "" {
		in.Scope = cakupanDokumen
	}
	if !validCakupan(in.Scope) {
		http.Error(w, "Cakupan percakapan tidak valid", 400)
		return
	}
	if in.Scope == cakupanSatuCatatan {
		if !validID(in.NoteID) {
			http.Error(w, "Cakupan satu Catatan wajib menyertakan ID catatan yang valid", 400)
			return
		}
		var n int
		err := s.db.QueryRowContext(r.Context(), `SELECT 1 FROM notes WHERE id=? AND notebook_id=? AND deleted_at IS NULL`, in.NoteID, id).Scan(&n)
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Catatan cakupan tidak ditemukan", 404)
			return
		}
		if err != nil {
			serverError(w, err)
			return
		}
	} else if in.NoteID != "" {
		http.Error(w, "ID catatan hanya berlaku untuk cakupan satu Catatan", 400)
		return
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	c := Conversation{ID: newID(), NotebookID: id, Title: in.Title, Scope: in.Scope, NoteID: in.NoteID, CreatedAt: now, UpdatedAt: now}
	var catatanID any
	if c.NoteID != "" {
		catatanID = c.NoteID
	}
	if _, err = s.db.ExecContext(r.Context(), `INSERT INTO conversations(id,notebook_id,title,cakupan,catatan_id,created_at,updated_at) VALUES(?,?,?,?,?,?,?)`,
		c.ID, c.NotebookID, c.Title, c.Scope, catatanID, c.CreatedAt, c.UpdatedAt); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, 201, c)
}

// deleteConversation menghapus percakapan beserta pesannya. Catatan dan Sumber tidak disentuh.
func (s *Server) deleteConversation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !validID(id) {
		http.Error(w, "ID percakapan tidak valid", 400)
		return
	}
	if !s.conversationOwnedBy(r, id) {
		http.NotFound(w, r)
		return
	}
	tx, err := s.db.BeginTx(r.Context(), nil)
	if err != nil {
		serverError(w, err)
		return
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(r.Context(), `DELETE FROM messages WHERE conversation_id=?`, id); err != nil {
		serverError(w, err)
		return
	}
	if _, err := tx.ExecContext(r.Context(), `DELETE FROM conversations WHERE id=?`, id); err != nil {
		serverError(w, err)
		return
	}
	if err := tx.Commit(); err != nil {
		serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// muatKonteksPercakapan menyusun konteks sesuai cakupan yang tersimpan pada
// percakapan. Pilihan pada pesan yang tidak sesuai cakupan ditolak, bukan
// diabaikan. Hasil keempat menyatakan apakah konteks wajib ada pada jawaban.
func (s *Server) muatKonteksPercakapan(r *http.Request, percakapanID, pertanyaan, noteID string, sourceIDs []string) ([]conversationNote, []conversationSource, bool, error) {
	var bukuID, cakupan string
	var catatanID sql.NullString
	err := s.db.QueryRowContext(r.Context(), `SELECT notebook_id,cakupan,catatan_id FROM conversations WHERE id=? AND notebook_id IN (SELECT id FROM notebooks WHERE owner_id=?)`,
		percakapanID, userID(r)).Scan(&bukuID, &cakupan, &catatanID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, false, &httpError{404, "Percakapan tidak ditemukan"}
	}
	if err != nil {
		return nil, nil, false, err
	}
	switch cakupan {
	case cakupanSemuaCatatan:
		if noteID != "" || len(sourceIDs) > 0 {
			return nil, nil, false, &httpError{400, "Cakupan semua Catatan tidak menerima pilihan Catatan atau Sumber"}
		}
		hasil, err := s.db.CariPotonganJenis(r.Context(), userID(r), bukuID, "note", pertanyaan, batasPotonganKonteks, anggaranKonteksByte)
		if err != nil {
			return nil, nil, false, err
		}
		return kelompokkanCatatan(hasil), nil, true, nil
	case cakupanSatuCatatan:
		if noteID != "" || len(sourceIDs) > 0 {
			return nil, nil, false, &httpError{400, "Cakupan satu Catatan tidak menerima pilihan Catatan atau Sumber"}
		}
		if !catatanID.Valid {
			return nil, nil, false, &httpError{409, "Percakapan tidak memiliki Catatan cakupan"}
		}
		note, err := s.loadConversationNote(r, percakapanID, catatanID.String)
		if err != nil {
			return nil, nil, false, err
		}
		catatan := []conversationNote{*note}
		if err := s.pilihPotongan(r, percakapanID, pertanyaan, catatan, nil); err != nil {
			return nil, nil, false, err
		}
		return catatan, nil, true, nil
	default:
		if noteID != "" {
			return nil, nil, false, &httpError{400, "Cakupan dokumen tidak menerima Catatan"}
		}
		sources, err := s.loadConversationSources(r, percakapanID, sourceIDs)
		if err != nil {
			return nil, nil, false, err
		}
		if err := s.pilihPotongan(r, percakapanID, pertanyaan, nil, sources); err != nil {
			return nil, nil, false, err
		}
		return nil, sources, len(sources) > 0, nil
	}
}

// kelompokkanCatatan mengelompokkan potongan hasil pengambilan per Catatan dengan
// urutan kemunculan pertama (paling relevan lebih dulu).
func kelompokkanCatatan(hasil []db.HasilPotongan) []conversationNote {
	catatan := []conversationNote{}
	indeks := map[string]int{}
	for _, h := range hasil {
		id := strings.TrimPrefix(h.DokumenID, "note:")
		i, ada := indeks[id]
		if !ada {
			i = len(catatan)
			indeks[id] = i
			catatan = append(catatan, conversationNote{ID: id, Title: h.Judul, DariPotongan: true})
		}
		catatan[i].Potongan = append(catatan[i].Potongan, h)
	}
	for i := range catatan {
		p := catatan[i].Potongan
		sort.SliceStable(p, func(a, b int) bool { return p[a].Urutan < p[b].Urutan })
	}
	return catatan
}
