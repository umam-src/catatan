# Arsitektur Pemrosesan Sumber dan Komponen Tambahan

Dokumen ini menetapkan arah pemrosesan sumber di Catatan. Tujuannya adalah menjaga aplikasi tetap ringan, luring-terlebih-dahulu, responsif, dan dapat berkembang tanpa menjadikan seluruh kemampuan pemrosesan sebagai bagian wajib dari instalasi dasar.

## Keputusan utama

Catatan mengadopsi konsep ekstraksi teks seperti yang digunakan proyek sejenis, tetapi tidak menyalin implementasinya secara mentah.

Prinsip yang dipilih:

> Berkas asli selalu dipertahankan → pemrosesan dilakukan sebagai pekerjaan latar belakang → hasil menjadi representasi kerja → pencarian, Chat, dan pratinjau memanfaatkannya → setiap hasil tetap dapat ditelusuri kembali ke sumber asli.

Pemrosesan latar belakang merupakan bagian dari desain **Sumber**, bukan fitur khusus AI. Dengan begitu fondasinya tetap berguna sebelum Chat atau AI tersedia.

## Berkas asli selalu dipertahankan

Berkas sumber adalah sumber utama dan tidak boleh digantikan oleh hasil ekstraksi.

Contoh konseptual:

```text
Sumber
└── SOP-Pengadaan.pdf
    ├── berkas asli
    └── hasil pemrosesan
        ├── teks
        ├── metadata
        └── struktur halaman
```

Hasil ekstraksi merupakan representasi kerja untuk kebutuhan berikutnya. Berkas asli tetap diperlukan untuk membuka sumber, pemeriksaan ulang, pratinjau, dan penelusuran kutipan.

## Alur pemrosesan

Ketika pengguna menambahkan berkas:

```text
pengguna menambahkan sumber
            │
            ▼
     ┌──────────────┐
     │ Sumber asli  │
     └──────┬───────┘
            │
       pekerjaan
      latar belakang
            │
            ▼
     ┌──────────────┐
     │  Pemrosesan  │
     │ teks + data  │
     └──────┬───────┘
            │
       ┌────┼────┐
       ▼    ▼    ▼
   Pencarian Chat Pratinjau
```

Pengguna tidak perlu menunggu pemrosesan selesai untuk melanjutkan pekerjaan. Sumber dapat langsung ditampilkan dengan status pemrosesan.

### Status minimum

```text
Ditambahkan
    ↓
Memproses
    ↓
Siap
```

Jalur kegagalan:

```text
Memproses
    ↓
Gagal
    ↓
Coba lagi
```

Detail teknis tidak perlu ditampilkan kepada pengguna kecuali memang membantu menyelesaikan masalah.

## Mengapa pekerjaan latar belakang?

Berkas dapat berukuran besar. PDF 200 halaman, misalnya, tidak seharusnya membuat antarmuka berhenti bekerja selama seluruh proses ekstraksi.

Pola yang dipilih:

```text
Tambah PDF
   ↓
Sumber langsung muncul
Status: Memproses…
   ↓
pengguna tetap dapat bekerja
   ↓
pemrosesan selesai
   ↓
Status: Siap
```

Pendekatan ini juga sesuai dengan luring-terlebih-dahulu: pekerjaan dilakukan secara lokal dan tidak membutuhkan layanan awan.

## Pemrosesan bertingkat

Catatan tidak akan langsung membangun seluruh rangkaian pemrosesan dokumen dan AI. Kemampuan ditambahkan berdasarkan kebutuhan nyata.

### Tingkat awal

```text
PDF/DOCX/TXT
     ↓
ekstraksi teks
     ↓
teks terstruktur
```

Prioritas awal adalah mendapatkan teks dan metadata yang dapat digunakan kembali.

Belum menjadi persyaratan awal:

- embedding;
- basis data vektor;
- RAG kompleks;
- OCR;
- peringkasan oleh AI.

Setelah pencarian lokal dibutuhkan:

```text
teks
  ↓
bagian teks
  ↓
indeks pencarian
  ↓
pencarian lokal
```

Setelah Chat/AI benar-benar dibangun dan kebutuhannya terbukti:

```text
teks
  ↓
bagian teks
  ↓
embedding
  ↓
pengambilan konteks
  ↓
AI
```

Dengan urutan ini, Catatan tidak membayar biaya komputasi AI sebelum memang diperlukan.

## Hubungan dengan sumber asli

Hasil pemrosesan harus mempertahankan hubungan dengan sumber dan lokasinya.

Model konseptual:

```text
Sumber
├── berkas asli
├── metadata
├── hasil ekstraksi
└── bagian teks
    ├── halaman 1
    ├── halaman 2
    ├── halaman 3
    └── ...
```

Jangan hanya menyimpan satu teks tanpa informasi asalnya.

Hal ini penting agar Chat nantinya dapat menghasilkan jawaban dengan rujukan seperti:

