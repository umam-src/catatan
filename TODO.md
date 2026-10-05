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
- [x] Struktur Go + SvelteKit tetap sederhana.
- [x] Program utama dapat dibangun menjadi satu berkas eksekusi.
- [x] UI web tertanam dalam program Go.
- [x] Server hanya mendengar `127.0.0.1` secara bawaan.
- [x] Verifikasi akhir bahwa distribusi tidak membutuhkan direktori UI terpisah saat runtime.

### 0.2 Basis data
- [x] Model inti menyiapkan buku, sumber, catatan, percakapan, dan pesan untuk perkembangan fase berikutnya.
- [x] SQLite berjalan tanpa CGO.
- [x] Basis data baru dapat dibuat dari nol.
- [x] Migrasi awal tersedia.
- [x] Pengujian integrasi basis data tersedia.
- [x] Pengujian migrasi tersedia.
- [x] Kebijakan perubahan skema pra-`v1.0.0` terdokumentasi.
- [ ] Tinjau kembali indeks dan struktur tabel berdasarkan kebutuhan fase berikutnya.
- [x] Rancang fondasi otorisasi berbasis ACL: User/Group → ACL → Resource.
- [x] Tentukan izin terstandar dan indeks ACL sebelum migrasi implementasi.
- [x] Jangan menambahkan Role atau ABAC tanpa kebutuhan yang terukur.
- [x] Pastikan kegagalan transaksi tidak meninggalkan data setengah tersimpan.

### 0.3 API dan keamanan masukan
**Issue:** #18
- [x] Validasi ID buku dan catatan.
- [x] Validasi masukan judul dan isi.
- [x] Batas ukuran badan permintaan.
- [x] Metode HTTP yang tidak didukung ditangani dengan benar.
- [x] Jalur API tidak membuka akses berkas di luar sumber daya.
- [x] Status HTTP untuk galat utama konsisten.
- [x] Isi catatan tidak ditulis ke log.
- [x] Pengujian masukan kosong, terlalu panjang, ID tidak valid, dan data tidak ditemukan.
- [x] Pemeriksaan awal rahasia dan data pribadi.
- [ ] Lakukan pemeriksaan akhir setelah seluruh perubahan Fase 0 digabung.

### 0.4 UI produksi dan alur dasar
**Issue:** #19
- [x] Bangun UI produksi dan verifikasi hasil tertanam melalui CI.
- [ ] Uji buka buku.
- [ ] Uji daftar catatan.
- [ ] Uji buka catatan.
- [ ] Uji membuat catatan.
- [ ] Uji mengubah dan menyimpan catatan.
- [ ] Uji penghapusan catatan.
- [ ] Pastikan pemuatan ulang data tidak dilakukan berulang tanpa kebutuhan.
- [ ] Pastikan kegagalan layanan tidak menghilangkan data lokal.
- [ ] Pastikan kondisi kosong dan galat ditampilkan dengan jelas.
- [x] Hindari pustaka UI berat tanpa manfaat nyata.

### 0.5 Autentikasi dan identitas
**Issue:** #26 · **Blocker Fase 0**
- [x] Tetapkan identitas internal pengguna yang stabil.
- [x] Pisahkan identitas pengguna dari metode autentikasi.
- [x] Implementasikan login lokal.
- [x] Hash kata sandi dengan metode yang aman.
- [x] Implementasikan session yang aman dan dapat dicabut.
- [x] Implementasikan logout.
- [x] Hubungkan data dengan pemilik pengguna.
- [x] Uji isolasi data antar pengguna.
- [x] Rancang batas penyedia untuk LDAP/AD/Synology dan SSO.
- [x] Pastikan login lokal tidak bergantung pada layanan awan.
- [x] Jangan mencatat password, token, cookie, atau kredensial.
- [x] Uji kasus login gagal dan session kedaluwarsa.

### 0.6 Optimasi dan ukuran
- [x] Inventarisasi dependensi Go dan antarmuka pada manifest proyek.
- [ ] Hapus dependensi yang tidak diperlukan.
- [ ] Evaluasi kontribusi ukuran `modernc.org/sqlite`.
- [ ] Optimalkan aset UI produksi berdasarkan pengukuran.
- [x] Ukur ukuran program hasil build.
- [x] Target normal <20 MiB diterapkan.
- [x] Beri peringatan pada >=30 MiB.
- [x] Gagal pada >=50 MiB.
- [x] Ukur ukuran gzip.
- [x] Catat hasil pengukuran untuk pembanding fase berikutnya: biner 9.793.796 byte dan gzip 4.226.513 byte pada CI terakhir.
- [ ] Ukur arsip rilis setelah format artefak rilis ditetapkan.

### 0.7 CI dan build
- [x] Instalasi npm menggunakan `npm ci`.
- [x] Verifikasi modul Go.
- [x] `go vet ./...`.
- [x] Build dan pemeriksaan UI tertanam.
- [x] Pastikan CI hijau pada commit terakhir.
- [ ] Pisahkan pemeriksaan CI umum dari pekerjaan rilis.
- [ ] Jalankan build rilis hanya saat diperlukan.
- [ ] Pastikan bump patch juga memicu build rilis.
- [ ] Verifikasi artefak rilis.
- [x] Hindari langkah CI duplikat yang tidak diperlukan untuk menghemat menit.

