package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/umam-src/catatan/internal/ai"
	"github.com/umam-src/catatan/internal/db"
)

type fakeConversationProvider struct {
	response string
	requests []ai.Request
}

func (p *fakeConversationProvider) Probe(context.Context) ai.ProviderStatus { return ai.ProviderReady }
func (p *fakeConversationProvider) Generate(_ context.Context, req ai.Request) (ai.Response, error) {
	p.requests = append(p.requests, req)
	return ai.Response{Text: p.response, Model: req.Model}, nil
}

func conversationHandler(t *testing.T, provider *fakeConversationProvider) http.Handler {
	t.Helper()
	d, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "catatan.db"))
	if err != nil { t.Fatal(err) }
	t.Cleanup(func() { _ = d.Close() })

	h := NewWithAI(d, provider)
	setup := requestUji(t, h, http.MethodPost, "/api/auth/setup", map[string]string{
		"username": "pengguna", "email": "pengguna@lokal.invalid",
		"display_name": "Pengguna Uji", "password": "kata-sandi-uji-aman",
	})
	if setup.Code != http.StatusCreated { t.Fatalf("setup: %d", setup.Code) }
	cookie := setup.Result().Cookies()[0]
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.Clone(r.Context())
		r.AddCookie(cookie)
		h.ServeHTTP(w, r)
	})
}

func buatPercakapanUji(t *testing.T, h http.Handler) Conversation {
	t.Helper()
	r := requestUji(t, h, http.MethodPost, "/api/notebooks/default/conversations", map[string]string{"title": "Uji"})
	if r.Code != http.StatusCreated { t.Fatalf("buat percakapan: %d %s", r.Code, r.Body.String()) }
	var c Conversation
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil { t.Fatal(err) }
	return c
}

func TestConversationContextAndCitation(t *testing.T) {
	old := os.Getenv("CATATAN_AI_MODEL")
	t.Cleanup(func() { _ = os.Setenv("CATATAN_AI_MODEL", old) })
	_ = os.Setenv("CATATAN_AI_MODEL", "uji-model")

	p := &fakeConversationProvider{response: "Jawaban [S1:L1-L2]"}
	h := conversationHandler(t, p)
	c := buatPercakapanUji(t, h)

	s := permintaanSumber(t, h, "default", "bahan.txt", "Bahan", "baris satu\nbaris dua\nbaris tiga")
	if s.Code != http.StatusCreated { t.Fatalf("sumber: %d", s.Code) }
	var src Source
	if err := json.NewDecoder(s.Body).Decode(&src); err != nil { t.Fatal(err) }

	r := requestUji(t, h, http.MethodPost, "/api/conversations/"+c.ID+"/messages",
		map[string]any{"content": "Apa isi bahan?", "source_ids": []string{src.ID}})
	if r.Code != http.StatusCreated { t.Fatalf("pesan: %d %s", r.Code, r.Body.String()) }

	var out map[string]any
	if err := json.NewDecoder(r.Body).Decode(&out); err != nil { t.Fatal(err) }
	assistant := out["assistant_message"].(map[string]any)
	if len(assistant["citations"].([]any)) != 1 { t.Fatalf("kutipan: %#v", assistant["citations"]) }

	var prompt string
	for _, m := range p.requests[0].Messages { prompt += m.Content }
	if !strings.Contains(prompt, "baris satu") || !strings.Contains(prompt, "[L2] baris dua") {
		t.Fatal("konteks tidak terkirim")
	}
	last := p.requests[0].Messages[len(p.requests[0].Messages)-1]
	if last.Role != ai.RoleUser || !strings.Contains(last.Content, "KONTEKS TERPILIH:") || !strings.Contains(last.Content, "PERTANYAAN PENGGUNA:") {
		t.Fatalf("konteks harus menyertai pertanyaan pengguna: %#v", last)
	}
	system := p.requests[0].Messages[0]
	if system.Role != ai.RoleSystem ||
		!strings.Contains(system.Content, "jangan keluarkan kutipan baris") ||
		!strings.Contains(system.Content, "judul Catatan/dokumen") ||
		!strings.Contains(system.Content, noInfoMarker) ||
		!strings.Contains(system.Content, "jangan menambahkan klaim faktual atau kutipan") {
		t.Fatalf("instruksi kutipan sumber harus eksplisit: %#v", system)
	}
}

