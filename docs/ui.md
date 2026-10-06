# Fondasi antarmuka

Dokumen ini mencatat aturan dasar antarmuka yang dipakai selama Fase 2. Tujuannya menjaga tampilan tetap konsisten tanpa memperluas fitur.

## Prinsip

- Antarmuka mengikuti rancangan yang telah disetujui; perubahan di luar rancangan harus memiliki alasan yang dapat ditinjau.
- Alur utama tetap dapat digunakan tanpa jaringan.
- Pola yang sama menggunakan ukuran, jarak, dan keadaan visual yang sama.
- Fokus keyboard terlihat jelas.
- Keadaan kosong, galat, dan penyimpanan ditampilkan dengan bahasa yang singkat dan mudah dipahami.
- Tampilan kecil tidak menghilangkan alur utama.
- Tidak menambahkan pustaka antarmuka berat hanya untuk kebutuhan visual.

## Dasar visual

Variabel warna, radius, batas, dan bayangan berada di `web/static/ui.css`. Aset tersebut dimuat sekali dari `web/src/app.html` agar halaman memakai dasar yang sama.

Komponen halaman yang sudah ada menggunakan dasar ini untuk kartu masuk, tombol utama, bidang masukan, navigasi buku, daftar catatan, editor, keadaan galat, dan tampilan kosong.

## Responsif

Pada layar kecil, navigasi buku tetap terlihat di bagian atas dan daftar catatan tetap dapat dibuka. Editor menjadi bagian utama halaman dan tidak memerlukan jaringan tambahan.

## Aksesibilitas dasar

- Elemen interaktif menggunakan kontrol HTML yang sesuai.
- Label formulir tetap tersedia.
- Fokus keyboard tidak disembunyikan.
- Keadaan galat memakai `role="alert"` pada alur yang sudah memilikinya.
- Kontras warna dasar dipertahankan untuk teks dan tindakan utama.

## Batasan Fase 2

Dokumen ini bukan spesifikasi redesign menyeluruh. Audit pengalaman pengguna, penyederhanaan alur besar, dan penyempurnaan visual lanjutan tetap menjadi pekerjaan Fase 3.
