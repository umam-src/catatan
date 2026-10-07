# Changelog

## [0.3.0] - 2026-10-07

### Ditambahkan

- Pengelolaan sumber teks lokal dengan checksum, penanda lokasi, ekspor, serta cadangan dan pemulihan lokal.
- Pengujian round-trip cadangan dan pemulihan yang memastikan isi serta metadata sumber tetap utuh.

### Diperbaiki

- Dokumentasi status proyek, sumber, ekspor, dan pemulihan diselaraskan dengan implementasi Fase 2.

## [0.2.9] - 2026-10-07

### Diubah

- Tema antarmuka dirombak agar lebih nyaman: huruf lebih besar (tidak ada teks di bawah 12 px), kontras teks memenuhi 4,5:1, jarak lebih lapang, dan warna dipusatkan pada token di `ui.css`.
- Isi catatan memakai huruf berserif bawaan sistem dengan lebar baris dibatasi sekitar 70 karakter. Tidak ada font yang diunduh.
- Dialog pengaturan dibuat lebih tegas: bayangan berlapis, sudut lebih bulat, sakelar yang jelas, keterangan singkat, dan tombol Selesai.
- Label huruf kapital kecil (BUKU, RUANG KERJA, PREFERENSI) dihapus; judul tampil dengan huruf biasa.
- Tata letak desktop tetap N | C | S/A. Lebar kolom kini diatur lewat token.
- Ponsel didesain ulang: satu panel pada satu waktu (daftar catatan lalu editor layar penuh) dengan tombol kembali, bilah bawah Catatan/Sumber/Artefak, tombol catatan baru yang mengambang, dan Sumber/Artefak sebagai lembar bawah.
- Daftar buku dan panel Sumber/Artefak selalu tertutup saat aplikasi dibuka di ponsel.

### Ditambahkan

- Tema gelap otomatis mengikuti pengaturan perangkat.
- Dialog aplikasi untuk membuat buku, mengubah nama buku atau judul catatan, dan konfirmasi penghapusan. Dialog ini menggantikan dialog bawaan peramban, memakai elemen `<dialog>` modal (fokus terkunci, Esc menutup), dan tampil sebagai lembar bawah di ponsel.

### Diperbaiki

- Penanda buku dan catatan yang sedang dipilih tidak pernah tampil karena kelas `aktif` dipasang pada pembungkus, bukan pada tombol yang diberi gaya.
- Membuka sumber di ponsel tidak lagi tertutup lembar Sumber.
- Tombol lipat daftar buku tidak lagi diposisikan dengan angka tetap di `app.html`; gaya sebaris dipindahkan ke `ui.css`.
- Menu ⋯ muncul saat butir disorot atau terpilih pada perangkat dengan penunjuk, dan selalu tampak pada layar sentuh. Target sentuh minimal 44 px.

## [0.2.8] - 2026-10-06

### Ditambahkan

- Konfirmasi kata sandi pada formulir pembuatan akun.
- Menu pengguna di kanan atas yang menampilkan nama pengguna dan menyediakan aksi keluar serta tempat profil.
- Tombol Pengaturan di bagian bawah kolom buku.
- Tampilan tombol Pengaturan menggunakan ikon yang serasi dengan tombol Keluar sebelumnya.
- Menu pengguna dibuka dari ikon saja; nama pengguna tampil terpisah di sampingnya.
- Dialog pengaturan dengan ukuran yang menyesuaikan layar dan tidak mengikuti panjang isi.
- Sakelar untuk menampilkan atau menyembunyikan opsi pembuatan akun di halaman masuk, tersimpan pada perangkat.

### Diperbaiki

- Ikon huruf pada merek diganti dengan ikon buku catatan.
- Judul buku tetap rata kiri saat drawer buku disembunyikan.
- Versi aplikasi ditampilkan di tengah kolom navigasi.
- Status drawer buku dipertahankan setelah halaman dimuat ulang.
- Status drawer Sumber dan Artefak dipertahankan setelah halaman dimuat ulang.
- Navigasi dan ruang kerja mobile ditata ulang dengan drawer, lembar bawah Sumber/Artefak, dan daftar catatan horizontal.
- Menu konteks buku dan catatan tampil di atas panel lain, dan ikon merek dipusatkan di dalam latarnya.

## [0.2.7] - 2026-10-06

### Diperbaiki

- Menu konteks buku dan catatan tetap tersedia dan kini ditutup saat pengguna mengklik di luar menu.
- Tombol `+` untuk membuat catatan baru tersedia satu kali di header daftar catatan, tanpa menambah tombol berulang pada setiap catatan.

## [0.2.6] - 2026-10-06

### Ditambahkan

- Menu konteks untuk buku dan catatan dengan opsi ubah nama, izin akses yang disiapkan untuk tahap berikutnya, dan penghapusan.
- Pengelolaan nama buku melalui API.
- Penghapusan buku dengan penghapusan turunan catatan, sumber, dan isi terkait.
- Keadaan kosong yang lebih ringkas untuk buku dan catatan.

### Diperbaiki

- Judul buku tetap berada di sisi kiri saat drawer buku dilipat.
- Tombol aksi catatan yang berulang dihapus agar pembuatan catatan tersedia dari keadaan kosong.
- Kontrol lipat panel Sumber dan Artefak dikelola langsung oleh komponen antarmuka tanpa manipulasi DOM global.

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
