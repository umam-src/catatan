package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/umam-src/catatan/internal/db"
)

func TestSusunKutipanMenggabungkanRentangDanMengabaikanYangTidakCocok(t *testing.T) {
	catatan := []conversationNote{{ID: "n1", Potongan: []db.HasilPotongan{
		{BarisAwal: 1, BarisAkhir: 3, Cocok: true},
		{BarisAwal: 3, BarisAkhir: 6, Cocok: true},
		{BarisAwal: 20, BarisAkhir: 22, Cocok: true},
		{BarisAwal: 40, BarisAkhir: 41, Cocok: false},
	}}, {ID: "n2", Potongan: []db.HasilPotongan{{BarisAwal: 5, BarisAkhir: 6, Cocok: true}}}}
	sumber := []conversationSource{
		{ID: "s1", Potongan: []db.HasilPotongan{{BarisAwal: 1, BarisAkhir: 2, Cocok: false}}},
		{ID: "s2", Potongan: []db.HasilPotongan{{BarisAwal: 7, BarisAkhir: 9, Cocok: true}, {BarisAwal: 1, BarisAkhir: 6, Cocok: true}}},
	}
	got := susunKutipan(catatan, sumber)
	want := []Citation{
		{Kind: "note", SourceID: "n1", SourceRef: "N1", StartLine: 1, EndLine: 6},
		{Kind: "note", SourceID: "n1", SourceRef: "N1", StartLine: 20, EndLine: 22},
		{Kind: "note", SourceID: "n2", SourceRef: "N2", StartLine: 5, EndLine: 6},
		{Kind: "source", SourceID: "s2", SourceRef: "S2", StartLine: 1, EndLine: 9},
	}
	if len(got) != len(want) {
		t.Fatalf("kutipan = %#v, ingin %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("kutipan[%d] = %#v, ingin %#v", i, got[i], want[i])
		}
	}
	if k := susunKutipan(nil, nil); k == nil || len(k) != 0 {
		t.Fatalf("tanpa konteks harus berupa daftar kosong non-nil: %#v", k)
	}
}

func TestBersihkanJawaban(t *testing.T) {
	cases := []struct{ nama, masuk, ingin string }{
		{"tanpa penanda", "Status selesai.", "Status selesai."},
		{"kutipan model dibuang", "Status selesai [S1:L2-L3].", "Status selesai."},
		{"kutipan tidak sah dibuang", "A [S9:L1] B [N7:L0] C.", "A B C."},
		{"nomor baris dibuang", "[L1] Kode KJ-7319 [L2-L3]", "Kode KJ-7319"},
		{"penanda lama dibuang", "[TIDAK_DITEMUKAN] Informasi tidak ditemukan dalam konteks.", "Informasi tidak ditemukan dalam konteks."},
		{"hanya kutipan", "[S1:L1]", jawabanTidakDitemukan},
		{"kosong", "  ", jawabanTidakDitemukan},
	}
	for _, c := range cases {
		t.Run(c.nama, func(t *testing.T) {
			if got := bersihkanJawaban(c.masuk); got != c.ingin {
				t.Fatalf("bersihkanJawaban(%q) = %q, ingin %q", c.masuk, got, c.ingin)
			}
		})
	}
}

func TestConversationJawabanModelTidakDitolakKarenaFormatKutipan(t *testing.T) {
	p := &fakeConversationProvider{response: "Kode KJ-7319 [S1:L99] [N7:L1] [L3]"}
	h := conversationHandler(t, p)
	c := buatPercakapanUji(t, h)
	s := permintaanSumber(t, h, "default", "bahan.txt", "Bahan", "Kode pengiriman: KJ-7319")
	var src Source
	if err := json.NewDecoder(s.Body).Decode(&src); err != nil {
		t.Fatal(err)
	}
	r := requestUji(t, h, http.MethodPost, "/api/conversations/"+c.ID+"/messages",
		map[string]any{"content": "Apa kode pengiriman?", "source_ids": []string{src.ID}})
	if r.Code != http.StatusCreated {
		t.Fatalf("pesan: %d %s", r.Code, r.Body.String())
	}
	var out struct {
		Asisten struct {
			Isi     string     `json:"content"`
			Kutipan []Citation `json:"citations"`
		} `json:"assistant_message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Asisten.Isi != "Kode KJ-7319" {
		t.Fatalf("jawaban = %q", out.Asisten.Isi)
	}
	if len(out.Asisten.Kutipan) != 1 || out.Asisten.Kutipan[0].StartLine != 1 || out.Asisten.Kutipan[0].EndLine != 1 {
		t.Fatalf("kutipan harus dari potongan terpakai, bukan dari model: %#v", out.Asisten.Kutipan)
	}
}

func TestConversationDokumenTanpaIsiTidakMemanggilModel(t *testing.T) {
	p := &fakeConversationProvider{response: "tidak boleh dipakai"}
	h := conversationHandler(t, p)
	note := buatCatatanUji(t, h, "Kosong", "")
	c := buatPercakapanCakupan(t, h, cakupanSatuCatatan, note.ID)
	r := requestUji(t, h, http.MethodPost, "/api/conversations/"+c.ID+"/messages",
		map[string]any{"content": "Apa isinya?"})
	if r.Code != http.StatusCreated {
		t.Fatalf("pesan: %d %s", r.Code, r.Body.String())
	}
	if len(p.requests) != 0 {
		t.Fatal("model dipanggil padahal tidak ada bahan konteks")
	}
	if !strings.Contains(r.Body.String(), jawabanTidakDitemukan) {
		t.Fatalf("jawaban tidak ditemukan tidak dikembalikan: %s", r.Body.String())
	}
}