> Jawaban ini berasal dari SOP Pengadaan.pdf, halaman 17.

Rujukan tersebut harus dapat membuka sumber pada bagian yang relevan. Nilai utama bukan hanya memiliki teks hasil ekstraksi, tetapi dapat menelusuri kembali setiap hasil ke sumber aslinya.

## Pratinjau

Pratinjau menggunakan sumber asli dan hasil pemrosesan sesuai kebutuhan.

Konsep UI:

```text
Sumber
┌──────────────────────────┐
│ SOP Pengadaan.pdf     ×  │
│ Laporan Q3.pdf         × │
├──────────────────────────┤
│                          │
│ Pratinjau                │
│ halaman / teks asli      │
│                          │
├──────────────────────────┤
│ [Buka]                   │
└──────────────────────────┘
```

Chat tidak harus membaca dan mengolah berkas PDF mentah setiap kali. Chat menggunakan representasi teks yang sudah diproses, sementara pengguna tetap dapat kembali ke sumber asli.

## Pembagian kemampuan

Tidak semua pemroses harus menjadi komponen tambahan. Kemampuan yang ringan dan umum tetap menjadi bagian dari aplikasi utama.

| Kemampuan | Aplikasi utama | Komponen tambahan |
| --- | :---: | :---: |
| TXT → teks | ✅ | |
| Markdown → teks | ✅ | |
| PDF teks → teks | ✅ | |
| Metadata PDF | ✅ | |
| DOCX → teks | ✅ jika dependensi tetap ringan | mungkin |
| PDF hasil pindai → teks | | ✅ |
| OCR gambar | | ✅ |
| Transkripsi audio | | ✅ |
| Video → teks | | ✅ |
| Embedding lokal | | ✅ |
| Model AI lokal | | ✅ |

### Mengapa PDF teks tetap bawaan?

Ekstraksi PDF biasa relatif ringan dibanding OCR. Menjadikannya komponen tambahan hanya untuk mengurangi ukuran akan membuat kebutuhan umum menjadi lebih rumit bagi pengguna.

Menambahkan PDF adalah kebutuhan dasar sumber, bukan kemampuan eksotis. Karena itu PDF teks tetap menjadi kemampuan bawaan selama ukuran dan dependensinya tetap sesuai batas proyek.

## Komponen tambahan

Komponen tambahan ditujukan untuk pekerjaan yang berat, opsional, atau membutuhkan dependensi khusus.

Tiga tingkat yang direkomendasikan:

### Tingkat 1 — bawaan

Pemrosesan dokumen dasar:

- TXT;
- Markdown;
- PDF teks;
- metadata;
- DOCX jika tetap ringan dan layak dipelihara.

### Tingkat 2 — pemrosesan tambahan

Pemrosesan media atau dokumen yang lebih berat:

- OCR;
- gambar;
- audio;
- video.

### Tingkat 3 — AI

Kemampuan kecerdasan lokal:

- embedding;
- model bahasa lokal;
- peringkat ulang hasil;
- transkripsi berbasis model;
- peringkasan.

Komponen tambahan sebaiknya:

- berjalan lokal;
- tidak membutuhkan layanan awan;
- dapat diaktifkan hanya ketika diperlukan;
- versinya dapat diverifikasi;
- kegagalannya tidak membuat aplikasi utama gagal;
- menghasilkan data yang tetap dapat digunakan kembali oleh Catatan.

Jika OCR tidak tersedia, Catatan tetap harus dapat membuka PDF, menampilkan berkas, dan menggunakan teks yang memang tersedia. Hanya kemampuan OCR yang tidak tersedia.

## Desain internal tanpa sistem plugin penuh

Catatan **belum perlu membangun sistem plugin penuh**. Sistem plugin sendiri dapat menjadi beban arsitektur yang lebih besar daripada masalah yang hendak diselesaikan.

Sebagai gantinya, fondasi internal dapat menggunakan pemroses sumber yang dapat dipilih atau didaftarkan berdasarkan jenis sumber.

Contoh konseptual:

```text
Sumber
  ↓
pemroses berdasarkan jenis
  ├── PDFTextProcessor
  ├── TextProcessor
  ├── DocxProcessor
  └── OCRProcessor (opsional)
```

Nama di atas bersifat konseptual dan bukan API yang harus langsung diimplementasikan.

Desain ini memberi ruang agar komponen tambahan nantinya dapat menyediakan pemroses baru tanpa membuat aplikasi utama bergantung pada seluruh alat pemrosesan tersebut.

## OCR dan PDF hasil pindai

PDF dapat berupa:

1. PDF berbasis teks;
2. PDF hasil pindai;
3. PDF campuran yang memiliki teks dan gambar.

OCR tidak menjadi syarat awal.

Strategi awal:

