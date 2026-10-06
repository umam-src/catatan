# Changelog

## [0.2.5] - 2026-10-06

### Diperbaiki

- Panel Sumber dan Artefak kini dapat dilipat agar ruang dokumen dan catatan lebih luas.
- Tombol untuk membuka kembali panel Sumber dan Artefak tetap tersedia di bagian atas.

## [0.2.4] - 2026-10-06

### Diperbaiki

- Kontrol daftar buku dipindahkan ke bagian atas agar posisi tombol lipat dan buka konsisten.
- Daftar buku yang dilipat kini benar-benar disembunyikan sehingga ruang kerja tidak terdorong ke bawah.
- Tombol untuk membuka kembali daftar buku tetap tersedia di bagian atas.


Semua perubahan penting pada proyek ini dicatat di sini.

Format mengikuti Keep a Changelog dan versi mengikuti Semantic Versioning.

## [0.2.3] - 2026-10-06

### Diperbaiki

- Tombol untuk melipat daftar buku kini benar-benar ditampilkan pada panel buku.
- Nama folder data bawaan kini `catatan` dan nama basis data bawaan kini `catatan.db`.

## [0.2.2] - 2026-10-06

### Diperbaiki

- Daftar buku kini dapat dilipat dan ditampilkan kembali agar ruang kerja dapat diperluas.

## [0.2.1] - 2026-10-06

### Diperbaiki

- Buku bawaan dengan ID `default` kini diterima oleh endpoint catatan, sumber, ekspor, dan cadangan.
- Pengujian memastikan buku bawaan dapat diakses tanpa galat `ID buku tidak valid`.

## [0.2.0] - 2026-10-06

### Diubah

- Otomasi rilis menerbitkan tag melalui GitHub Release setelah pemeriksaan berhasil.

### Ditambahkan

- Alur catatan teks lengkap untuk membuat, membuka, mengubah, memilih, menyimpan otomatis, dan menghapus catatan.
- Penyimpanan otomatis dengan debounce dan perlindungan terhadap penimpaan perubahan terbaru.
- Validasi batas judul dan isi catatan.
- Pengujian alur catatan utama tanpa jaringan.

## [0.1.0] - 2026-10-06

### Ditambahkan

- Fondasi aplikasi satu proses berbasis Go.
- Penyimpanan SQLite lokal dengan migrasi berurutan.
- Skema buku, sumber, catatan, percakapan, dan pesan.
- Antarmuka SvelteKit yang disiapkan untuk ditanam ke program.
- Editor catatan teks dan pengelolaan buku dasar.
- Autosave catatan teks dengan debounce dan indikator status penyimpanan.
- Penghapusan catatan secara lunak.
- Catatan baru dapat dibuat tanpa isi awal.
- Identitas pengguna internal, login lokal, session, logout, dan isolasi data pengguna.
- Hash kata sandi dan pembatasan percobaan login.
- Pengujian basis data, migrasi, API, autentikasi, session, isolasi data, dan validasi masukan.
- Dokumentasi arsitektur, skema data, keputusan desain, glosarium, dan roadmap.
- Lisensi MIT.

### Keamanan

- Session disimpan di basis data menggunakan hash token.
- Cookie session menggunakan HttpOnly dan SameSite=Strict.
- Akses data dibatasi berdasarkan pemilik buku.
- Asal permintaan lintas situs ditolak untuk operasi autentikasi.

### Build

- CI memverifikasi dependensi Go, pengujian, `go vet`, build UI, UI tertanam, build program `catatan`, dan ukuran biner.
- Ukuran biner CI terakhir: 9.793.796 byte.
- Ukuran gzip CI terakhir: 4.226.412 byte.

## [Belum dirilis]

Belum ada perubahan.

### Kebijakan versi

- Penyelesaian satu fase menaikkan versi minor.
- Perubahan kode di dalam fase dapat menaikkan versi patch.
- Perubahan dokumentasi saja tidak menaikkan versi patch.
- Fase 6 menjadi `v1.0.0` sebagai rilis mayor.