func TestConversationRejectsInvalidCitationAndUnselectedSource(t *testing.T) {
	old := os.Getenv("CATATAN_AI_MODEL")
	t.Cleanup(func() { _ = os.Setenv("CATATAN_AI_MODEL", old) })
	_ = os.Setenv("CATATAN_AI_MODEL", "uji-model")

	p := &fakeConversationProvider{response: "Jawaban [S9:L1-L2]"}
	h := conversationHandler(t, p)
	c := buatPercakapanUji(t, h)
	a := permintaanSumber(t, h, "default", "a.txt", "A", "rahasia-a")
	b := permintaanSumber(t, h, "default", "b.txt", "B", "rahasia-b")
	var sa, sb Source
	if err := json.NewDecoder(a.Body).Decode(&sa); err != nil { t.Fatal(err) }
	if err := json.NewDecoder(b.Body).Decode(&sb); err != nil { t.Fatal(err) }

	r := requestUji(t, h, http.MethodPost, "/api/conversations/"+c.ID+"/messages",
		map[string]any{"content": "Tanya", "source_ids": []string{sa.ID}})
	if r.Code != http.StatusUnprocessableEntity { t.Fatalf("kutipan tidak sah: %d", r.Code) }
	if !strings.Contains(r.Body.String(), "(kode: ") { t.Fatalf("kode alasan tidak ada: %s", r.Body.String()) }

	p.response = "Jawaban [S1:L1]"
	r = requestUji(t, h, http.MethodPost, "/api/conversations/"+c.ID+"/messages",
		map[string]any{"content": "Tanya", "source_ids": []string{sa.ID}})
	if r.Code != http.StatusCreated { t.Fatalf("pesan kedua: %d %s", r.Code, r.Body.String()) }
	prompt := p.requests[len(p.requests)-1].Messages[1].Content
	if strings.Contains(prompt, "rahasia-b") { t.Fatal("sumber yang tidak dipilih ikut terkirim") }
	_ = sb
}


func TestConversationIncludesActiveNoteContextAndCitation(t *testing.T) {
	old := os.Getenv("CATATAN_AI_MODEL")
	t.Cleanup(func() { _ = os.Setenv("CATATAN_AI_MODEL", old) })
	_ = os.Setenv("CATATAN_AI_MODEL", "uji-model")

	p := &fakeConversationProvider{response: "Kode catatan KJ-7319 [N1:L1-L2]"}
	h := conversationHandler(t, p)
	conversation := buatPercakapanUji(t, h)
	noteResponse := requestUji(t, h, http.MethodPost, "/api/notebooks/default/notes", map[string]string{
		"title": "Catatan pengiriman",
		"content": "Kode pengiriman: KJ-7319\nStatus: Menunggu pemeriksaan gudang.",
	})
	if noteResponse.Code != http.StatusCreated {
		t.Fatalf("buat catatan: %d %s", noteResponse.Code, noteResponse.Body.String())
	}
	var note Note
	if err := json.NewDecoder(noteResponse.Body).Decode(&note); err != nil { t.Fatal(err) }

	response := requestUji(t, h, http.MethodPost, "/api/conversations/"+conversation.ID+"/messages", map[string]any{
		"content": "Apa kode dan status pengiriman?",
		"note_id": note.ID,
		"source_ids": []string{},
	})
	if response.Code != http.StatusCreated { t.Fatalf("pesan: %d %s", response.Code, response.Body.String()) }
	last := p.requests[0].Messages[len(p.requests[0].Messages)-1]
	if !strings.Contains(last.Content, "[N1] Catatan: Catatan pengiriman") ||
		!strings.Contains(last.Content, "[L1] Kode pengiriman: KJ-7319") ||
		!strings.Contains(last.Content, "[L2] Status: Menunggu pemeriksaan gudang.") {
		t.Fatalf("isi Catatan tidak masuk ke konteks provider: %q", last.Content)
	}
	var out map[string]any
	if err := json.NewDecoder(response.Body).Decode(&out); err != nil { t.Fatal(err) }
	assistant := out["assistant_message"].(map[string]any)
	citations := assistant["citations"].([]any)
	if len(citations) != 1 { t.Fatalf("kutipan: %#v", citations) }
	citation := citations[0].(map[string]any)
	if citation["kind"] != "note" || citation["source_id"] != note.ID || citation["source_ref"] != "N1" {
		t.Fatalf("provenance kutipan Catatan salah: %#v", citation)
	}
	userMessage := out["user_message"].(map[string]any)
	if userMessage["note_id"] != note.ID { t.Fatalf("note_id tidak dikembalikan: %#v", userMessage) }
}

