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

	s := permintaanSumber(t, h, "default", "bahan.txt", "Bahan", "baris satu\\nbaris dua\\nbaris tiga")
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

	p.response = "Jawaban [S1:L1]"
	r = requestUji(t, h, http.MethodPost, "/api/conversations/"+c.ID+"/messages",
		map[string]any{"content": "Tanya", "source_ids": []string{sa.ID}})
	if r.Code != http.StatusCreated { t.Fatalf("pesan kedua: %d %s", r.Code, r.Body.String()) }
	prompt := p.requests[len(p.requests)-1].Messages[1].Content
	if strings.Contains(prompt, "rahasia-b") { t.Fatal("sumber yang tidak dipilih ikut terkirim") }
	_ = sb
}
