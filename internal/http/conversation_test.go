package http

import (
	"context"
	"encoding/json"
	"net/http"
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
		!strings.Contains(system.Content, "WAJIB diikuti kutipan") ||
		!strings.Contains(system.Content, "[S1:L1]") ||
		!strings.Contains(system.Content, "jangan membuat klaim faktual tanpa dukungan") {
		t.Fatalf("instruksi kutipan sumber harus eksplisit: %#v", system)
	}
}

func TestConversationRejectsInvalidCitationAndUnselectedSource(t *testing.T) {
	old := os.Getenv("CATATAN_AI_MODEL")
	oldDiagnostic := os.Getenv("CATATAN_AI_CITATION_DIAGNOSTIC")
	t.Cleanup(func() { _ = os.Setenv("CATATAN_AI_MODEL", old); _ = os.Setenv("CATATAN_AI_CITATION_DIAGNOSTIC", oldDiagnostic) })
	_ = os.Setenv("CATATAN_AI_MODEL", "uji-model")
	_ = os.Setenv("CATATAN_AI_CITATION_DIAGNOSTIC", "false")

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


func TestConversationCitationDiagnosticMode(t *testing.T) {
	oldModel, modelSet := os.LookupEnv("CATATAN_AI_MODEL")
	oldDiagnostic, diagnosticSet := os.LookupEnv("CATATAN_AI_CITATION_DIAGNOSTIC")
	t.Cleanup(func() {
		if modelSet { _ = os.Setenv("CATATAN_AI_MODEL", oldModel) } else { _ = os.Unsetenv("CATATAN_AI_MODEL") }
		if diagnosticSet { _ = os.Setenv("CATATAN_AI_CITATION_DIAGNOSTIC", oldDiagnostic) } else { _ = os.Unsetenv("CATATAN_AI_CITATION_DIAGNOSTIC") }
	})
	_ = os.Setenv("CATATAN_AI_MODEL", "uji-model")
	_ = os.Setenv("CATATAN_AI_CITATION_DIAGNOSTIC", "true")

	t.Run("jawaban tanpa kutipan dikembalikan dengan status invalid", func(t *testing.T) {
		p := &fakeConversationProvider{response: "Kode pengiriman adalah KJ-7319"}
		h := conversationHandler(t, p)
		c := buatPercakapanUji(t, h)
		s := permintaanSumber(t, h, "default", "bahan.txt", "Bahan", "Kode pengiriman: KJ-7319")
		if s.Code != http.StatusCreated { t.Fatalf("sumber: %d %s", s.Code, s.Body.String()) }
		var src Source
		if err := json.NewDecoder(s.Body).Decode(&src); err != nil { t.Fatal(err) }

		r := requestUji(t, h, http.MethodPost, "/api/conversations/"+c.ID+"/messages", map[string]any{"content": "Apa kodenya?", "source_ids": []string{src.ID}})
		if r.Code != http.StatusCreated { t.Fatalf("mode diagnostik harus mengembalikan jawaban: %d %s", r.Code, r.Body.String()) }
		var out map[string]any
		if err := json.NewDecoder(r.Body).Decode(&out); err != nil { t.Fatal(err) }
		assistant := out["assistant_message"].(map[string]any)
		validation := assistant["citation_validation"].(map[string]any)
		if validation["required"] != true || validation["valid"] != false || validation["diagnostic_mode"] != true {
			t.Fatalf("status validasi diagnostik salah: %#v", validation)
		}
		diagnostics := assistant["context_diagnostics"].(map[string]any)
		if diagnostics["source_count"] != float64(1) || diagnostics["content_bytes"] == float64(0) {
			t.Fatalf("diagnostik konteks sumber salah: %#v", diagnostics)
		}
	})

	t.Run("kutipan dengan referensi dan baris valid diterima", func(t *testing.T) {
		p := &fakeConversationProvider{response: "Kode pengiriman adalah KJ-7319 [S1:L1]."}
		h := conversationHandler(t, p)
		c := buatPercakapanUji(t, h)
		s := permintaanSumber(t, h, "default", "bahan.txt", "Bahan", "Kode pengiriman: KJ-7319")
		if s.Code != http.StatusCreated { t.Fatalf("sumber: %d %s", s.Code, s.Body.String()) }
		var src Source
		if err := json.NewDecoder(s.Body).Decode(&src); err != nil { t.Fatal(err) }

		r := requestUji(t, h, http.MethodPost, "/api/conversations/"+c.ID+"/messages", map[string]any{"content": "Apa kodenya?", "source_ids": []string{src.ID}})
		if r.Code != http.StatusCreated { t.Fatalf("pesan: %d %s", r.Code, r.Body.String()) }
		var out map[string]any
		if err := json.NewDecoder(r.Body).Decode(&out); err != nil { t.Fatal(err) }
		assistant := out["assistant_message"].(map[string]any)
		validation := assistant["citation_validation"].(map[string]any)
		if validation["required"] != true || validation["valid"] != true || validation["diagnostic_mode"] != true {
			t.Fatalf("kutipan valid tidak dikenali: %#v", validation)
		}
		citations := assistant["citations"].([]any)
		if len(citations) != 1 { t.Fatalf("kutipan: %#v", citations) }
	})
}