func TestConversationRejectsNoteFromAnotherNotebook(t *testing.T) {
	p := &fakeConversationProvider{response: "Jawaban [N1:L1]"}
	h := conversationHandler(t, p)
	conversation := buatPercakapanUji(t, h)
	notebookResponse := requestUji(t, h, http.MethodPost, "/api/notebooks", map[string]string{"title": "Buku lain"})
	if notebookResponse.Code != http.StatusCreated { t.Fatalf("buat buku: %d %s", notebookResponse.Code, notebookResponse.Body.String()) }
	var notebook Notebook
	if err := json.NewDecoder(notebookResponse.Body).Decode(&notebook); err != nil { t.Fatal(err) }
	noteResponse := requestUji(t, h, http.MethodPost, "/api/notebooks/"+notebook.ID+"/notes", map[string]string{"title": "Catatan privat", "content": "tidak boleh ikut"})
	if noteResponse.Code != http.StatusCreated { t.Fatalf("buat catatan: %d %s", noteResponse.Code, noteResponse.Body.String()) }
	var note Note
	if err := json.NewDecoder(noteResponse.Body).Decode(&note); err != nil { t.Fatal(err) }

	response := requestUji(t, h, http.MethodPost, "/api/conversations/"+conversation.ID+"/messages", map[string]any{"content": "Tanya", "note_id": note.ID})
	if response.Code != http.StatusNotFound { t.Fatalf("catatan lintas buku harus ditolak: %d %s", response.Code, response.Body.String()) }
	if len(p.requests) != 0 { t.Fatal("provider dipanggil meski Catatan tidak diizinkan") }
}

func TestValidateCitationsReasonCodes(t *testing.T) {
	refs := map[string]contextRef{
		"S1": {Kind: "source", SourceID: "s1", LineCount: 10},
		"S2": {Kind: "source", SourceID: "s2", LineCount: 100},
		"N1": {Kind: "note", SourceID: "n1", LineCount: 5},
	}
	cases := []struct {
		nama   string
		teks   string
		kode   string
		jumlah int
	}{
		{"sumber valid", "Fakta [S1:L3].", "", 1},
		{"rentang valid", "Fakta [S1:L2-L4].", "", 1},
		{"catatan valid", "Fakta [N1:L5].", "", 1},
		{"kutipan sama diringkas", "A [S1:L3] B [S1:L3].", "", 1},
		{"rentang 50 baris masih valid", "Fakta [S2:L1-L51].", "", 1},
		{"tanpa kutipan", "Informasi tidak ditemukan dalam konteks.", "tanpa_kutipan", 0},
		{"sumber tidak dipilih", "Fakta [S3:L1].", "referensi_tidak_dikenal", 0},
		{"catatan tidak ada", "Fakta [N2:L1].", "referensi_tidak_dikenal", 0},
		{"satu salah menolak semua", "Benar [S1:L1] salah [S3:L1].", "referensi_tidak_dikenal", 0},
		{"baris nol", "Fakta [S1:L0].", "rentang_tidak_valid", 0},
		{"melewati akhir", "Fakta [S1:L11].", "rentang_tidak_valid", 0},
		{"rentang terbalik", "Fakta [S1:L4-L2].", "rentang_tidak_valid", 0},
		{"rentang terlalu panjang", "Fakta [S2:L1-L52].", "rentang_terlalu_panjang", 0},
	}
	for _, c := range cases {
		t.Run(c.nama, func(t *testing.T) {
			matches, kode := validateCitations(c.teks, refs)
			if kode != c.kode { t.Fatalf("kode = %q, ingin %q", kode, c.kode) }
			if len(matches) != c.jumlah { t.Fatalf("jumlah kutipan = %d, ingin %d", len(matches), c.jumlah) }
		})
	}
}


