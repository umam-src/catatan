package db

import (
	"context"
	"strings"
	"testing"
	"time"
)

// Set evaluasi pengambilan konteks chat (#124). Seluruh data fiktif.
// Setiap pertanyaan menunjuk catatan yang harus muncul pada k teratas.

type catatanEvaluasi struct{ id, judul, isi string }

type pertanyaanEvaluasi struct {
	tanya    string
	diminta  string // id catatan yang harus ditemukan
	frasaIsi string // frasa yang harus ada pada potongan hasil
}

var catatanSetEvaluasi = []catatanEvaluasi{
	{"e1", "Jadwal Rapat", "# Jadwal Rapat\nRapat mingguan tim logistik setiap Selasa pukul 09.30 di ruang Anggrek.\n\n## Agenda\nEvaluasi keterlambatan kiriman dan pembagian tugas gudang."},
	{"e2", "Resep Sambal", "# Resep Sambal Matah\nIris tipis bawang merah, serai, dan cabai rawit. Tuang minyak kelapa panas lalu aduk bersama perasan jeruk limau."},
	{"e3", "Anggaran Renovasi", "# Anggaran Renovasi Dapur\nTotal anggaran Rp45 juta. Lemari atas Rp12 juta, kompor tanam Rp6 juta, keramik lantai Rp9 juta.\n\n## Catatan\nPembayaran tahap kedua dilakukan setelah pemasangan lemari selesai."},
	{"e4", "Perjalanan Bandung", "# Perjalanan ke Bandung\nBerangkat Sabtu pagi dengan kereta pukul 06.10 dari Gambir. Menginap dua malam di penginapan dekat Dago."},
	{"e5", "Pengaturan Router", "# Pengaturan Router\nGanti kanal Wi-Fi ke 11 bila sinyal tetangga mengganggu. Perbarui firmware setiap tiga bulan."},
	{"e6", "Buku Bacaan", "# Daftar Bacaan\nBuku Laut Bercerita dibaca sampai bab tujuh. Berikutnya Pulang karya Leila, lalu buku esai tentang kota."},
	{"e7", "Perawatan Sepeda", "# Perawatan Sepeda\nRantai dilumasi tiap dua minggu. Tekanan ban belakang 55 psi, ban depan 50 psi. Rem cakram diganti bila pad tipis."},
	{"e8", "Rencana Belajar", "# Rencana Belajar Bahasa Jepang\nTarget hiragana selesai akhir bulan. Latihan menulis dua puluh menit setiap malam memakai kartu kosakata."},
	{"e9", "Pajak Tahunan", "# Pelaporan Pajak\nBatas pelaporan SPT orang pribadi 31 Maret. Siapkan bukti potong dari kantor dan daftar harta."},
	{"e10", "Tanaman Balkon", "# Tanaman Balkon\nPenyiraman monstera dua kali seminggu. Pupuk cair diberikan sebulan sekali, hindari sinar matahari siang."},
}

var pertanyaanSetEvaluasi = []pertanyaanEvaluasi{
	{"Kapan rapat mingguan tim logistik?", "e1", "Selasa"},
	{"Apa saja bahan sambal matah?", "e2", "serai"},
	{"Berapa anggaran kompor tanam untuk renovasi dapur?", "e3", "kompor"},
	{"Kereta ke Bandung berangkat jam berapa?", "e4", "06.10"},
	{"Bagaimana mengatasi gangguan sinyal Wi-Fi di router?", "e5", "kanal"},
	{"Sampai bab berapa buku Laut Bercerita dibaca?", "e6", "bab tujuh"},
	{"Berapa tekanan ban sepeda belakang?", "e7", "55 psi"},
	{"Berapa menit latihan menulis hiragana tiap malam?", "e8", "dua puluh menit"},
	{"Kapan batas pelaporan SPT pajak?", "e9", "31 Maret"},
	{"Seberapa sering menyiram monstera di balkon?", "e10", "dua kali seminggu"},
	{"Kapan pembayaran tahap kedua renovasi?", "e3", "tahap kedua"},
	{"Kapan rantai sepeda perlu dilumasi?", "e7", "dua minggu"},
}

