package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func buatCatatanUji(t *testing.T, h http.Handler, judul, isi string) Note {
	t.Helper()
	r := requestUji(t, h, http.MethodPost, "/api/notebooks/default/notes", map[string]string{"title": judul, "content": isi})
	if r.Code != http.StatusCreated {
		t.Fatalf("buat catatan: %d %s", r.Code, r.Body.String())
	}
	var n Note
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func buatPercakapanCakupan(t *testing.T, h http.Handler, cakupan, catatanID string) Conversation {
	t.Helper()
	body := map[string]string{"title": "Uji", "scope": cakupan}
	if catatanID != "" {
		body["note_id"] = catatanID
	}
	r := requestUji(t, h, http.MethodPost, "/api/notebooks/default/conversations", body)
	if r.Code != http.StatusCreated {
		t.Fatalf("buat percakapan %s: %d %s", cakupan, r.Code, r.Body.String())
	}
	var c Conversation
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		t.Fatal(err)
	}
	return c
}

type keluaranPesan struct {
	Asisten struct {
		Isi     string         `json:"content"`
		Kutipan []Citation     `json:"citations"`
		Konteks []ContextLabel `json:"contexts"`
	} `json:"assistant_message"`
}

func kirimPesanUji(t *testing.T, h http.Handler, percakapanID string, body map[string]any) (int, keluaranPesan) {
	t.Helper()
	r := requestUji(t, h, http.MethodPost, "/api/conversations/"+percakapanID+"/messages", body)
	var out keluaranPesan
	if r.Code == http.StatusCreated {
		if err := json.NewDecoder(r.Body).Decode(&out); err != nil {
			t.Fatal(err)
		}
	}
	return r.Code, out
}

func promptTerakhir(p *fakeConversationProvider) string {
	m := p.requests[len(p.requests)-1].Messages
	return m[len(m)-1].Content
}

func TestCakupanSemuaCatatanMengambilLintasCatatanTanpaSumber(t *testing.T) {
	p := &fakeConversationProvider{response: "Dua gudang."}
	h := conversationHandler(t, p)
	a := buatCatatanUji(t, h, "Gudang Jakarta", "Status gudang Jakarta: menunggu pemeriksaan.\nPetugas: Rani")
	b := buatCatatanUji(t, h, "Gudang Bandung", "Status gudang Bandung: selesai diperiksa.")
	buatCatatanUji(t, h, "Resep", "campur tepung dan telur")
	src := permintaanSumber(t, h, "default", "g.txt", "Gudang sumber", "status gudang sumber rahasia-sumber")
	if src.Code != http.StatusCreated {
		t.Fatalf("sumber: %d", src.Code)
	}
	c := buatPercakapanCakupan(t, h, cakupanSemuaCatatan, "")
	if c.Scope != cakupanSemuaCatatan {
		t.Fatalf("cakupan = %q", c.Scope)
	}

	kode, out := kirimPesanUji(t, h, c.ID, map[string]any{"content": "status gudang"})
	if kode != http.StatusCreated {
		t.Fatalf("pesan: %d", kode)
	}
	prompt := promptTerakhir(p)
	if !strings.Contains(prompt, "Status gudang Jakarta") || !strings.Contains(prompt, "Status gudang Bandung") {
		t.Fatalf("catatan relevan harus ikut dari seluruh buku: %.400s", prompt)
	}
	if strings.Contains(prompt, "campur tepung") || strings.Contains(prompt, "rahasia-sumber") {
		t.Fatalf("catatan tidak relevan atau Sumber ikut terkirim: %.400s", prompt)
	}
	if len(out.Asisten.Kutipan) != 2 || len(out.Asisten.Konteks) != 2 {
		t.Fatalf("kutipan/konteks = %#v / %#v", out.Asisten.Kutipan, out.Asisten.Konteks)
	}
	ids := map[string]string{}
	for _, k := range out.Asisten.Kutipan {
		if k.Kind != "note" {
			t.Fatalf("kutipan harus berjenis note: %#v", k)
		}
		ids[k.SourceRef] = k.SourceID
	}
	if (ids["N1"] != a.ID && ids["N1"] != b.ID) || (ids["N2"] != a.ID && ids["N2"] != b.ID) || ids["N1"] == ids["N2"] {
		t.Fatalf("rujukan N1/N2 harus menunjuk dua Catatan berbeda: %#v", ids)
	}
}

func TestCakupanSemuaCatatanTanpaHasilTidakMemanggilModel(t *testing.T) {
	p := &fakeConversationProvider{response: "tidak boleh dipakai"}
	h := conversationHandler(t, p)
	buatCatatanUji(t, h, "Resep", "campur tepung dan telur")
	c := buatPercakapanCakupan(t, h, cakupanSemuaCatatan, "")
	kode, out := kirimPesanUji(t, h, c.ID, map[string]any{"content": "kata yang tidak pernah ada"})
	if kode != http.StatusCreated {
		t.Fatalf("pesan: %d", kode)
	}
	if len(p.requests) != 0 {
		t.Fatal("model dipanggil padahal tidak ada potongan relevan")
	}
	if out.Asisten.Isi != jawabanTidakDitemukan || len(out.Asisten.Kutipan) != 0 {
		t.Fatalf("jawaban = %#v", out.Asisten)
	}
}