### 0.8 Dokumentasi dan kesiapan rilis
**Issue:** #21
- [x] Lisensi MIT tersedia.
- [x] README diselaraskan dengan kondisi aktual.
- [x] Selaraskan README, TODO, ROADMAP, CHANGELOG, dan CONTRIBUTING pada struktur dan status yang telah diverifikasi.
- [x] Pastikan GLOSARIUM menggunakan istilah Indonesia baku.
- [x] Dokumentasikan kebijakan skema pra-`v1.0.0`.
- [x] Dokumentasikan cara menjalankan.
- [x] Dokumentasikan cara menguji.
- [x] Dokumentasikan cara membangun.
- [x] Pastikan dokumentasi tidak menjanjikan fitur yang belum tersedia.
- [x] Pemeriksaan akhir rahasia dan data sensitif pada sumber kode dan dokumentasi yang dapat dicari.

### 0.9 Gerbang `0.1.0`
**Issue:** #9 dan #10
- [ ] Seluruh butir Fase 0 selesai.
- [x] Seluruh pengujian otomatis lulus pada CI terakhir.
- [x] CI hijau.
- [x] Ukuran artefak terukur dan memenuhi batas.
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
- [x] Membuat catatan kosong.
- [x] Mengubah isi catatan.
- [x] Penyimpanan otomatis dengan debounce.
- [x] Indikator status penyimpanan.
- [ ] Pastikan judul dapat diubah tanpa mengganggu isi.
- [ ] Tangani konflik antara perubahan lama dan perubahan terbaru.
- [ ] Kurangi permintaan penyimpanan yang tidak diperlukan.

### 1.2 Penghapusan dan daftar
- [x] Penghapusan lunak.
- [x] Migrasi `deleted_at`.
- [x] Daftar hanya menampilkan catatan aktif.
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

### 2.1 Model sumber
- [ ] Tentukan entitas sumber.
- [ ] Tentukan hubungan sumber dengan buku.
- [ ] Tentukan hubungan sumber dengan catatan.
- [ ] Tentukan metadata minimum.
- [ ] Tentukan identitas sumber yang stabil.
- [ ] Tambahkan pengujian model dan migrasi.

### 2.2 Sumber lokal
- [ ] Dukungan sumber URL sebagai fitur opsional setelah kebutuhan dan batas keamanan ditetapkan.
- [ ] Tambahkan sumber teks lokal.
- [ ] Impor berkas ringan yang relevan.
- [ ] Validasi tipe berkas.
- [ ] Batasi ukuran impor.
- [ ] Tampilkan pratinjau.
- [ ] Simpan metadata yang diperlukan.

### 2.3 Pengelolaan sumber
- [ ] Ubah metadata sumber.
- [ ] Hapus sumber.
- [ ] Tangani sumber yang hilang atau rusak.
- [ ] Tambahkan checksum bila benar-benar diperlukan.
- [ ] Uji pemulihan sumber dari cadangan.

### 2.4 Kutipan dan ekspor
- [ ] Simpan penanda lokasi sumber untuk kutipan.
- [ ] Tautkan kutipan ke sumber.
- [ ] Ekspor data buku beserta sumber.
- [ ] Cadangkan dan pulihkan sumber.

### 2.5 Gerbang `0.3.0`
- [ ] Seluruh pekerjaan sumber selesai.
- [ ] Pengujian migrasi dan API lulus.
- [ ] CI hijau.
- [ ] Ukuran rilis diverifikasi.
- [ ] CHANGELOG diperbarui.
- [ ] Rilis `0.3.0` terverifikasi.

---

## Fase 3 — Pencarian dan AI Lokal → `0.4.0`

### 3.1 Pencarian lokal
- [ ] Tetapkan indeks pencarian.
- [ ] Indeks catatan.
- [ ] Indeks sumber.
- [ ] Perbarui indeks secara efisien saat data berubah.
- [ ] Uji pencarian pada data kecil dan besar.
- [ ] Ukur waktu pencarian.

### 3.2 Kontrak model
- [ ] Tetapkan kontrak penyedia yang kompatibel dengan pola OpenAI tanpa mengikat model data inti.
- [ ] Tetapkan antarmuka penyedia model sederhana.
- [ ] Pisahkan konfigurasi dari logika aplikasi.
- [ ] Hindari ketergantungan provider tertentu pada model data inti.
- [ ] Jangan menyimpan rahasia provider di repository.

### 3.3 Model lokal
- [ ] Dukungan server yang kompatibel dengan OpenAI.
- [ ] Integrasi opsional `llama-server`.
- [ ] Integrasi opsional Ollama.
- [ ] Tangani provider tidak tersedia.
- [ ] Tetapkan batas penggunaan sumber daya.

### 3.4 Percakapan dan konteks
- [ ] Percakapan per buku.
- [ ] Beberapa percakapan dalam satu buku bila diperlukan.
- [ ] Pemilihan konteks sumber.
- [ ] Batas konteks agar penggunaan sumber daya terkendali.
- [ ] Kutipan sumber pada jawaban.
- [ ] Mode tanpa AI tetap lengkap.

### 3.5 Gerbang `0.4.0`
- [ ] Pencarian lokal stabil.
- [ ] AI sepenuhnya opsional.
- [ ] Pengujian provider dan mode tanpa AI lulus.
- [ ] CI hijau.
- [ ] Ukuran dan penggunaan sumber daya diverifikasi.
- [ ] Rilis `0.4.0` terverifikasi.

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

- [x] Model data awal menyiapkan buku, sumber, catatan, percakapan, dan pesan.
- [ ] Sumber URL diperlakukan sebagai fitur opsional pada fase sumber.
- [ ] Kontrak model mengikuti pola OpenAI-compatible tanpa mengikat data inti pada satu penyedia.
- [ ] Konsep Open Notebook diadopsi secara selektif, bukan sebagai salinan arsitektur.
- [ ] Modul pemrosesan media tetap opsional agar program dasar tetap kecil.
