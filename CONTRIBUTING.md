# Kontribusi

## Bahasa

Kode, nama berkas, komentar, dokumentasi, dan pesan commit menggunakan Bahasa Indonesia baku jika tidak ada alasan kuat untuk mempertahankan istilah asing.

## Prinsip

- Utamakan perbaikan dan optimasi daripada menambah fitur.
- Jangan merusak data yang sudah ada.
- Hindari dependensi besar untuk kebutuhan kecil.
- Pisahkan logika bisnis dari HTTP dan antarmuka.
- Jangan menyimpan rahasia, token, atau data pribadi ke repositori.
- Setiap perubahan skema harus memiliki migrasi.
- Setiap perubahan perilaku penting harus memiliki pengujian.

## Alur kerja

1. Buat cabang kecil dengan tujuan yang jelas.
2. Baca dokumentasi arsitektur dan skema sebelum mengubah fondasi.
3. Ubah kode secukupnya.
4. Jalankan pemeriksaan lokal.
5. Perbarui dokumentasi jika keputusan atau perilaku berubah.
6. Periksa ukuran hasil build.
7. Buat commit dengan pesan singkat dan jelas.

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

Jangan mengubah migrasi yang sudah dirilis. Tambahkan nomor migrasi berikutnya. Uji terhadap basis data kosong dan basis data yang telah memiliki data.