func TestValidateNoInfoMarkerAndCitationNormalization(t *testing.T) {
	refs := map[string]contextRef{
		"S1": {Kind: "source", SourceID: "s1", LineCount: 10},
		"N1": {Kind: "note", SourceID: "n1", LineCount: 5},
	}
	cases := []struct {
		name string
		text string
		wantReason string
		wantCount int
	}{
		{"marker no info valid", "[TIDAK_DITEMUKAN] Informasi tidak ditemukan dalam konteks.", "", 0},
		{"marker must be first", "Catatan: [TIDAK_DITEMUKAN] Informasi tidak ditemukan dalam konteks.", "format_tidak_ditemukan", 0},
		{"marker with citation rejected", "[TIDAK_DITEMUKAN] Informasi tidak ditemukan dalam konteks. [S1:L1]", "format_tidak_ditemukan", 0},
		{"marker with factual claim rejected", "[TIDAK_DITEMUKAN] Status pengiriman adalah selesai.", "format_tidak_ditemukan", 0},
		{"marker too long rejected", "[TIDAK_DITEMUKAN] Informasi tidak ditemukan dalam konteks. " + strings.Repeat("x", 240), "format_tidak_ditemukan", 0},
		{"factual answer without citation rejected", "Status pengiriman selesai.", "tanpa_kutipan", 0},
		{"space after colon normalized", "Fakta [S1: L3].", "", 1},
		{"short range normalized", "Fakta [S1:L3-5].", "", 1},
		{"unicode dash normalized", "Fakta [S1:L3–L5].", "", 1},
		{"note citation normalized", "Fakta [N1 : L2 - 4].", "", 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			matches, reason := validateCitations(tc.text, refs)
			if reason != tc.wantReason { t.Fatalf("reason = %q, want %q", reason, tc.wantReason) }
			if len(matches) != tc.wantCount { t.Fatalf("citation count = %d, want %d", len(matches), tc.wantCount) }
		})
	}
}

func TestConversationAcceptsAndStripsNoInfoMarker(t *testing.T) {
	old := os.Getenv("CATATAN_AI_MODEL")
	t.Cleanup(func() { _ = os.Setenv("CATATAN_AI_MODEL", old) })
	_ = os.Setenv("CATATAN_AI_MODEL", "uji-model")

	p := &fakeConversationProvider{response: "[TIDAK_DITEMUKAN] Informasi tidak ditemukan dalam konteks."}
	h := conversationHandler(t, p)
	c := buatPercakapanUji(t, h)
	s := permintaanSumber(t, h, "default", "bahan.txt", "Bahan", "Status: belum diketahui")
	var src Source
	if err := json.NewDecoder(s.Body).Decode(&src); err != nil { t.Fatal(err) }

	r := requestUji(t, h, http.MethodPost, "/api/conversations/"+c.ID+"/messages",
		map[string]any{"content": "Apa statusnya?", "source_ids": []string{src.ID}})
	if r.Code != http.StatusCreated { t.Fatalf("pesan: %d %s", r.Code, r.Body.String()) }
	var out map[string]any
	if err := json.NewDecoder(r.Body).Decode(&out); err != nil { t.Fatal(err) }
	assistant := out["assistant_message"].(map[string]any)
	if assistant["content"] != "Informasi tidak ditemukan dalam konteks." { t.Fatalf("marker internal bocor ke jawaban: %#v", assistant["content"]) }
	if citations, ok := assistant["citations"].([]any); ok && len(citations) != 0 {
		t.Fatalf("jawaban tidak ditemukan seharusnya tanpa kutipan: %#v", citations)
	}
	list := requestUji(t, h, http.MethodGet, "/api/conversations/"+c.ID+"/messages", nil)
	if list.Code != http.StatusOK { t.Fatalf("riwayat: %d %s", list.Code, list.Body.String()) }
	if strings.Contains(list.Body.String(), noInfoMarker) { t.Fatalf("marker internal tersimpan di riwayat: %s", list.Body.String()) }
}