```text
PDF
 ↓
coba ekstraksi teks
 ↓
teks cukup?
 ├─ ya → selesai
 └─ tidak → tandai perlu OCR
```

OCR dapat menjadi pekerjaan latar belakang berikutnya melalui komponen tambahan. Dengan demikian aplikasi utama tetap ringan dan PDF yang tidak membutuhkan OCR tidak ikut menanggung beban pemrosesan tersebut.

## Antrean pekerjaan lokal

Pekerjaan latar belakang pada tahap awal tidak perlu menggunakan sistem antrean yang kompleks.

Gunakan antrean pekerjaan lokal yang sederhana dan dapat dilanjutkan setelah aplikasi dibuka kembali.

Secara konseptual:

```text
Sumber
  ↓
Pekerjaan tertunda
  ↓
Antrean lokal
  ↓
Pemroses
  ↓
Hasil + status
```

Syarat desain awal:

- pekerjaan dapat diketahui statusnya;
- pekerjaan yang gagal dapat dicoba kembali;
- pekerjaan tidak mengunci antarmuka;
- hasil tidak hilang ketika aplikasi ditutup;
- aplikasi dapat melanjutkan pekerjaan yang tertunda setelah dibuka kembali;
- kegagalan satu pekerjaan tidak menghentikan pekerjaan lain;
- pemrosesan tetap dilakukan secara lokal.

Belum perlu membangun orkestrasi pekerjaan yang terdistribusi, layanan terpisah, atau sistem antrean eksternal.

## Batas tanggung jawab

Aplikasi utama bertanggung jawab atas:

- penyimpanan sumber asli;
- status sumber;
- metadata dan hubungan sumber;
- antrean pekerjaan lokal;
- pemrosesan ringan yang umum;
- penyimpanan hasil pemrosesan;
- penelusuran hasil ke sumber asli.

Komponen tambahan bertanggung jawab atas kemampuan yang memang dipisahkan, misalnya OCR atau model AI lokal.

Kegagalan komponen tambahan tidak boleh merusak sumber asli atau data hasil pemrosesan yang sudah tersedia.

## Prinsip pengembangan

1. **Sumber asli tidak pernah diganti oleh hasil ekstraksi.**
2. **Pemrosesan tidak boleh membuat UI menunggu tanpa alasan.**
3. **Pemrosesan lokal menjadi pilihan utama.**
4. **Kemampuan ringan dan umum tetap bawaan.**
5. **Kemampuan berat dan opsional dapat dipisahkan.**
6. **Jangan membuat sistem plugin penuh sebelum ada kebutuhan nyata.**
7. **Hasil pemrosesan harus dapat digunakan kembali.**
8. **Setiap hasil harus dapat ditelusuri ke sumber dan lokasinya.**
9. **Kegagalan komponen tambahan tidak boleh membuat aplikasi utama gagal.**
10. **Optimasi dan pengukuran didahulukan daripada menambah kemampuan.**

## Urutan penerapan

Dokumen ini tidak berarti seluruh kemampuan harus dibangun sekarang. Urutan yang disarankan:

1. model sumber dan penyimpanan berkas asli;
2. status pemrosesan;
3. antrean pekerjaan lokal sederhana;
4. ekstraksi teks PDF dan teks biasa;
5. metadata dan hubungan lokasi teks dengan sumber;
6. pratinjau dan penelusuran ke sumber;
7. indeks pencarian lokal;
8. komponen tambahan untuk OCR atau pemrosesan berat;
9. embedding dan AI lokal setelah kebutuhan terbukti.

Dengan urutan tersebut, setiap tahap memberikan manfaat yang dapat digunakan tanpa mengharuskan tahap berikutnya sudah tersedia.

## Keputusan arsitektur

Catatan mengadopsi arsitektur bertingkat:

```text
                    ┌───────────────────┐
                    │   Berkas sumber   │
                    │       asli        │
                    └─────────┬─────────┘
                              │
                              ▼
                    ┌───────────────────┐
                    │ Pekerjaan lokal   │
                    │ latar belakang    │
                    └─────────┬─────────┘
                              │
                 ┌────────────┴────────────┐
                 ▼                         ▼
          Ekstraksi ringan          Komponen tambahan
                 │                         │
                 ▼                         ▼
          Teks + metadata       OCR / media / AI
                 │                         │
                 └────────────┬────────────┘
                              ▼
                    Representasi terstruktur
                              │
                    ┌─────────┼─────────┐
                    ▼         ▼         ▼
               Pencarian    Chat    Pratinjau
```

**Keputusan:** PDF → teks tetap bawaan. OCR dan pemrosesan berat dirancang sebagai komponen tambahan, tetapi sistem plugin penuh belum dibangun. Fondasi pemroses sumber harus cukup longgar agar pemroses tambahan dapat ditambahkan kemudian tanpa mengubah model sumber atau membuat aplikasi utama bergantung pada seluruh alat tersebut.
