# TODO

## Wajib sebelum rilis awal

- [ ] Perbaiki alur UI agar tidak memuat ulang daftar catatan berlebihan.
- [ ] Tambahkan debounce penyimpanan otomatis.
- [ ] Tambahkan pengujian API dan migrasi.
- [ ] Tambahkan pencarian lokal.
- [ ] Tambahkan ekspor dan pencadangan dari UI.
- [ ] Bangun UI produksi dan verifikasi satu berkas.
- [ ] Ukur ukuran gzip dan ukuran arsip rilis.
- [ ] Tambahkan pemeriksaan bahwa data lama dapat dibuka setelah pembaruan.
- [ ] Tetapkan lisensi.

## Optimasi

- [ ] Pangkas dependensi Go yang tidak perlu.
- [ ] Evaluasi ukuran `modernc.org/sqlite` setelah build pertama.
- [ ] Hindari pustaka UI besar jika komponen lokal Svelte sudah cukup.
- [ ] Aktifkan kompresi aset produksi.
- [ ] Pisahkan fitur berat sebagai modul opsional.

## Keamanan

- [ ] Pastikan layanan hanya mendengar `127.0.0.1` secara bawaan.
- [ ] Batasi ukuran badan permintaan.
- [ ] Validasi jalur dan masukan API.
- [ ] Jangan mencatat isi catatan pengguna ke log.
- [ ] Tambahkan pemeriksaan dependensi.

## Kebijakan versi

- Penyelesaian satu fase menaikkan versi minor.
- Perubahan kode yang dirilis di dalam fase berjalan menaikkan patch.
- Perubahan dokumentasi saja tidak menaikkan patch.
- Target fase: `0.1.0`, `0.2.0`, `0.3.0`, `0.4.0`, `0.5.0`, `0.6.0`, lalu `1.0.0`.
