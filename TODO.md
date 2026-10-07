# TODO

> Rincian pekerjaan berjalan yang diturunkan dari `ROADMAP.md`. Roadmap menentukan arah; TODO memecahnya menjadi pekerjaan yang dapat dikerjakan dan diverifikasi. Rincian teknis yang lebih dalam tetap berada pada issue dan PR.

**Status:** `[~]` sedang dikerjakan · `[ ]` belum dikerjakan

## Aturan kerja

- Kerjakan dari fase aktif sebelum membuka pekerjaan fase berikutnya.
- Setiap pekerjaan besar memiliki issue; implementasi dilakukan secara langsung di `main` untuk perubahan pengembangan yang disepakati.
- Perpindahan fase adalah blocker sampai rilis minor fase tersebut selesai.
- Perubahan kode yang dirilis dalam fase berjalan → patch.
- Penyelesaian fase → minor.
- Dokumentasi saja → tidak menaikkan versi.
- Utamakan perbaikan, pengujian, keamanan, ukuran, dan waktu CI sebelum menambah fitur.
- Jangan mempertahankan kompatibilitas basis data pra-`v1.0.0` jika desain yang lebih baik membutuhkan perubahan besar.

---

## Fase 3 — Pencarian dan AI Lokal → `0.4.0`

**Issue induk:** #5 · **Gerbang rilis:** #13

