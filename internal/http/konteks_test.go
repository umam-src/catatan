package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/umam-src/catatan/internal/ai"
	"github.com/umam-src/catatan/internal/db"
)

func TestConversationMengambilPotonganRelevanDariSumberBesar(t *testing.T) {
	p := &fakeConversationProvider{response: "Menunggu pemeriksaan gudang."}
	h := conversationHandler(t, p)
	c := buatPercakapanUji(t, h)

	var isi strings.Builder
	for i := 1; i <= 3000; i++ {
		if i == 1500 {
			isi.WriteString("Kode pengiriman KJ-7319 status menunggu pemeriksaan gudang.\n")
			continue
		}
		fmt.Fprintf(&isi, "Baris pengisi nomor %d tentang topik lain.\n", i)
	}
	if isi.Len() < 100<<10 {
		t.Fatalf("data uji terlalu kecil: %d", isi.Len())
	}
	s := permintaanSumber(t, h, "default", "besar.txt", "Dokumen besar", isi.String())
	if s.Code != http.StatusCreated {
		t.Fatalf("sumber: %d %s", s.Code, s.Body.String())
	}
	var src Source
	if err := json.NewDecoder(s.Body).Decode(&src); err != nil {
		t.Fatal(err)
	}

	r := requestUji(t, h, http.MethodPost, "/api/conversations/"+c.ID+"/messages",
		map[string]any{"content": "Apa status pengiriman KJ-7319?", "source_ids": []string{src.ID}})
	if r.Code != http.StatusCreated {
		t.Fatalf("sumber besar harus dapat dipakai: %d %s", r.Code, r.Body.String())
	}

	var keluar struct {
		Asisten struct {
			Kutipan []Citation `json:"citations"`
		} `json:"assistant_message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&keluar); err != nil {
		t.Fatal(err)
	}
	if len(keluar.Asisten.Kutipan) == 0 || keluar.Asisten.Kutipan[0].StartLine > 1500 || keluar.Asisten.Kutipan[0].EndLine < 1500 {
		t.Fatalf("kutipan harus mencakup baris 1500: %#v", keluar.Asisten.Kutipan)
	}
	messages := p.requests[0].Messages
	last := messages[len(messages)-1]
	if last.Role != ai.RoleUser || !strings.Contains(last.Content, "[L1500] Kode pengiriman KJ-7319") {
		t.Fatalf("potongan relevan tidak terkirim dengan nomor baris asli: %.300s", last.Content)
	}
	if len(last.Content) > anggaranKonteksByte+2<<10 {
		t.Fatalf("konteks melebihi anggaran: %d byte", len(last.Content))
	}
	if strings.Contains(last.Content, "nomor 10 tentang") {
		t.Fatal("bagian tidak relevan ikut terkirim")
	}
}

func TestConversationTanpaBagianRelevanMenandaiKonteksKosong(t *testing.T) {
	p := &fakeConversationProvider{response: "[TIDAK_DITEMUKAN] Informasi tidak ditemukan dalam konteks."}
	h := conversationHandler(t, p)
	c := buatPercakapanUji(t, h)

	s := permintaanSumber(t, h, "default", "kecil.txt", "Kecil", "# Resep\ncampur tepung dan telur")
	var src Source
	if err := json.NewDecoder(s.Body).Decode(&src); err != nil {
		t.Fatal(err)
	}
	// Cakupan eksplisit tanpa kata cocok memakai potongan awal, sehingga konteks tidak kosong.
	r := requestUji(t, h, http.MethodPost, "/api/conversations/"+c.ID+"/messages",
		map[string]any{"content": "Siapa presidennya?", "source_ids": []string{src.ID}})
	if r.Code != http.StatusCreated {
		t.Fatalf("pesan: %d %s", r.Code, r.Body.String())
	}
	messages := p.requests[0].Messages
	if !strings.Contains(messages[len(messages)-1].Content, "[L2] campur tepung dan telur") {
		t.Fatalf("potongan awal tidak dipakai: %.300s", messages[len(messages)-1].Content)
	}
}

func TestTulisIsiKonteksMenandaiCelahDanMenghapusTumpangTindih(t *testing.T) {
	var b strings.Builder
	tulisIsiKonteks(&b, "", true, nil)
	if !strings.Contains(b.String(), "tidak ada bagian yang relevan") {
		t.Fatalf("keluaran kosong = %q", b.String())
	}

	b.Reset()
	tulisIsiKonteks(&b, "", true, []db.HasilPotongan{
		{Urutan: 0, BarisAwal: 1, BarisAkhir: 3, Isi: "satu\ndua\ntiga"},
		{Urutan: 1, BarisAwal: 3, BarisAkhir: 5, Isi: "tiga\nempat\nlima"},
		{Urutan: 4, BarisAwal: 10, BarisAkhir: 11, Isi: "sepuluh\nsebelas"},
	})
	want := "[L1] satu\n[L2] dua\n[L3] tiga\n[L4] empat\n[L5] lima\n...\n[L10] sepuluh\n[L11] sebelas\n"
	if b.String() != want {
		t.Fatalf("keluaran = %q, ingin %q", b.String(), want)
	}
}
