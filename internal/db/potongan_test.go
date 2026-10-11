package db

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestPecahTeksMemisahPerHeadingDanMenjagaNomorBaris(t *testing.T) {
	isi := "# Satu\nisi satu\n\n# Dua\nisi dua\nbaris tiga"
	bagian := PecahTeks(isi)
	if len(bagian) != 2 {
		t.Fatalf("jumlah potongan = %d, ingin 2: %#v", len(bagian), bagian)
	}
	if bagian[0].BarisAwal != 1 || bagian[0].BarisAkhir != 2 || bagian[0].Isi != "# Satu\nisi satu" {
		t.Fatalf("potongan 1 = %#v", bagian[0])
	}
	if bagian[1].BarisAwal != 4 || bagian[1].BarisAkhir != 6 {
		t.Fatalf("potongan 2 = %#v", bagian[1])
	}
}

func TestPecahTeksTeksKosongTidakMenghasilkanPotongan(t *testing.T) {
	for _, isi := range []string{"", "\n\n  \n"} {
		if got := PecahTeks(isi); len(got) != 0 {
			t.Fatalf("isi %q menghasilkan %#v", isi, got)
		}
	}
}

func TestPecahTeksMembatasiUkuranDanMemberiTumpangTindih(t *testing.T) {
	var b strings.Builder
	for i := 1; i <= 60; i++ {
		fmt.Fprintf(&b, "baris nomor %02d dengan isi yang cukup panjang untuk menambah ukuran\n", i)
	}
	bagian := PecahTeks(b.String())
	if len(bagian) < 3 {
		t.Fatalf("potongan terlalu sedikit: %d", len(bagian))
	}
	for i, p := range bagian {
		if len(p.Isi) > batasPotonganByte {
			t.Fatalf("potongan %d berukuran %d > %d", i, len(p.Isi), batasPotonganByte)
		}
		if i > 0 && p.BarisAwal > bagian[i-1].BarisAkhir {
			t.Fatalf("tidak ada tumpang tindih antara potongan %d dan %d: %#v %#v", i-1, i, bagian[i-1], p)
		}
	}
	if bagian[len(bagian)-1].BarisAkhir != 60 {
		t.Fatalf("potongan terakhir tidak mencapai baris akhir: %#v", bagian[len(bagian)-1])
	}
}

func TestPecahTeksMemotongBarisPanjangDiBatasRune(t *testing.T) {
	isi := strings.Repeat("é", 2000) // 4000 byte, tanpa spasi
	bagian := PecahTeks(isi)
	if len(bagian) < 3 {
		t.Fatalf("potongan = %d", len(bagian))
	}
	gabung := ""
	for _, p := range bagian {
		if len(p.Isi) > batasPotonganByte {
			t.Fatalf("ukuran %d melebihi batas", len(p.Isi))
		}
		if p.BarisAwal != 1 || p.BarisAkhir != 1 {
			t.Fatalf("rentang baris = %#v", p)
		}
		gabung += p.Isi
	}
	if gabung != isi {
		t.Fatal("isi hilang atau rusak setelah pemotongan baris panjang")
	}
}

func TestEkspresiPengambilan(t *testing.T) {
	if got := ekspresiPengambilan("Apa status pengiriman KJ-7319?"); got != `"status" OR "pengiriman" OR "kj" OR "7319"` {
		t.Fatalf("ekspresi = %q", got)
	}
	if got := ekspresiPengambilan(`apa yang "?" ... `); got != "" {
		t.Fatalf("ekspresi kosong ingin '', dapat %q", got)
	}
	if got := ekspresiPengambilan(`gudang" OR x`); got != `"gudang" OR "or"` {
		t.Fatalf("tanda kutip harus diperlakukan sebagai teks: %q", got)
	}
}

