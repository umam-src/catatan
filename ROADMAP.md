# Roadmap Buku Catatan

Roadmap ini menjadi arah pengembangan Buku Catatan sampai rilis `v1.0.0` dan setelahnya.

> **Status proyek:** masih dalam pengembangan. Sampai `v1.0.0`, perubahan besar pada kode, struktur aplikasi, API, dan basis data diperbolehkan tanpa mempertahankan kompatibilitas mundur.

## Prinsip

- Utamakan kestabilan, kualitas kode, dan optimasi daripada menambah fitur sebanyak mungkin.
- Offline-first: aplikasi dasar harus tetap berguna tanpa layanan awan.
- Satu program utama dengan antarmuka web yang disematkan.
- Ukuran rilis normal ditargetkan di bawah 20 MB; 30 MB adalah batas peringatan dan 50 MB adalah batas gagal.
- Fitur kecerdasan buatan bersifat tambahan, bukan syarat agar aplikasi dasar dapat digunakan.
- Struktur data harus cukup matang untuk menampung perkembangan fitur tanpa mengunci implementasi pada teknologi tertentu.
- Dokumentasi diperbarui bersama perubahan penting.
- Versi minor naik setiap fase selesai; patch hanya naik bila ada perubahan kode yang dirilis; dokumentasi saja tidak menaikkan patch.

## Tahap 0 — Fondasi `0.x`

**Status: berjalan**

- [x] Struktur proyek Go.
- [x] SQLite tanpa CGO.
- [x] Migrasi basis data awal.
- [x] HTTP server lokal.
- [x] Antarmuka SvelteKit yang disematkan ke program utama.
- [x] Buku dan catatan teks dasar.
- [x] CI dasar untuk pengujian dan pemeriksaan ukuran biner.
- [x] Dokumentasi dasar proyek.
- [x] Glosarium istilah Indonesia.
- [ ] Lisensi MIT.
- [ ] Uji integrasi API dan basis data yang lebih lengkap.
- [ ] Validasi masukan dan batas ukuran permintaan.
- [ ] Penyempurnaan pengalaman pengguna dasar.

### Kebijakan basis data selama `0.x`

Selama belum mencapai `v1.0.0`, basis data **belum dianggap sebagai kontrak stabil**.

- Skema boleh dirombak total.
- Tabel dan kolom boleh diganti nama, digabung, dipisah, atau dihapus.
- Migrasi eksperimental boleh dihapus dan dibuat ulang.
- Data pengembangan tidak wajib dipertahankan ketika desain baru lebih baik.
- Tidak perlu membuat lapisan kompatibilitas hanya untuk mempertahankan desain yang masih terbukti kurang baik.
- Jika diperlukan, basis data dapat dibangun ulang dari nol sebelum `v1.0.0`.

Setelah `v1.0.0`, perubahan basis data harus mengikuti kebijakan migrasi yang lebih konservatif dan terdokumentasi.

## Tahap 1 — Catatan Teks `0.x`

**Tujuan: pengalaman mencatat yang solid sebelum menambah sumber dan AI.**

- [ ] Penyimpanan otomatis dengan debounce.
- [ ] Indikator status penyimpanan.
- [ ] Penghapusan catatan.
- [ ] Pengubahan judul tanpa mengganggu isi.
- [ ] Pencarian catatan lokal.
- [ ] Penyortiran dan pengelompokan yang sederhana.
- [ ] Penanganan kondisi kosong dan kesalahan yang lebih baik.
- [ ] Pintasan papan ketik dasar.
- [ ] Dialog aplikasi menggantikan `prompt()`.
- [ ] Uji perilaku antarmuka utama.

## Tahap 2 — Sumber

**Tujuan: menjadikan buku sebagai tempat mengumpulkan bahan.**

- [ ] Model sumber yang dapat diperluas.
- [ ] Sumber teks lokal.
- [ ] Impor berkas yang ringan dan relevan.
- [ ] Metadata sumber.
- [ ] Penanda lokasi sumber untuk kutipan.
- [ ] Cek integritas atau checksum bila diperlukan.
- [ ] Penghapusan dan pembaruan sumber.
- [ ] Ekspor data buku.
- [ ] Cadangan dan pemulihan lokal.

## Tahap 3 — Pencarian dan Kecerdasan Buatan Lokal

**Tujuan: menambah bantuan AI tanpa membuat aplikasi bergantung pada layanan awan.**

