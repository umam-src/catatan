# AGENTS.md

## Tujuan

Dokumen ini mencatat pembelajaran dari pekerjaan pengembangan Catatan agar kesalahan yang sama tidak terulang. Isinya menjadi panduan kerja untuk agen dan pengembang.

## Aturan Umum

- Gunakan Obsidian sebagai memori proyek.
- Gunakan Bahasa Indonesia baku.
- Verifikasi kondisi repositori sebelum mengubah berkas.
- Jangan menganggap perubahan berhasil sebelum diverifikasi.
- Jangan mengklaim CI berhasil tanpa memeriksa hasil CI.
- Jangan membuat atau menebak SHA berkas, commit, atau branch.
- Perlakukan perubahan fondasi dan skema basis data sebagai perubahan yang perlu ditinjau dengan cermat.
- Utamakan perubahan kecil, terukur, dan mudah dibalik.
- Utamakan optimasi daripada penambahan fitur.
- Jaga prinsip offline-first.
- Jangan memasukkan data sensitif, rahasia, atau data contoh yang menyerupai data pengguna nyata.

## Pembelajaran

### 1. Jangan memperbarui berkas tanpa SHA yang terverifikasi

**Masalah:** Perubahan berkas GitHub pernah dicoba menggunakan SHA yang belum diverifikasi.

**Risiko:** Perubahan dapat gagal, menimpa versi yang salah, atau membuat status repositori tidak jelas.

**Aturan:** Selalu ambil kondisi terbaru berkas dari branch tujuan dan verifikasi SHA sebelum melakukan pembaruan.

### 2. Jangan menyatakan perubahan berhasil tanpa verifikasi

**Masalah:** Respons alat dapat terlihat berhasil, tetapi kondisi repositori belum tentu sesuai.

**Aturan:**

1. Periksa branch dan commit.
2. Periksa diff atau isi berkas.
3. Jalankan atau periksa pengujian dan CI yang relevan.
4. Baru nyatakan perubahan berhasil.

### 3. Jangan mengandalkan jaringan dari lingkungan kerja lokal

**Pembelajaran:** Lingkungan kerja tidak selalu memiliki akses jaringan keluar.

**Aturan:** Gunakan integrasi GitHub yang tersedia untuk operasi repositori. Gunakan lingkungan lokal hanya untuk operasi yang memang tersedia.

### 4. Perubahan skema harus diperlakukan sebagai perubahan berisiko

**Aturan:** Perubahan migrasi basis data, indeks, dan struktur penyimpanan harus diverifikasi dengan pengujian migrasi dan pengujian perilaku yang terkait.

### 5. Optimasi harus berdasarkan jalur nyata

**Aturan:** Jangan menambahkan cache atau abstraksi hanya karena terlihat lebih cepat. Identifikasi jalur yang benar-benar mahal, ukur bila memungkinkan, lalu buat perubahan terkecil yang menyelesaikan masalah.

### 6. Jangan memperbaiki kesalahan dengan perubahan yang lebih besar

**Pembelajaran:** Ketika operasi perubahan gagal atau statusnya tidak jelas, jangan langsung mengganti seluruh berkas atau melakukan perubahan tambahan untuk menutupinya.

**Aturan:** Hentikan operasi, verifikasi keadaan repositori, lalu lanjutkan dari keadaan yang sudah diketahui benar.

### 7. Pastikan kelas penanda keadaan dipasang pada elemen yang diberi gaya

**Masalah:** Kelas `aktif` dipasang pada pembungkus butir, sedangkan CSS menyasar `.item-buku.aktif` dan `.item-catatan.aktif`. Penanda butir terpilih tidak pernah tampil, dan tidak ada pengujian yang menangkapnya.

**Aturan:** Saat mengubah gaya, render tampilan dan periksa setiap keadaan (aktif, hover, fokus, kosong, galat) pada desktop, ponsel, dan tema gelap. Jangan menganggap selektor CSS berlaku hanya karena berkas lolos pemeriksaan.

## Prosedur Sebelum Commit

- [ ] Pastikan branch benar.
- [ ] Periksa perubahan yang sudah ada.
- [ ] Pastikan perubahan sesuai issue dan ruang lingkup fase.
- [ ] Pastikan tidak ada data sensitif.
- [ ] Jalankan pengujian yang relevan.
- [ ] Periksa diff.
- [ ] Gunakan pesan commit yang jelas.

## Prosedur Setelah Commit

- [ ] Pastikan commit berada di branch yang benar.
- [ ] Periksa isi atau diff commit.
- [ ] Periksa CI.
- [ ] Periksa hasil akhir repositori.
- [ ] Catat pembelajaran baru jika terjadi kesalahan.

## Penempatan Informasi

AGENTS.md hanya untuk pembelajaran dan aturan kerja agen/pengembang yang berasal dari pengalaman proyek. Kebijakan proyek yang sudah stabil sebaiknya ditempatkan pada dokumen yang sesuai, seperti `CONTRIBUTING.md`, `ROADMAP.md`, atau dokumentasi teknis.

Data memory Obsidian dan skill Obsidian bukan bagian dari source tree Catatan dan tidak boleh ditambahkan ke repository hanya untuk mendukung pekerjaan agen.
