# TODO

> Rincian pekerjaan berjalan yang diturunkan dari `ROADMAP.md`. Roadmap menentukan arah; TODO memecahnya menjadi pekerjaan yang dapat dikerjakan dan diverifikasi. Rincian teknis yang lebih dalam tetap berada pada issue dan PR.

**Status:** `[x]` selesai · `[~]` sedang dikerjakan · `[ ]` belum dikerjakan

## Aturan kerja

- Kerjakan dari fase aktif sebelum membuka pekerjaan fase berikutnya.
- Setiap pekerjaan besar memiliki issue; implementasi dilakukan melalui PR.
- Perpindahan fase adalah blocker sampai rilis minor fase tersebut selesai.
- Perubahan kode yang dirilis dalam fase berjalan → patch.
- Penyelesaian fase → minor.
- Dokumentasi saja → tidak menaikkan versi.
- Utamakan perbaikan, pengujian, keamanan, ukuran, dan waktu CI sebelum menambah fitur.
- Jangan mempertahankan kompatibilitas basis data pra-`v1.0.0` jika desain yang lebih baik membutuhkan perubahan besar.

---

## Fase 0 — Fondasi → `0.1.0`

**Issue utama:** #16 · **Gerbang:** #9 dan #10 · **PR aktif:** #28

### 0.1 Struktur aplikasi

### 0.2 Basis data
- [ ] Tinjau kembali indeks dan struktur tabel berdasarkan kebutuhan fase berikutnya.

### 0.3 API dan keamanan masukan
**Issue:** #18
- [ ] Lakukan pemeriksaan akhir setelah seluruh perubahan Fase 0 digabung.

### 0.4 UI produksi dan alur dasar
**Issue:** #19
- [ ] Uji buka buku.
- [ ] Uji daftar catatan.
- [ ] Uji buka catatan.
- [ ] Uji membuat catatan.
- [ ] Uji mengubah dan menyimpan catatan.
- [ ] Uji penghapusan catatan.
- [ ] Pastikan pemuatan ulang data tidak dilakukan berulang tanpa kebutuhan.
- [ ] Pastikan kegagalan layanan tidak menghilangkan data lokal.
- [ ] Pastikan kondisi kosong dan galat ditampilkan dengan jelas.

### 0.5 Autentikasi dan identitas
**Issue:** #26 · **Blocker Fase 0**

### 0.6 Optimasi dan ukuran
- [ ] Hapus dependensi yang tidak diperlukan.
- [ ] Evaluasi kontribusi ukuran `modernc.org/sqlite`.
- [ ] Optimalkan aset UI produksi berdasarkan pengukuran.
- [ ] Ukur arsip rilis setelah format artefak rilis ditetapkan.

### 0.7 CI dan build
- [ ] Pisahkan pemeriksaan CI umum dari pekerjaan rilis.
- [ ] Jalankan build rilis hanya saat diperlukan.
- [ ] Pastikan bump patch juga memicu build rilis.
- [ ] Verifikasi artefak rilis.

### 0.8 Dokumentasi dan kesiapan rilis
**Issue:** #21

### 0.9 Gerbang `0.1.0`
**Issue:** #9 dan #10
- [ ] Seluruh butir Fase 0 selesai.
- [ ] Pemeriksaan keamanan akhir lulus setelah seluruh perubahan Fase 0 digabung.
- [ ] CHANGELOG mencatat `0.1.0`.
- [ ] Nomor versi aplikasi/artefak ditetapkan bila berlaku.
- [ ] Buat tag `v0.1.0`.
- [ ] Verifikasi tag dan artefak.
- [ ] Tutup gerbang rilis.
- [ ] Baru setelah itu Fase 1 menjadi fase resmi.

---

## Fase 1 — Catatan Teks → `0.2.0`

**PR terkait:** #27

### 1.1 Editor dan penyimpanan
- [ ] Pastikan judul dapat diubah tanpa mengganggu isi.
- [ ] Tangani konflik antara perubahan lama dan perubahan terbaru.
- [ ] Kurangi permintaan penyimpanan yang tidak diperlukan.

### 1.2 Penghapusan dan daftar
- [ ] Tinjau alur pemulihan atau penghapusan permanen jika memang dibutuhkan.
- [ ] Optimalkan pemuatan daftar catatan.

### 1.3 Pencarian dan pengelompokan
- [ ] Pencarian lokal berdasarkan judul dan isi.
- [ ] Penyortiran sederhana.
- [ ] Pengelompokan yang mudah dipahami.
- [ ] Kondisi tanpa hasil.
- [ ] Pengujian pencarian pada data kosong dan data besar.

### 1.4 Pengalaman penggunaan
- [ ] Pintasan papan ketik dasar.
- [ ] Ganti `prompt()` dengan dialog aplikasi.
- [ ] Penanganan kondisi kosong.
- [ ] Penanganan galat penyimpanan.
- [ ] Uji alur antarmuka utama.

### 1.5 Data pengguna
- [ ] Ekspor catatan ke Markdown.
- [ ] Ekspor data terstruktur ke JSON.
- [ ] Cadangan lokal.
- [ ] Pemulihan cadangan.
- [ ] Uji bahwa ekspor tidak kehilangan isi penting.

### 1.6 Gerbang `0.2.0`
- [ ] Seluruh kriteria Fase 1 terpenuhi.
- [ ] Pengujian regresi lulus.
- [ ] CI hijau.
- [ ] Ukuran rilis diverifikasi.
- [ ] CHANGELOG diperbarui.
- [ ] Rilis `0.2.0` dan tag dibuat.
- [ ] Fase 2 baru dinyatakan resmi setelah rilis terverifikasi.

---

## Fase 2 — Sumber → `0.3.0`

Status implementasi sumber: **selesai**. Issue Fase 2 #33–#39 telah diselesaikan; #40 menjadi pemeriksaan akhir fase.

### 2.0 Antarmuka dan identitas
- [ ] Gunakan satu sumber versi untuk seluruh artefak rilis.
- [ ] Uji versi CLI dan halaman web setelah build tertanam.

### 2.1 Model dan sumber lokal

### 2.2 Pengelolaan sumber

### 2.3 Integritas dan lokasi sumber

### 2.4 Hubungan sumber dan catatan

### 2.5 Ekspor, cadangan, dan pemulihan

### 2.6 Evaluasi format tambahan

### 2.7 Gerbang `0.3.0`
- [ ] Pemeriksaan terpadu #40 selesai.
- [ ] CI hijau pada perubahan terakhir.
- [ ] Ukuran rilis diverifikasi pada perubahan terakhir.
- [ ] CHANGELOG dan dokumentasi akhir diselaraskan.
- [ ] Gerbang rilis #12 diverifikasi sebelum tag `v0.3.0`.

## Fase 3 — Pencarian dan AI Lokal → `0.4.0`

**Issue induk:** #5 · **Gerbang rilis:** #13

### 3.1 Pencarian lokal
**Issue:** #83
- [ ] Tetapkan strategi indeks pencarian berdasarkan kebutuhan nyata.
- [ ] Indeks judul dan isi catatan.
- [ ] Indeks isi sumber lokal.
- [ ] Pastikan hasil terisolasi berdasarkan buku/pemilik.
- [ ] Perbarui indeks secara efisien saat data berubah.
- [ ] Tangani kueri kosong, data kosong, dan hasil tidak ditemukan.
- [ ] Uji data kecil dan data lebih besar.
- [ ] Ukur waktu pencarian dan dampak penyimpanan.
- [ ] Hindari indeks/dependensi yang lebih berat daripada manfaatnya.

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
