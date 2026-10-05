# TODO

## Fase 1 — Catatan teks

- [ ] Perbaiki alur UI agar tidak memuat ulang daftar catatan berlebihan.
- [ ] Tambahkan pintasan papan ketik untuk pengeditan.
- [ ] Tambahkan pencarian lokal cepat.
- [ ] Tambahkan riwayat perubahan sederhana.
- [ ] Tambahkan ekspor Markdown dan JSON.

## Wajib sebelum rilis awal

- [ ] Bangun UI produksi dan verifikasi satu berkas.
- [ ] Ukur ukuran gzip dan ukuran arsip rilis.
- [ ] Tambahkan pemeriksaan bahwa data lama dapat dibuka setelah pembaruan.
- [ ] Tambahkan pencadangan dan ekspor dari UI.
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

## Gerbang rilis minor — blocker fase berikutnya

Setiap perpindahan fase **wajib melewati rilis minor**. Penyelesaian issue fase saja tidak cukup.

- [ ] Tutup seluruh pekerjaan fase saat ini dan verifikasi kriterianya.
- [ ] Pastikan CI hijau pada perubahan terakhir.
- [ ] Verifikasi ukuran rilis dan pemeriksaan wajib fase.
- [ ] Naikkan versi minor untuk hasil fase, misalnya `0.1.0 → 0.2.0`.
- [ ] Perbarui `CHANGELOG.md` dengan entri versi tersebut.
- [ ] Buat tag rilis yang sesuai, misalnya `v0.2.0`.
- [ ] Verifikasi rilis/tag sebelum fase berikutnya dianggap resmi dimulai.

> **Blocker:** jika rilis minor belum ditetapkan dan diverifikasi, fase berikutnya belum boleh dinyatakan selesai atau diterima.

## Kebijakan versi

- Penyelesaian satu fase menaikkan versi minor.
- Perubahan kode yang dirilis di dalam fase berjalan menaikkan patch.
- Perubahan dokumentasi saja tidak menaikkan patch.
- Target fase: `0.1.0`, `0.2.0`, `0.3.0`, `0.4.0`, `0.5.0`, `0.6.0`, lalu `1.0.0`.
