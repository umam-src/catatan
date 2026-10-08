# Glosarium Buku Catatan

Dokumen ini menjadi **rujukan utama pembakuan istilah** untuk proyek Buku Catatan. Tujuannya mencegah satu konsep memiliki beberapa nama, baik dalam dokumentasi, antarmuka, basis data, maupun kode.

> **Aturan utama:** gunakan istilah Bahasa Indonesia yang tercantum pada kolom **Istilah baku**. Istilah lain hanya digunakan ketika merupakan nama resmi dari teknologi, protokol, pustaka, perintah, atau API eksternal.

## 1. Prinsip pembakuan

1. **Bahasa Indonesia menjadi bahasa utama proyek.** Dokumentasi, teks antarmuka, pesan galat, komentar, nama berkas, nama tabel, nama kolom, serta nama fungsi dan variabel baru menggunakan Bahasa Indonesia jika secara teknis memungkinkan.
2. **Satu konsep, satu istilah.** Jangan memperkenalkan sinonim baru untuk konsep yang sudah memiliki istilah baku.
3. **Nama teknologi resmi tidak diterjemahkan.** Contoh: Go, SQLite, SvelteKit, JavaScript, TypeScript, HTTP, JSON, API, Git, GitHub.
4. **Kontrak eksternal dipertahankan.** Nama yang diwajibkan oleh pustaka, protokol, format data, atau API pihak lain boleh tetap menggunakan istilah aslinya.
5. **Kode baru mengikuti glosarium.** Sebelum membuat nama tabel, kolom, fungsi, variabel, tipe, atau berkas baru, periksa dokumen ini terlebih dahulu.
6. **Perubahan istilah adalah perubahan terencana.** Jika istilah baku perlu diubah, perbarui glosarium terlebih dahulu dan catat dampaknya pada `CHANGELOG.md` bila perubahan memengaruhi pengguna atau data.
7. **Migrasi basis data tidak menghapus sejarah.** Nama lama yang sudah pernah digunakan pada skema yang dirilis tidak diubah sembarangan. Perubahan nama dilakukan melalui migrasi berurutan dan harus menjaga data lama.

## 2. Istilah inti

| Istilah baku | Jangan gunakan sebagai istilah internal | Makna |
|---|---|---|
| Buku | notebook, workspace, project | Wadah utama yang mengelompokkan sumber, catatan, dan percakapan. |
| Catatan | note, memo | Isi yang dibuat atau disimpan pengguna di dalam buku. |
| Sumber | source, bahan | Materi yang menjadi bahan acuan buku, misalnya teks atau dokumen pada tahap berikutnya. |
| Percakapan | conversation, chat | Rangkaian pesan antara pengguna dan sistem dalam konteks sebuah buku. |
| Pesan | message | Satu bagian komunikasi di dalam percakapan. |
| Pengguna | user | Orang yang menggunakan aplikasi. |
| Isi | content, body | Muatan utama teks atau data yang disimpan. |
| Judul | title | Nama yang ditampilkan untuk buku, catatan, sumber, atau percakapan. |
| Uraian | description | Penjelasan singkat mengenai suatu objek. |
| Jenis | type | Pengelompokan atau kategori suatu objek. |
| Peran | role | Kedudukan pengirim pesan dalam percakapan. |
| Metadata | metadata | Data tambahan yang menjelaskan suatu objek. Istilah teknis ini dipertahankan karena merupakan istilah baku lintas teknologi. |
| Pencarian | search | Proses menemukan buku, sumber, catatan, atau isi berdasarkan kriteria tertentu. |
| Penyimpanan | storage | Lapisan yang menyimpan dan mengambil data. |
| Basis data | database, DB | Tempat penyimpanan data terstruktur. |
| Migrasi | migration | Perubahan skema atau data yang diterapkan secara berurutan. |
| Skema | schema | Struktur basis data yang disepakati pada suatu versi. |
| Cadangan | backup | Salinan data untuk pemulihan. |
| Pemulihan | restore | Proses mengembalikan data dari cadangan. |
| Pengaturan | settings, configuration | Pilihan yang mengatur perilaku aplikasi. |
| Penyedia | provider | Komponen atau layanan yang menyediakan kemampuan tertentu, terutama untuk AI. |
| Model | model | Model AI atau model data sesuai konteks. |
| Lampiran | attachment | Berkas yang disertakan pada catatan, sumber, atau pesan. |
| Ekspor | export | Mengeluarkan data ke format yang dapat dipindahkan atau dibaca aplikasi lain. |
| Impor | import | Memasukkan data dari format atau aplikasi lain. |
| Perubahan | change | Modifikasi terhadap kode, data, atau perilaku aplikasi. |
| Rilis | release | Versi aplikasi yang dipublikasikan untuk digunakan. |

## 3. Istilah antarmuka

