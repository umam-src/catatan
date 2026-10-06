# Changelog

## [0.2.8] - 2026-10-06

### Ditambahkan

- Menu pengguna di kanan atas dengan identitas pengguna dan aksi Keluar; ruang Profil disiapkan untuk tahap berikutnya.
- Tombol Pengaturan pada kolom Buku dengan dialog yang mengikuti ruang layar dan tidak bergantung pada panjang isi.
- Pengaturan awal Pengguna boleh mendaftar aktif dan disimpan pada perangkat.
- Konfirmasi kata sandi pada alur pembuatan akun.

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
