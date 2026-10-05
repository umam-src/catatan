# Buku Catatan

Buku Catatan adalah aplikasi catatan lokal yang mengutamakan privasi, cepat dibuka, mudah dipindahkan, dan dapat berkembang menjadi ruang kerja penelitian berbasis AI.

## Arah proyek

- Satu berkas program untuk menjalankan aplikasi.
- Data pengguna tetap lokal dan mudah dicadangkan.
- Antarmuka bergaya buku catatan modern, terinspirasi pengalaman awal NotebookLM tanpa menyalin identitas visualnya.
- Tahap pertama hanya mendukung teks.
- Struktur data disiapkan agar fitur sumber, percakapan, pencarian semantik, dan integrasi AI dapat ditambahkan tanpa merusak data lama.
- Ukuran rilis ditargetkan di bawah 20 MB; 30 MB masih diterima; 50 MB adalah batas gagal.

## Prinsip teknis

Buku Catatan menggunakan satu program utama berbasis Go dengan basis data SQLite lokal dan antarmuka SvelteKit yang ditanam ke dalam program. SQLite dipilih untuk menghindari layanan basis data terpisah, sedangkan lapisan penyimpanan dipisahkan dari model domain agar format data dapat berkembang tanpa mengunci fitur masa depan.

Struktur konsep mengikuti model **buku → sumber → catatan → percakapan**, yang selaras dengan konsep inti Open Notebook. Open Notebook saat ini memisahkan antarmuka, API, dan SurrealDB serta menggunakan migrasi berurutan; Buku Catatan mengambil prinsip pemisahan tersebut tanpa membawa seluruh beban dependensinya. citeturn0search0turn0search2

## Status

Tahap saat ini: **fondasi aplikasi**.

Fitur awal:

- membuat, mengubah, menghapus, dan memilih buku;
- membuat dan mengubah catatan teks;
- pencarian teks lokal;
- penyimpanan otomatis;
- tema terang/gelap mengikuti sistem;
- pencadangan basis data;
- migrasi skema berurutan;
- API lokal yang tetap menjadi kontrak untuk pengembangan berikutnya.

## Menjalankan dari sumber

Prasyarat:

- Go 1.23+
- Node.js 22+

```bash
go run ./cmd/buku-catatan
```

Untuk pengembangan antarmuka:

```bash
cd web
npm install
npm run build
```

## Rilis satu berkas

Program produksi akan menanam hasil build antarmuka ke dalam berkas eksekusi. Data pengguna berada di luar berkas program sehingga memperbarui program tidak menimpa catatan.

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

- [TODO](TODO.md)
- [CHANGELOG](CHANGELOG.md)
- [CONTRIBUTING](CONTRIBUTING.md)
- [Arsitektur](docs/arsitektur.md)
- [Skema data](docs/skema-data.md)
- [Roadmap](docs/roadmap.md)
- [Keputusan desain](docs/keputusan-desain.md)

## Lisensi

Belum ditetapkan.