func sisipCatatanUji(t *testing.T, d *DB, id, buku, judul, isi string) {
	t.Helper()
	if _, err := d.Exec(`INSERT INTO notes(id, notebook_id, title, content, created_at, updated_at) VALUES (?, ?, ?, ?, datetime("now"), datetime("now"))`, id, buku, judul, isi); err != nil {
		t.Fatal(err)
	}
}

func TestCariPotonganMengambilYangRelevanDanMenghormatiPemilik(t *testing.T) {
	d, _ := bukaDBUji(t)
	ctx := context.Background()
	if _, err := d.Exec(`INSERT INTO users(id, username, display_name, created_at, updated_at) VALUES ("user-2", "pengguna2", "Pengguna 2", datetime("now"), datetime("now"))`); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Exec(`INSERT INTO notebooks(id, owner_id, title, description, created_at, updated_at) VALUES ("buku-2", "user-2", "Buku Dua", "", datetime("now"), datetime("now"))`); err != nil {
		t.Fatal(err)
	}
	sisipCatatanUji(t, d, "n1", "default", "Pengiriman", "# Pengiriman\nKode pengiriman KJ-7319 menunggu pemeriksaan gudang.")
	sisipCatatanUji(t, d, "n2", "default", "Resep", "# Resep\nCampur tepung dan telur lalu panggang.")
	sisipCatatanUji(t, d, "n3", "buku-2", "Rahasia", "# Rahasia\nStatus pengiriman milik pengguna lain.")

	hasil, err := d.CariPotongan(ctx, "local", "default", nil, "Apa status pengiriman KJ-7319?", 5, 4096)
	if err != nil {
		t.Fatal(err)
	}
	if len(hasil) != 1 || hasil[0].DokumenID != "note:n1" || hasil[0].Jenis != "note" || hasil[0].Judul != "Pengiriman" {
		t.Fatalf("hasil = %#v", hasil)
	}
	if hasil[0].BarisAwal != 1 || hasil[0].BarisAkhir != 2 {
		t.Fatalf("rentang baris = %#v", hasil[0])
	}

	lain, err := d.CariPotongan(ctx, "local", "buku-2", nil, "pengiriman", 5, 4096)
	if err != nil {
		t.Fatal(err)
	}
	if len(lain) != 0 {
		t.Fatalf("buku milik pengguna lain tidak boleh terbaca: %#v", lain)
	}
}

func TestCariPotonganMembatasiCakupanDokumenDanMemakaiPotonganAwal(t *testing.T) {
	d, _ := bukaDBUji(t)
	ctx := context.Background()
	sisipCatatanUji(t, d, "a", "default", "Catatan A", "# Awal\nkata gudang pada catatan A")
	sisipCatatanUji(t, d, "b", "default", "Catatan B", "# Awal\nkata gudang pada catatan B")

	hasil, err := d.CariPotongan(ctx, "local", "default", []string{"note:b"}, "gudang", 5, 4096)
	if err != nil {
		t.Fatal(err)
	}
	if len(hasil) != 1 || hasil[0].DokumenID != "note:b" {
		t.Fatalf("cakupan dokumen tidak dihormati: %#v", hasil)
	}

	// Tanpa istilah yang cocok, cakupan eksplisit tetap menghasilkan potongan awal.
	hasil, err = d.CariPotongan(ctx, "local", "default", []string{"note:a"}, "rangkum seluruhnya", 5, 4096)
	if err != nil {
		t.Fatal(err)
	}
	if len(hasil) != 1 || hasil[0].DokumenID != "note:a" || hasil[0].Relevansi != 0 {
		t.Fatalf("potongan awal = %#v", hasil)
	}

	// Tanpa cakupan eksplisit dan tanpa kecocokan, hasilnya kosong.
	hasil, err = d.CariPotongan(ctx, "local", "default", nil, "zzz tidak ada", 5, 4096)
	if err != nil {
		t.Fatal(err)
	}
	if len(hasil) != 0 {
		t.Fatalf("hasil seharusnya kosong: %#v", hasil)
	}
}