func TestCakupanSatuCatatanHanyaCatatanItu(t *testing.T) {
	p := &fakeConversationProvider{response: "Kode alfa."}
	h := conversationHandler(t, p)
	a := buatCatatanUji(t, h, "Catatan A", "kode alfa-111 gudang")
	buatCatatanUji(t, h, "Catatan B", "kode beta-222 gudang")
	c := buatPercakapanCakupan(t, h, cakupanSatuCatatan, a.ID)
	if c.NoteID != a.ID {
		t.Fatalf("note_id = %q", c.NoteID)
	}
	kode, out := kirimPesanUji(t, h, c.ID, map[string]any{"content": "kode gudang"})
	if kode != http.StatusCreated {
		t.Fatalf("pesan: %d", kode)
	}
	prompt := promptTerakhir(p)
	if !strings.Contains(prompt, "alfa-111") || strings.Contains(prompt, "beta-222") {
		t.Fatalf("hanya Catatan cakupan yang boleh terkirim: %.300s", prompt)
	}
	if len(out.Asisten.Kutipan) != 1 || out.Asisten.Kutipan[0].SourceID != a.ID {
		t.Fatalf("kutipan = %#v", out.Asisten.Kutipan)
	}
}

func TestCakupanDokumenHanyaSumberTerpilih(t *testing.T) {
	p := &fakeConversationProvider{response: "Isi dokumen."}
	h := conversationHandler(t, p)
	buatCatatanUji(t, h, "Catatan", "status gudang dari catatan rahasia-catatan")
	s1 := permintaanSumber(t, h, "default", "a.txt", "A", "status gudang dokumen-a")
	s2 := permintaanSumber(t, h, "default", "b.txt", "B", "status gudang dokumen-b")
	var a, b Source
	if err := json.NewDecoder(s1.Body).Decode(&a); err != nil {
		t.Fatal(err)
	}
	if err := json.NewDecoder(s2.Body).Decode(&b); err != nil {
		t.Fatal(err)
	}
	c := buatPercakapanCakupan(t, h, cakupanDokumen, "")
	kode, _ := kirimPesanUji(t, h, c.ID, map[string]any{"content": "status gudang", "source_ids": []string{a.ID}})
	if kode != http.StatusCreated {
		t.Fatalf("pesan: %d", kode)
	}
	prompt := promptTerakhir(p)
	if !strings.Contains(prompt, "dokumen-a") || strings.Contains(prompt, "dokumen-b") || strings.Contains(prompt, "rahasia-catatan") {
		t.Fatalf("hanya dokumen terpilih yang boleh terkirim: %.300s", prompt)
	}
}

func TestCakupanMenolakPilihanYangTidakSesuai(t *testing.T) {
	p := &fakeConversationProvider{response: "x"}
	h := conversationHandler(t, p)
	n := buatCatatanUji(t, h, "Catatan", "isi")
	s := permintaanSumber(t, h, "default", "a.txt", "A", "isi")
	var src Source
	if err := json.NewDecoder(s.Body).Decode(&src); err != nil {
		t.Fatal(err)
	}
	semua := buatPercakapanCakupan(t, h, cakupanSemuaCatatan, "")
	satu := buatPercakapanCakupan(t, h, cakupanSatuCatatan, n.ID)
	dokumen := buatPercakapanCakupan(t, h, cakupanDokumen, "")
	cases := []struct {
		nama string
		id   string
		body map[string]any
	}{
		{"semua dengan sumber", semua.ID, map[string]any{"content": "isi", "source_ids": []string{src.ID}}},
		{"semua dengan catatan", semua.ID, map[string]any{"content": "isi", "note_id": n.ID}},
		{"satu dengan sumber", satu.ID, map[string]any{"content": "isi", "source_ids": []string{src.ID}}},
		{"satu dengan catatan", satu.ID, map[string]any{"content": "isi", "note_id": n.ID}},
		{"dokumen dengan catatan", dokumen.ID, map[string]any{"content": "isi", "note_id": n.ID}},
	}
	for _, c := range cases {
		t.Run(c.nama, func(t *testing.T) {
			if kode, _ := kirimPesanUji(t, h, c.id, c.body); kode != http.StatusBadRequest {
				t.Fatalf("kode = %d, ingin 400", kode)
			}
		})
	}
	if len(p.requests) != 0 {
		t.Fatal("model dipanggil untuk pilihan yang ditolak")
	}
}