| Istilah baku | Jangan gunakan | Contoh |
|---|---|---|
| Simpan | Save | `Simpan catatan` |
| Batal | Cancel | `Batal` |
| Buat | Create, New | `Buat buku`, `Buat catatan` |
| Hapus | Delete | `Hapus catatan` |
| Ubah | Edit | `Ubah judul` |
| Pilih | Select | `Pilih buku` |
| Tutup | Close | `Tutup` |
| Cari | Search | `Cari catatan` |
| Muat | Load | `Memuat catatan` |
| Memuat | Loading | `Sedang memuat` |
| Berhasil | Success | `Catatan berhasil disimpan` |
| Gagal | Failed, Error | `Gagal menyimpan catatan` |
| Terhubung | Connected | `Terhubung` |
| Offline | Offline | Dipertahankan karena merupakan istilah teknologi yang lazim dan jelas. |
| Dialog | Popup, Modal | `Dialog pengaturan` |
| Lembar bawah | Bottom sheet | `Lembar bawah Sumber` |
| Bilah bawah | Bottom bar, Tab bar | `Bilah bawah ponsel` |
| Tema gelap | Dark mode | `Tema gelap mengikuti perangkat` |
| Token | Variabel gaya | `Token warna` (nilai gaya bernama di `:root`) |
|

## 4. Istilah kode

Nama kode baru **diutamakan menggunakan Bahasa Indonesia**, dengan aturan penamaan mengikuti gaya bahasa pemrograman yang digunakan.

### 4.1 Go

Gunakan nama yang jelas dan konsisten. Untuk kode domain baru, gunakan istilah Indonesia.

| Konsep | Nama baku yang dianjurkan |
|---|---|
| Buku | `Buku` |
| Catatan | `Catatan` |
| Sumber | `Sumber` |
| Percakapan | `Percakapan` |
| Pesan | `Pesan` |
| Pengguna | `Pengguna` |
| Judul | `Judul` |
| Isi | `Isi` |
| Waktu dibuat | `DibuatPada` |
| Waktu diubah | `DiubahPada` |
| Buat ID | `BuatID` |
| Simpan | `Simpan` |
| Ambil | `Ambil` |
| Hapus | `Hapus` |
| Ubah | `Ubah` |
| Daftar | `Daftar` |
| Cari | `Cari` |
| Buka basis data | `BukaBasisData` |
| Migrasikan | `Migrasikan` |

Nama fungsi pembuka yang diekspor ke paket lain tetap mengikuti kebutuhan Go dan harus memiliki nama yang mudah dipahami. Jangan menerjemahkan nama fungsi dari pustaka pihak ketiga.

### 4.2 Svelte/TypeScript/JavaScript

Nama komponen, fungsi, variabel, dan tipe domain baru menggunakan istilah Indonesia jika tidak dipaksa oleh kerangka kerja atau pustaka.

Contoh:

```ts
let bukuTerpilih;
let catatanAktif;
let sedangMemuat = false;

async function muatBuku() {}
async function muatCatatan() {}
async function simpanCatatan() {}
```

Nama khusus kerangka kerja tetap mengikuti kontraknya. Contoh: `load`, `actions`, `PageData`, dan nama berkas yang diwajibkan SvelteKit tidak diterjemahkan.

## 5. Nama basis data

Mulai dari skema baru, nama tabel dan kolom **diutamakan menggunakan Bahasa Indonesia dengan huruf kecil dan garis bawah (`snake_case`)**.

### 5.1 Pemetaan inti

| Nama lama/rancangan awal | Nama baku ke depan |
|---|---|
| `metadata` | `metadata` |
| `notebooks` | `buku` |
| `sources` | `sumber` |
| `notes` | `catatan` |
| `conversations` | `percakapan` |
| `messages` | `pesan` |
| `id` | `id` |
| `title` | `judul` |
| `description` | `uraian` |
| `archived` | `diarsipkan` |
| `kind` | `jenis` |
| `content` | `isi` |
| `locator` | `lokasi` |
| `checksum` | `cek_sum` |
| `metadata_json` | `metadata_json` |
| `notebook_id` | `buku_id` |
| `note_type` | `jenis_catatan` |
| `conversation_id` | `percakapan_id` |
| `role` | `peran` |
| `created_at` | `dibuat_pada` |
| `updated_at` | `diubah_pada` |

**Catatan penting:** tabel dan kolom pada migrasi `001_init.sql` yang sudah dibuat tidak boleh langsung diganti nama hanya untuk mengikuti glosarium. Karena skema tersebut sudah menjadi sejarah versi awal, perubahan nama harus dilakukan melalui migrasi berikutnya dan diuji terhadap data lama.

### 5.2 Istilah yang sengaja tidak diterjemahkan

Beberapa nama dipertahankan karena merupakan bagian dari kontrak teknologi atau istilah teknis yang sudah mapan:

- `id` — identitas umum yang digunakan luas dan menjadi bagian dari banyak kontrak.
- `metadata` — istilah teknis lintas sistem.
- `JSON` — nama format resmi.
- `HTTP` — nama protokol resmi.
- `API` — singkatan teknis yang sudah baku.
- `SQLite` — nama teknologi.
- `WAL` — nama mekanisme SQLite.

## 6. Nama berkas dan direktori

Nama berkas dan direktori yang dibuat oleh proyek **diutamakan dalam Bahasa Indonesia**.

Contoh yang dianjurkan:

