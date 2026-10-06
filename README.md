# Buku Catatan

Buku Catatan adalah aplikasi catatan lokal yang mengutamakan privasi, cepat dibuka, mudah dipindahkan, dan dapat berkembang menjadi ruang kerja penelitian berbasis AI.

## Arah proyek

- Satu berkas program untuk menjalankan aplikasi.
- Data pengguna tetap lokal dan mudah dicadangkan.
- Antarmuka bergaya buku catatan modern, terinspirasi pengalaman awal NotebookLM tanpa menyalin identitas visualnya.
- Tahap pertama hanya mendukung teks.
- Struktur data disiapkan agar sumber, percakapan, pencarian semantik, dan integrasi AI dapat ditambahkan tanpa merusak data lama.
- Ukuran rilis ditargetkan di bawah 20 MB; 30 MB masih diterima; 50 MB adalah batas gagal.

## Arah bahasa dan penamaan

Bahasa Indonesia adalah standar utama proyek. Istilah, nama berkas, nama basis data, dan nama kode baru mengikuti [GLOSARIUM](GLOSARIUM.md) agar satu konsep tidak memiliki banyak nama. Nama resmi teknologi dan kontrak eksternal seperti Go, SQLite, SvelteKit, HTTP, JSON, serta nama berkas yang diwajibkan ekosistem tetap dipertahankan.

## Prinsip

Buku Catatan dirancang sebagai aplikasi lokal yang sederhana. Data disimpan pada perangkat pengguna dan antarmuka menjadi bagian dari program utama, sehingga pengguna tidak perlu menyiapkan layanan tambahan untuk penggunaan dasar.

Struktur konsep mengikuti model **buku → sumber → catatan → percakapan**. Prinsip ini mengambil gagasan yang baik dari Open Notebook, tetapi implementasinya dibuat lebih kecil dan mandiri.

## Status

Tahap saat ini: **fondasi aplikasi**.

Fitur awal:

- membuat dan memilih buku;
- membuat dan mengubah catatan teks;
- penyimpanan lokal;
- migrasi data berurutan;
- layanan lokal untuk pengembangan berikutnya;
- antarmuka responsif.

## Menjalankan dari sumber

Prasyarat:

- Go 1.23+
- Node.js 22+

Jalankan aplikasi:

```bash
go run ./cmd/catatan
```

Untuk memasang dependensi dan membangun antarmuka:

```bash
cd web
npm ci
npm run build
npm run check
```

## Menguji

Pengujian Go mencakup basis data, migrasi, API, autentikasi, session, isolasi data, dan validasi masukan.

```bash
go test ./...
go vet ./...
```

Pemeriksaan antarmuka:

```bash
cd web
npm ci
npm exec -- svelte-kit sync
npm run build
npm run check
```

CI menjalankan pemeriksaan tersebut serta memastikan hasil antarmuka tersedia untuk disematkan ke program utama.

## Membangun program produksi

Dari direktori akar proyek:

```bash
go build -trimpath -ldflags='-s -w' -o catatan ./cmd/catatan
```

Program produksi menanam hasil antarmuka ke dalam berkas eksekusi. Data pengguna berada di luar berkas program sehingga pembaruan program tidak menimpa catatan.

## Ukuran

Ukuran rilis diperiksa di CI. Target:

| Ukuran terkompresi | Status |
|---:|---|
| < 20 MB | Target |
| 20–30 MB | Diterima dengan alasan |
| > 30 MB | Perlu optimasi |
| ≥ 50 MB | Gagal |

Ukuran program tidak menghitung basis data pengguna atau model AI.

## Dokumentasi

- [GLOSARIUM](GLOSARIUM.md)
- [TODO](TODO.md)
- [CHANGELOG](CHANGELOG.md)
- [CONTRIBUTING](CONTRIBUTING.md)
- [Arsitektur](docs/ARSITEKTUR.md)
- [Skema data](docs/SKEMA-DATA.md)
- [Roadmap](ROADMAP.md)
- [Keputusan desain](docs/KEPUTUSAN-DESAIN.md)

## Lisensi

Proyek ini menggunakan lisensi MIT. Lihat berkas `LICENSE` untuk teks lengkap.