func TestCariPotonganMengikutiSuntingHapusDanBatasAnggaran(t *testing.T) {
	d, _ := bukaDBUji(t)
	ctx := context.Background()
	sisipCatatanUji(t, d, "n", "default", "Uji", "# Satu\nkata lama")

	cari := func(q string) []HasilPotongan {
		t.Helper()
		h, err := d.CariPotongan(ctx, "local", "default", nil, q, 5, 4096)
		if err != nil {
			t.Fatal(err)
		}
		return h
	}
	if len(cari("lama")) != 1 {
		t.Fatal("potongan awal tidak ditemukan")
	}
	if _, err := d.Exec(`UPDATE notes SET content = "# Satu\nkata baru" WHERE id = "n"`); err != nil {
		t.Fatal(err)
	}
	if len(cari("lama")) != 0 || len(cari("baru")) != 1 {
		t.Fatal("potongan tidak mengikuti suntingan")
	}
	if _, err := d.Exec(`UPDATE notes SET deleted_at = datetime("now") WHERE id = "n"`); err != nil {
		t.Fatal(err)
	}
	if len(cari("baru")) != 0 {
		t.Fatal("potongan Catatan terhapus masih ditemukan")
	}
	if n := hitungBaris(t, d, `SELECT COUNT(*) FROM potongan WHERE dokumen_id = 'note:n'`); n != 0 {
		t.Fatalf("potongan yatim = %d", n)
	}

	for i := 0; i < 5; i++ {
		sisipCatatanUji(t, d, fmt.Sprintf("m%d", i), "default", "Banyak", "# Kata\n"+strings.Repeat("anggaran ", 60))
	}
	h, err := d.CariPotongan(ctx, "local", "default", nil, "anggaran", 5, 700)
	if err != nil {
		t.Fatal(err)
	}
	if len(h) < 1 || len(h) > 2 {
		t.Fatalf("anggaran byte tidak membatasi hasil: %d potongan", len(h))
	}
}

func TestMigrasiPotonganMemecahDokumenLamaSaatPertamaDipakai(t *testing.T) {
	d, _ := bukaDBUji(t)
	sisipCatatanUji(t, d, "lama", "default", "Lama", "# A\nisi dokumen lama")
	if n := hitungBaris(t, d, `SELECT COUNT(*) FROM potongan_tertunda`); n != 1 {
		t.Fatalf("dokumen tertunda = %d, ingin 1", n)
	}
	if err := d.IndeksPotonganTertunda(context.Background(), "default"); err != nil {
		t.Fatal(err)
	}
	if n := hitungBaris(t, d, `SELECT COUNT(*) FROM potongan_tertunda`); n != 0 {
		t.Fatalf("antrean belum kosong: %d", n)
	}
	if n := hitungBaris(t, d, `SELECT COUNT(*) FROM potongan WHERE dokumen_id = 'note:lama'`); n != 1 {
		t.Fatalf("potongan = %d, ingin 1", n)
	}
}

func BenchmarkCariPotongan(b *testing.B) {
	d, _ := bukaDBUji(b)
	isi := strings.Repeat("isi catatan dengan kata kunci benchmark dan pencarian offline untuk uji pengambilan\n", 40)
	for i := 0; i < 1000; i++ {
		if _, err := d.Exec(`INSERT INTO notes(id, notebook_id, title, content, created_at, updated_at) VALUES (?, "default", ?, ?, datetime("now"), datetime("now"))`, newIDForBenchmark(i), "Catatan lokal", isi); err != nil {
			b.Fatal(err)
		}
	}
	ctx := context.Background()
	mulai := time.Now()
	if err := d.IndeksPotonganTertunda(ctx, "default"); err != nil {
		b.Fatal(err)
	}
	b.Logf("pemecahan awal 1000 Catatan (±%d KB): %v", len(isi)*1000>>10, time.Since(mulai))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := d.CariPotongan(ctx, "local", "default", nil, "benchmark pengambilan", 8, 12<<10); err != nil {
			b.Fatal(err)
		}
	}
}