func TestBuatPercakapanValidasiCakupan(t *testing.T) {
	h := conversationHandler(t, &fakeConversationProvider{response: "x"})
	n := buatCatatanUji(t, h, "Catatan", "isi")
	cases := []struct {
		nama string
		body map[string]string
		kode int
	}{
		{"cakupan tidak dikenal", map[string]string{"scope": "lainnya"}, 400},
		{"satu catatan tanpa id", map[string]string{"scope": cakupanSatuCatatan}, 400},
		{"satu catatan id tidak ada", map[string]string{"scope": cakupanSatuCatatan, "note_id": strings.Repeat("a", 32)}, 404},
		{"id catatan pada cakupan lain", map[string]string{"scope": cakupanDokumen, "note_id": n.ID}, 400},
		{"tanpa cakupan = dokumen", map[string]string{"title": "Lama"}, 201},
	}
	for _, c := range cases {
		t.Run(c.nama, func(t *testing.T) {
			if r := requestUji(t, h, http.MethodPost, "/api/notebooks/default/conversations", c.body); r.Code != c.kode {
				t.Fatalf("kode = %d, ingin %d: %s", r.Code, c.kode, r.Body.String())
			}
		})
	}
	satu := buatPercakapanCakupan(t, h, cakupanSatuCatatan, n.ID)
	r := requestUji(t, h, http.MethodGet, "/api/notebooks/default/conversations", nil)
	var daftar []Conversation
	if err := json.NewDecoder(r.Body).Decode(&daftar); err != nil {
		t.Fatal(err)
	}
	var ditemukan bool
	for _, c := range daftar {
		if c.ID == satu.ID {
			ditemukan = c.Scope == cakupanSatuCatatan && c.NoteID == n.ID
		}
		if c.Scope == "" {
			t.Fatalf("cakupan kosong pada daftar: %#v", c)
		}
	}
	if !ditemukan {
		t.Fatalf("cakupan dan catatan tidak dikembalikan pada daftar: %#v", daftar)
	}
}

func TestHapusPercakapanTidakMenghapusIsiDanRiwayatBaruTerpisah(t *testing.T) {
	p := &fakeConversationProvider{response: "Jawaban."}
	h := conversationHandler(t, p)
	n := buatCatatanUji(t, h, "Catatan", "status gudang aman")
	s := permintaanSumber(t, h, "default", "a.txt", "A", "status gudang dokumen")
	var src Source
	if err := json.NewDecoder(s.Body).Decode(&src); err != nil {
		t.Fatal(err)
	}
	c := buatPercakapanCakupan(t, h, cakupanDokumen, "")
	if kode, _ := kirimPesanUji(t, h, c.ID, map[string]any{"content": "pertanyaan lama gudang", "source_ids": []string{src.ID}}); kode != http.StatusCreated {
		t.Fatalf("pesan: %d", kode)
	}

	if r := requestUji(t, h, http.MethodDelete, "/api/conversations/"+c.ID, nil); r.Code != http.StatusNoContent {
		t.Fatalf("hapus: %d %s", r.Code, r.Body.String())
	}
	if r := requestUji(t, h, http.MethodGet, "/api/conversations/"+c.ID+"/messages", nil); r.Code != http.StatusNotFound {
		t.Fatalf("riwayat percakapan terhapus harus 404: %d", r.Code)
	}
	if r := requestUji(t, h, http.MethodDelete, "/api/conversations/"+c.ID, nil); r.Code != http.StatusNotFound {
		t.Fatalf("hapus ulang harus 404: %d", r.Code)
	}
	if r := requestUji(t, h, http.MethodDelete, "/api/conversations/"+strings.Repeat("0", 32), nil); r.Code != http.StatusNotFound {
		t.Fatalf("ID tak dikenal harus 404: %d", r.Code)
	}
	if r := requestUji(t, h, http.MethodGet, "/api/notebooks/default/notes", nil); !strings.Contains(r.Body.String(), n.ID) {
		t.Fatal("Catatan ikut terhapus")
	}
	if r := requestUji(t, h, http.MethodGet, "/api/notebooks/default/sources", nil); !strings.Contains(r.Body.String(), src.ID) {
		t.Fatal("Sumber ikut terhapus")
	}

	baru := buatPercakapanCakupan(t, h, cakupanDokumen, "")
	if r := requestUji(t, h, http.MethodGet, "/api/conversations/"+baru.ID+"/messages", nil); strings.TrimSpace(r.Body.String()) != "[]" {
		t.Fatalf("percakapan baru harus berriwayat kosong: %s", r.Body.String())
	}
	if kode, _ := kirimPesanUji(t, h, baru.ID, map[string]any{"content": "pertanyaan baru gudang", "source_ids": []string{src.ID}}); kode != http.StatusCreated {
		t.Fatalf("pesan baru: %d", kode)
	}
	for _, m := range p.requests[len(p.requests)-1].Messages {
		if strings.Contains(m.Content, "pertanyaan lama") {
			t.Fatalf("riwayat lama ikut ke prompt: %q", m.Content)
		}
	}
}