```text
docs/
├── arsitektur.md
├── skema-data.md
├── peta-jalan.md
└── keputusan-desain.md

internal/
├── basisdata/
├── http/
└── antarmuka/
```

Namun, nama yang merupakan kontrak ekosistem tidak diterjemahkan. Contoh:

- `README.md`
- `CHANGELOG.md`
- `CONTRIBUTING.md`
- `.gitignore`
- `.github/`
- `go.mod`
- `go.sum`
- `package.json`
- `package-lock.json`
- berkas khusus SvelteKit seperti `+page.svelte` dan `svelte.config.js`

Pelacakan pekerjaan tidak disimpan sebagai berkas TODO di repository. Pekerjaan berjalan dilacak melalui issue GitHub, sedangkan catatan kerja rinci disimpan di Obsidian.

## 7. Istilah teknologi yang tetap asli

Istilah berikut **tidak perlu diterjemahkan** ketika merujuk pada teknologi atau nama produk:

- Go
- SQLite
- SvelteKit
- Svelte
- TypeScript
- JavaScript
- Vite
- Node.js
- Git
- GitHub
- HTTP
- JSON
- API
- URL
- HTML
- CSS
- WebAssembly
- OpenAI
- NotebookLM
- Open Notebook
- llama.cpp

Jika istilah tersebut dipakai sebagai kata umum dalam kalimat Bahasa Indonesia, bentuk kalimat tetap dibakukan. Contoh: `antarmuka SvelteKit`, bukan membuat nama teknologi baru dalam Bahasa Indonesia.

## 8. Istilah AI

| Istilah baku | Jangan gunakan | Makna |
|---|---|---|
| Kecerdasan buatan | artificial intelligence sebagai istilah utama Bahasa Indonesia | Teknologi kecerdasan buatan secara umum. |
| Model bahasa | language model sebagai istilah utama | Model yang memproses bahasa. |
| Model bahasa besar | large language model, LLM sebagai istilah utama | Model bahasa berukuran besar. |
| Penyematan | embedding | Representasi numerik untuk pencarian atau pengelompokan semantik. |
| Pencarian semantik | semantic search | Pencarian berdasarkan makna, bukan hanya kecocokan kata. |
| Kutipan sumber | citation | Penunjuk sumber yang mendukung jawaban atau isi. |
| Transformasi | transformation | Proses mengubah sumber menjadi bentuk lain, misalnya ringkasan. |
| Perintah | prompt | Masukan yang diberikan kepada model. |
| Konteks | context | Informasi yang diberikan kepada model untuk membantu menghasilkan keluaran. |

Singkatan teknis seperti `AI`, `LLM`, dan `RAG` boleh digunakan setelah istilah Indonesianya diperkenalkan dalam dokumentasi yang sama.

## 9. Aturan untuk kompatibilitas Open Notebook

Buku Catatan boleh mengadopsi gagasan dan data dari Open Notebook, tetapi **istilah internal Buku Catatan tetap mengikuti glosarium ini**.

Pemetaan konseptual yang disepakati:

| Open Notebook | Buku Catatan |
|---|---|
| Notebook | Buku |
| Source | Sumber |
| Note | Catatan |
| Chat / Chat Session | Percakapan |
| Message | Pesan |
| Transformation | Transformasi |
| Provider | Penyedia |
| Embedding | Penyematan |

Jika kelak dibuat impor langsung dari Open Notebook, nama asli dari data sumber boleh dipertahankan di lapisan adaptor agar proses impor tidak kehilangan informasi. Setelah masuk ke model internal, gunakan istilah Buku Catatan.

## 10. Aturan perubahan nama

Perubahan nama dilakukan dengan urutan berikut:

1. Tentukan istilah Indonesia yang baku.
2. Tambahkan atau ubah istilah di `GLOSARIUM.md`.
3. Periksa seluruh kode dan dokumentasi untuk istilah lama.
4. Jika menyangkut basis data, buat migrasi baru; jangan mengubah migrasi yang sudah dirilis.
5. Perbarui pengujian.
6. Perbarui dokumentasi dan `CHANGELOG.md` bila relevan.
7. Jalankan pemeriksaan lokal sebelum membuat komit.
8. Gunakan pesan komit yang menjelaskan perubahan.

## 11. Prioritas jika terjadi benturan

Jika dua aturan bertentangan, gunakan urutan prioritas berikut:

1. Kontrak protokol atau format data resmi.
2. Kontrak pustaka atau kerangka kerja yang tidak dapat diubah.
3. Kompatibilitas data lama.
4. Glosarium ini.
5. Kebiasaan penamaan lokal lainnya.

Dengan demikian, Bahasa Indonesia adalah standar utama proyek, tetapi tidak dipaksakan sampai merusak kompatibilitas atau kontrak teknologi.

## 12. Status dokumen

Glosarium ini berlaku untuk **kode dan dokumentasi baru sejak ditambahkan**. Kode lama boleh dimigrasikan secara bertahap. Migrasi besar-besaran hanya dilakukan jika manfaat pemeliharaan lebih besar daripada risiko perubahan dan biaya migrasinya.
