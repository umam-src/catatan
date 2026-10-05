# Changelog

Semua perubahan penting pada proyek ini dicatat di sini.

Format mengikuti Keep a Changelog dan versi mengikuti Semantic Versioning.

## [Belum dirilis]

### Ditambahkan

- Autosave catatan teks dengan debounce dan indikator status penyimpanan.
- Penghapusan catatan secara lunak.
- Catatan baru dapat dibuat tanpa isi awal.
- Pengujian API untuk pembuatan catatan kosong dan penghapusan catatan.
- Pengujian migrasi untuk kolom penghapusan lunak.

- Fondasi aplikasi satu proses berbasis Go.
- Penyimpanan SQLite lokal dengan migrasi berurutan.
- Skema buku, sumber, catatan, percakapan, dan pesan.
- Antarmuka SvelteKit yang disiapkan untuk ditanam ke program.
- Editor catatan teks dan pengelolaan buku dasar.
- Dokumentasi arsitektur, skema data, keputusan desain, dan roadmap.

### Kebijakan versi

- Penyelesaian satu fase menaikkan versi minor.
- Perubahan kode di dalam fase dapat menaikkan versi patch.
- Perubahan dokumentasi saja tidak menaikkan versi patch.
- Fase 6 menjadi `v1.0.0` sebagai rilis mayor.