// Pertanyaan sulit memakai sinonim atau kata yang tidak ada di catatan. Hanya dicatat
// (tidak ditegakkan): hasilnya menjadi dasar keputusan perlu tidaknya embedding opsional.
var pertanyaanSulitEvaluasi = []pertanyaanEvaluasi{
	{"Hari apa tim pengiriman barang berkumpul?", "e1", "Selasa"},
	{"Cara membuat bumbu pedas khas Bali?", "e2", "serai"},
	{"Jam berapa moda transportasi rel menuju kota kembang?", "e4", "06.10"},
	{"Seberapa sering merawat kendaraan roda dua tanpa mesin?", "e7", "dua minggu"},
}

// hitungRecallSet mengembalikan jumlah pertanyaan yang catatan dan frasa tujuannya
// ada pada k potongan teratas.
func hitungRecallSet(t testing.TB, d *DB, k int, set []pertanyaanEvaluasi) int {
	t.Helper()
	ctx := context.Background()
	tepat := 0
	for _, q := range set {
		hasil, err := d.CariPotongan(ctx, "local", "default", nil, q.tanya, k, 12<<10)
		if err != nil {
			t.Fatal(err)
		}
		for _, h := range hasil {
			if h.DokumenID == "note:"+q.diminta && h.Cocok && strings.Contains(h.Isi, q.frasaIsi) {
				tepat++
				break
			}
		}
	}
	return tepat
}

func siapkanSetEvaluasi(t testing.TB) *DB {
	t.Helper()
	d, _ := bukaDBUji(t)
	for _, c := range catatanSetEvaluasi {
		if _, err := d.Exec(`INSERT INTO notes(id, notebook_id, title, content, created_at, updated_at) VALUES (?, "default", ?, ?, datetime("now"), datetime("now"))`, c.id, c.judul, c.isi); err != nil {
			t.Fatal(err)
		}
	}
	return d
}

func TestEvaluasiPengambilanRecall(t *testing.T) {
	d := siapkanSetEvaluasi(t)
	total := len(pertanyaanSetEvaluasi)
	r1 := hitungRecallSet(t, d, 1, pertanyaanSetEvaluasi)
	r3 := hitungRecallSet(t, d, 3, pertanyaanSetEvaluasi)
	r5 := hitungRecallSet(t, d, 5, pertanyaanSetEvaluasi)
	t.Logf("recall@1=%d/%d recall@3=%d/%d recall@5=%d/%d", r1, total, r3, total, r5, total)
	// Batas bawah dijaga agar penurunan kualitas pengambilan tertangkap uji.
	if float64(r5) < 0.9*float64(total) {
		t.Fatalf("recall@5 turun: %d/%d (minimal 90%%)", r5, total)
	}
	if r3 < r1 || r5 < r3 {
		t.Fatalf("recall tidak boleh menurun terhadap k: %d, %d, %d", r1, r3, r5)
	}
}

func TestEvaluasiPengambilanPertanyaanSulit(t *testing.T) {
	d := siapkanSetEvaluasi(t)
	t.Logf("pertanyaan sulit: recall@5=%d/%d", hitungRecallSet(t, d, 5, pertanyaanSulitEvaluasi), len(pertanyaanSulitEvaluasi))
}

func TestEvaluasiPengambilanWaktu(t *testing.T) {
	d := siapkanSetEvaluasi(t)
	hitungRecallSet(t, d, 5, pertanyaanSetEvaluasi) // pemanasan: dokumen dipecah saat pertama dipakai
	mulai := time.Now()
	hitungRecallSet(t, d, 5, pertanyaanSetEvaluasi)
	rata := time.Since(mulai) / time.Duration(len(pertanyaanSetEvaluasi))
	t.Logf("rata-rata pengambilan: %v per pertanyaan", rata)
	if rata > 50*time.Millisecond {
		t.Fatalf("pengambilan terlalu lambat: %v per pertanyaan (maksimal 50 ms)", rata)
	}
}
