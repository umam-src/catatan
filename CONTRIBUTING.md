# Kontribusi

## Bahasa

Kode, nama berkas, komentar, dokumentasi, dan pesan commit menggunakan Bahasa Indonesia baku jika tidak ada alasan kuat untuk mempertahankan istilah asing.

## Prinsip

- Utamakan perbaikan dan optimasi daripada menambah fitur.
- Selama sebelum `v1.0.0`, basis data pengembangan boleh dirombak total; kompatibilitas mundur bukan persyaratan.
- Setelah `v1.0.0`, perubahan basis data harus melalui migrasi yang direncanakan.
- Hindari dependensi besar untuk kebutuhan kecil.
- Pisahkan logika bisnis dari HTTP dan antarmuka.
- Jangan menyimpan rahasia, token, atau data pribadi ke repositori.
- Setiap perubahan perilaku penting harus memiliki pengujian.

## Alur kerja

1. Buat cabang kecil dengan tujuan yang jelas.
2. Baca dokumentasi arsitektur dan skema sebelum mengubah fondasi.
3. Ubah kode secukupnya.
4. Jalankan pemeriksaan lokal.
5. Perbarui dokumentasi jika keputusan atau perilaku berubah.
6. Jika mengubah tabel, kolom, relasi, indeks, atau aturan skema, perbarui `docs/database.md` dan tambahkan atau sesuaikan pengujian migrasi.
7. Periksa ukuran hasil build.
8. Buat commit dengan pesan singkat dan jelas.

## Versi

Gunakan Semantic Versioning dengan kebijakan proyek berikut:

- Penyelesaian satu fase menaikkan **minor**.
- Perubahan kode yang masih berada dalam fase berjalan menaikkan **patch** jika memang menghasilkan rilis patch.
- Perubahan dokumentasi saja tidak menaikkan patch.
- Fase 0 selesai → `0.1.0`.
- Fase 1 selesai → `0.2.0`.
- Fase 2 selesai → `0.3.0`.
- Fase 3 selesai → `0.4.0`.
- Fase 4 selesai → `0.5.0`.
- Fase 5 selesai → `0.6.0`.
- Fase 6 adalah persiapan `1.0.0`, sehingga menjadi kenaikan mayor.

## Pesan commit

Gunakan bentuk:

```text
jenis(area): ringkasan
```

Contoh:

```text
fix(db): cegah migrasi merusak data lama
```

Jenis yang umum: `feat`, `fix`, `docs`, `refactor`, `test`, `build`, `ci`, `perf`.

## Perubahan basis data

Sebelum `v1.0.0`, migrasi eksperimental boleh diubah, dihapus, atau dibuat ulang jika desain baru lebih baik. Data pengembangan boleh dibangun ulang dari nol.

Setelah `v1.0.0`, skema menjadi kontrak stabil: perubahan harus menggunakan migrasi yang terdokumentasi dan diuji terhadap data yang sudah ada.