### 3.1 Pencarian lokal
**Issue:** #83
- [x] Tambahkan API pencarian dan alur pemanggilan dari antarmuka. (#88)
- [x] Integrasikan pencarian lokal pada antarmuka dengan konteks buku dan keadaan hasil. (#89)
- [x] Uji data lebih besar.
- [x] Catat waktu pencarian dan dampak penyimpanan.
- [x] Evaluasi optimasi tambahan hanya jika pengukuran membuktikan perlu; belum ada optimasi tambahan yang dibenarkan.

### 3.2 Kontrak penyedia model
**Issue:** #84
- [ ] Tetapkan antarmuka penyedia model yang minimal.
- [ ] Gunakan pola kompatibel OpenAI tanpa mengikat model data inti.
- [ ] Pisahkan konfigurasi penyedia dari logika aplikasi.
- [ ] Dukung konfigurasi lokal tanpa layanan awan wajib.
- [ ] Jangan menyimpan rahasia penyedia di repository, basis data, atau log.
- [ ] Tangani penyedia tidak tersedia, waktu habis, dan respons tidak sah.
- [ ] Tetapkan batas waktu dan ukuran permintaan/respons.
- [ ] Uji kontrak tanpa model sungguhan.

### 3.3 Integrasi model lokal
**Issue:** #85
- [ ] Pilih integrasi awal berdasarkan manfaat dan biaya pemeliharaan.
- [ ] Dukungan server yang kompatibel dengan OpenAI.
- [ ] Evaluasi integrasi opsional llama-server.
- [ ] Evaluasi integrasi opsional Ollama.
- [ ] Tangani penyedia tidak tersedia tanpa merusak alur non-AI.
- [ ] Uji setidaknya satu jalur model lokal secara menyeluruh.
- [ ] Ukur dampak ukuran program, waktu mulai, dan penggunaan memori.
- [ ] Jangan menjadikan model tertentu sebagai ketergantungan wajib.

### 3.4 Percakapan, konteks, dan kutipan
**Issue:** #86
- [ ] Percakapan terikat pada buku dan pemilik.
- [ ] Dukungan beberapa percakapan dalam satu buku bila terbukti perlu.
- [ ] Pemilihan sumber sebagai konteks.
- [ ] Batas jumlah dan ukuran konteks.
- [ ] Kutipan sumber yang dapat diverifikasi.
- [ ] Tandai/tolak konteks yang tidak cocok dengan checksum sumber.
- [ ] Jangan mengirim sumber yang tidak dipilih ke penyedia model.
- [ ] Uji isolasi dan ketepatan kutipan.
- [ ] Mode tanpa AI tetap lengkap.

### 3.5 Gerbang `0.4.0`
- [ ] #83–#86 selesai.
- [ ] Pencarian lokal stabil dan terukur.
- [ ] AI sepenuhnya opsional.
- [ ] Setidaknya satu penyedia model lokal teruji.
- [ ] Kutipan sumber dapat diverifikasi.
- [ ] Pengujian regresi dan mode tanpa AI lulus.
- [ ] CI hijau.
- [ ] Ukuran dan penggunaan sumber daya diverifikasi.
- [ ] CHANGELOG dan dokumentasi diselaraskan.
- [ ] Gerbang #13 diverifikasi dan rilis `0.4.0` terbit.

---

## Fase 4 — Adopsi Konsep Open Notebook → `0.5.0`

### 4.1 Evaluasi
- [ ] Petakan konsep yang relevan.
- [ ] Tentukan manfaat nyata setiap konsep.
- [ ] Tentukan biaya pemeliharaan.
- [ ] Tolak fitur yang hanya menambah kompleksitas tanpa manfaat.

### 4.2 Transformasi dan konteks
- [ ] Transformasi sumber menjadi keluaran terstruktur.
- [ ] Beberapa jenis transformasi yang dapat diperluas.
- [ ] Konteks percakapan lebih terkontrol.
- [ ] Dukungan beberapa percakapan dalam satu buku.
- [ ] Dukungan beberapa provider/model.

### 4.3 Data dan interoperabilitas
- [ ] Pertahankan interoperabilitas pada tingkat konsep, bukan menyalin skema atau arsitektur Open Notebook.
- [ ] Tentukan format import/ekspor konseptual.
- [ ] Dokumentasikan format.
- [ ] Uji kehilangan data.
- [ ] Pertahankan offline-first.

### 4.4 Gerbang `0.5.0`
- [ ] Fitur yang diadopsi telah dibuktikan manfaatnya.
- [ ] Regresi lulus.
- [ ] Kompleksitas tetap terkendali.
- [ ] CI dan ukuran rilis memenuhi batas.
- [ ] Rilis `0.5.0` terverifikasi.

---

## Fase 5 — Media Tambahan → `0.6.0`

### 5.1 Dukungan media
- [ ] Dokumen dengan format tambahan.
- [ ] Gambar.
- [ ] Audio.
- [ ] Video hanya jika kebutuhan membenarkannya.

### 5.2 Pengelolaan media
- [ ] Validasi tipe berkas.
- [ ] Batas ukuran.
- [ ] Penyimpanan lokal yang efisien.
- [ ] Metadata media.
- [ ] Pratinjau media.
- [ ] Penghapusan dan pemulihan media.

### 5.3 Pemrosesan
- [ ] Ekstraksi teks atau metadata jika manfaatnya jelas.
- [ ] Integrasi multimodal hanya jika manfaatnya jelas.
- [ ] Pastikan fitur media tidak menjadi ketergantungan untuk penggunaan dasar.

### 5.4 Cadangan dan pemulihan
- [ ] Ekspor media.
- [ ] Cadangan media.
- [ ] Pemulihan media.
- [ ] Uji integritas hasil pemulihan.

### 5.5 Gerbang `0.6.0`
- [ ] Media yang dipilih stabil.
- [ ] Ukuran aplikasi masih masuk akal.
- [ ] Offline-first tetap terjaga.
- [ ] Pengujian pemulihan lulus.
- [ ] Rilis `0.6.0` terverifikasi.

---

## Fase 6 — Menuju `1.0.0`

### 6.1 Stabilitas kontrak
- [ ] Bekukan API inti.
- [ ] Bekukan skema basis data.
- [ ] Tetapkan kebijakan migrasi pasca-`v1.0.0`.
- [ ] Stabilkan format ekspor/impor.
- [ ] Dokumentasikan kompatibilitas yang dijanjikan.

### 6.2 Pengujian akhir
- [ ] Unit test mencakup logika inti.
- [ ] Integrasi API mencakup alur utama.
- [ ] Uji basis data dan migrasi.
- [ ] Uji alur UI utama.
- [ ] Uji pencadangan dan pemulihan.
- [ ] Uji penggunaan offline.
- [ ] Uji migrasi dari versi pra-`1.0.0`.

### 6.3 Keamanan dan kualitas
- [ ] Audit API.
- [ ] Audit basis data.
- [ ] Audit dependensi.
- [ ] Audit keamanan.
- [ ] Audit rahasia dan data sensitif.
- [ ] Audit performa.
- [ ] Audit ukuran artefak.
- [ ] Hapus kode/dependensi yang tidak digunakan.

### 6.4 Dokumentasi dan rilis
- [ ] README sesuai perilaku final.
- [ ] CONTRIBUTING lengkap.
- [ ] CHANGELOG lengkap.
- [ ] TODO dan ROADMAP diselaraskan.
- [ ] Dokumentasi teknis lengkap.
- [ ] Lisensi dan atribusi lengkap.
- [ ] Tetapkan kebijakan dukungan.
- [ ] Rilis kandidat bila diperlukan.
- [ ] Verifikasi artefak final.
- [ ] Terbitkan `1.0.0`.

---

## Pemeliharaan setelah `1.0.0`

- [ ] Perbarui dependensi secara terukur.
- [ ] Tinjau keamanan berkala.
- [ ] Tinjau waktu dan biaya CI.
- [ ] Tinjau ukuran artefak.
- [ ] Perbaiki regresi sebelum menambah fitur.
- [ ] Pertahankan jalur migrasi data.
- [ ] Dokumentasikan keputusan penting.
- [ ] Perubahan besar harus memiliki issue, kriteria selesai, dan rencana rilis.

## Catatan desain yang diturunkan dari roadmap lama

Rincian yang sebelumnya berada di `docs/roadmap.md` telah dipindahkan ke TODO agar hanya ada satu roadmap. Butir yang dipertahankan di sini meliputi:

- [ ] Sumber URL diperlakukan sebagai fitur opsional pada fase sumber.
- [ ] Kontrak model mengikuti pola OpenAI-compatible tanpa mengikat data inti pada satu penyedia.
- [ ] Konsep Open Notebook diadopsi secara selektif, bukan sebagai salinan arsitektur.
- [ ] Modul pemrosesan media tetap opsional agar program dasar tetap kecil.
