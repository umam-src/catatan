# Format ekspor lokal

Catatan menyediakan ekspor satu buku melalui `GET /api/notebooks/{id}/export`. Ekspor berjalan sepenuhnya luring dan tidak membawa kredensial, sesi, atau data sistem.

## Bentuk

Hasil ekspor adalah JSON UTF-8 dengan struktur utama:

```json
{
  "format": "catatan-export",
  "version": 1,
  "notebook": {},
  "notes": [],
  "sources": []
}
```

`notebook` memuat pengenal, judul, deskripsi, status arsip, dan waktu pembuatan/perubahan. `notes` memuat catatan buku, termasuk isi, jenis, metadata, dan waktu. Catatan yang sudah dihapus secara lunak tetap dipertahankan agar data tidak hilang saat ekspor.

`sources` memuat isi sumber, checksum SHA-256, metadata, lokasi tambahan, serta penanda lokasi baris. Penanda lokasi diurutkan berdasarkan baris awal, baris akhir, lalu pengenal.

## Konsistensi

Data buku, catatan, sumber, dan penanda lokasi diurutkan dengan aturan tetap. Bentuk tanggal dan pengenal yang tersimpan dipertahankan apa adanya sehingga ekspor tidak membuat nilai baru. Isi sumber diverifikasi terhadap checksum sebelum hasil ekspor mulai dikirim.

Ekspor dikirim bertahap untuk menghindari pengumpulan seluruh isi sumber dalam satu struktur di memori. Satu sumber tetap mengikuti batas ukuran sumber yang berlaku pada penyimpanan lokal.

## Kompatibilitas pra-`v1.0.0`

Format `version: 1` adalah format awal dan belum menjadi kontrak kompatibilitas jangka panjang. Sebelum `v1.0.0`, perubahan format dapat dilakukan jika diperlukan. Setiap perubahan harus menaikkan versi format dan memperbarui pengujian serta dokumen ini.

Pemulihan dari format ekspor menjadi bagian pekerjaan cadangan dan pemulihan lokal. Sampai dukungan tersebut tersedia, berkas ekspor diperlakukan sebagai format data terbuka untuk pemeriksaan dan pemindahan, bukan sebagai janji bahwa setiap versi lama selalu dapat diimpor otomatis.

## Batasan

- hanya data buku yang diminta yang diekspor;
- tidak ada akses jaringan;
- tidak ada kata sandi, hash kata sandi, token sesi, atau rahasia autentikasi;
- format ini tidak mencakup berkas biner berat, OCR, audio, video, atau pemrosesan sumber jarak jauh.