func TestConversationAcceptsAnswerWithoutCitationAndPersistsContextTitle(t *testing.T) {
	old := os.Getenv("CATATAN_AI_MODEL")
	t.Cleanup(func() { _ = os.Setenv("CATATAN_AI_MODEL", old) })
	_ = os.Setenv("CATATAN_AI_MODEL", "uji-model")

	p := &fakeConversationProvider{response: "Status pengiriman selesai."}
	h := conversationHandler(t, p)
	c := buatPercakapanUji(t, h)
	s := permintaanSumber(t, h, "default", "pengiriman.txt", "Dokumen Pengiriman", "Status: selesai")
	var src Source
	if err := json.NewDecoder(s.Body).Decode(&src); err != nil { t.Fatal(err) }

	r := requestUji(t, h, http.MethodPost, "/api/conversations/"+c.ID+"/messages",
		map[string]any{"content": "Apa statusnya?", "source_ids": []string{src.ID}})
	if r.Code != http.StatusCreated { t.Fatalf("jawaban tanpa kutipan ditolak: %d %s", r.Code, r.Body.String()) }
	var out map[string]any
	if err := json.NewDecoder(r.Body).Decode(&out); err != nil { t.Fatal(err) }
	assistant := out["assistant_message"].(map[string]any)
	if assistant["content"] != "Status pengiriman selesai." { t.Fatalf("jawaban berubah: %#v", assistant["content"]) }
	contexts, ok := assistant["contexts"].([]any)
	if !ok || len(contexts) != 1 { t.Fatalf("label konteks tidak dikembalikan: %#v", assistant["contexts"]) }
	label := contexts[0].(map[string]any)
	if label["title"] != "Dokumen Pengiriman" || label["kind"] != "source" || label["source_id"] != src.ID {
		t.Fatalf("label konteks salah: %#v", label)
	}

	list := requestUji(t, h, http.MethodGet, "/api/conversations/"+c.ID+"/messages", nil)
	if list.Code != http.StatusOK { t.Fatalf("riwayat: %d %s", list.Code, list.Body.String()) }
	if !strings.Contains(list.Body.String(), "Dokumen Pengiriman") { t.Fatalf("label konteks tidak tersimpan dalam riwayat: %s", list.Body.String()) }
}


func TestConversationHistoryUsesInsertionOrderWhenTimestampsAndIDsConflict(t *testing.T) {
	d, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "catatan.db"))
	if err != nil { t.Fatal(err) }
	defer d.Close()

	provider := &fakeConversationProvider{response: "Jawaban pertama"}
	handler := NewWithAI(d, provider)
	setup := requestUji(t, handler, http.MethodPost, "/api/auth/setup", map[string]string{
		"username": "pengguna", "email": "pengguna@lokal.invalid",
		"display_name": "Pengguna Uji", "password": "kata-sandi-uji-aman",
	})
	if setup.Code != http.StatusCreated { t.Fatalf("setup: %d %s", setup.Code, setup.Body.String()) }
	cookie := setup.Result().Cookies()[0]
	authenticated := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.Clone(r.Context())
		r.AddCookie(cookie)
		handler.ServeHTTP(w, r)
	})
	conversation := buatPercakapanUji(t, authenticated)
	response := requestUji(t, authenticated, http.MethodPost, "/api/conversations/"+conversation.ID+"/messages", map[string]any{
		"content": "Pertanyaan pertama",
	})
	if response.Code != http.StatusCreated { t.Fatalf("pesan: %d %s", response.Code, response.Body.String()) }

	// Simulasikan kondisi ketika timestamp pesan sama dan ID acak berlawanan
	// dengan urutan percakapan: pengurutan berdasarkan ID akan membalik peran.
	if _, err := d.ExecContext(context.Background(), `UPDATE messages SET id = CASE role WHEN 'user' THEN 'z-user' ELSE 'a-assistant' END WHERE conversation_id=?`, conversation.ID); err != nil {
		t.Fatalf("atur ID uji: %v", err)
	}
	server := &Server{db: d}
	history, err := server.loadConversationHistory(httptest.NewRequest(http.MethodGet, "/", nil), conversation.ID)
	if err != nil { t.Fatal(err) }
	if len(history) != 2 || history[0].Role != ai.RoleUser || history[1].Role != ai.RoleAssistant {
		t.Fatalf("riwayat harus mempertahankan urutan penyisipan user lalu assistant: %#v", history)
	}

	list := requestUji(t, authenticated, http.MethodGet, "/api/conversations/"+conversation.ID+"/messages", nil)
	if list.Code != http.StatusOK { t.Fatalf("daftar pesan: %d %s", list.Code, list.Body.String()) }
	var messages []ConversationMessage
	if err := json.NewDecoder(list.Body).Decode(&messages); err != nil { t.Fatal(err) }
	if len(messages) != 2 || messages[0].Role != "user" || messages[1].Role != "assistant" {
		t.Fatalf("daftar pesan harus mempertahankan urutan penyisipan: %#v", messages)
	}
}