- [ ] Kontrak penyedia model yang sederhana.
- [ ] Dukungan model lokal melalui server yang kompatibel dengan OpenAI.
- [ ] Integrasi opsional dengan `llama-server`.
- [ ] Integrasi opsional dengan Ollama.
- [ ] Percakapan per buku.
- [ ] Pemilihan konteks sumber.
- [ ] Kutipan sumber pada jawaban.
- [ ] Pencarian semantik bila manfaatnya sudah jelas.
- [ ] Konfigurasi penyedia tanpa menyimpan rahasia secara sembarangan.
- [ ] Mode tanpa AI tetap lengkap dan nyaman digunakan.

## Tahap 4 — Adopsi Konsep Open Notebook

**Tujuan: mengambil fitur yang terbukti berguna tanpa menyalin seluruh arsitektur Open Notebook.**

- [ ] Transformasi sumber menjadi keluaran terstruktur.
- [ ] Beberapa jenis transformasi yang dapat diperluas.
- [ ] Konteks percakapan yang lebih terkontrol.
- [ ] Beberapa percakapan dalam satu buku.
- [ ] Dukungan beberapa penyedia/model.
- [ ] Import/ekspor konseptual dengan format yang terdokumentasi.
- [ ] Evaluasi fitur berdasarkan kebutuhan nyata, bukan sekadar kesetaraan fitur.

## Tahap 5 — Media Tambahan

**Hanya dikerjakan jika kebutuhan dan ukuran aplikasi tetap masuk akal.**

- [ ] Dokumen dengan format tambahan.
- [ ] Gambar.
- [ ] Audio.
- [ ] Video.
- [ ] Ekstraksi teks atau metadata dari media.
- [ ] Integrasi multimodal bila manfaatnya jelas.

Fitur media tidak boleh mengorbankan ukuran, kesederhanaan distribusi, atau pengalaman offline-first tanpa alasan yang kuat.

## Tahap 6 — Menuju `v1.0.0`

**Tujuan: membekukan kontrak yang sudah matang.**

- [ ] API inti stabil.
- [ ] Skema basis data stabil.
- [ ] Kebijakan migrasi pasca-`v1.0.0` terdokumentasi.
- [ ] Format ekspor dan impor stabil.
- [ ] Pengujian unit, integrasi, dan alur utama mencukupi.
- [ ] Pemeriksaan keamanan dasar.
- [ ] Tidak ada data sensitif di repositori.
- [ ] CI efisien dan tidak menjalankan pekerjaan yang tidak perlu.
- [ ] Ukuran biner memenuhi batas rilis.
- [ ] README, CONTRIBUTING, CHANGELOG, TODO, dan dokumentasi teknis selaras.
- [ ] Lisensi dan atribusi proyek lengkap.
- [ ] Checklist rilis `v1.0.0` selesai.

## Kebijakan versi per fase

| Fase selesai | Versi target |
|---|---:|
| Tahap 0 — Fondasi | `0.1.0` |
| Tahap 1 — Catatan Teks | `0.2.0` |
| Tahap 2 — Sumber | `0.3.0` |
| Tahap 3 — Pencarian dan AI Lokal | `0.4.0` |
| Tahap 4 — Adopsi Konsep Open Notebook | `0.5.0` |
| Tahap 5 — Media Tambahan | `0.6.0` |
| Tahap 6 — Menuju `v1.0.0` | `1.0.0` |

## Setelah `v1.0.0`

Setelah kontrak `v1.0.0` dibekukan, prioritas bergeser dari eksperimen fondasi menjadi kestabilan:

1. Perbaikan bug dan keamanan.
2. Optimasi kinerja dan ukuran.
3. Perubahan basis data melalui migrasi terencana.
4. Fitur baru hanya jika memberi manfaat nyata.
5. Kompatibilitas data dan format publik dipertahankan sejauh memungkinkan.
6. Perubahan yang mematahkan kontrak harus melalui versi mayor atau rencana migrasi yang jelas.

## Kriteria kualitas setiap tahap

Sebuah tahap dianggap siap dilanjutkan jika:

- data tidak rusak pada alur utama;
- kesalahan ditangani dengan jelas;
- tidak ada rahasia atau data sensitif yang masuk ke repositori;
- aplikasi dasar tetap dapat berjalan secara lokal;
- perubahan tidak menambah dependensi tanpa alasan yang kuat;
- ukuran dan waktu CI tetap terkendali;
- dokumentasi sesuai dengan implementasi aktual.
