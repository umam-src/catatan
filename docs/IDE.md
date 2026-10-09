# Ide dan Usulan

Dokumen ini menampung usulan yang **belum menjadi keputusan**. Isinya bukan daftar pekerjaan: pekerjaan berjalan dilacak melalui issue GitHub. Usulan yang disetujui dipindahkan ke `KEPUTUSAN-DESAIN.md` atau dokumen teknis terkait, lalu dihapus dari sini.

Semua usulan mengikuti prinsip `AGENTS.md`: perubahan kecil, terukur, mudah dibalik, dan optimasi didahulukan daripada fitur baru. Perubahan skema basis data diperlakukan sebagai perubahan berisiko.

## 1. Kutipan sumber pada percakapan

### 1.1 Kondisi saat ini

Berdasarkan `internal/http/conversation.go` pada `main` (versi 0.3.6):

- Konteks terdiri dari Catatan aktif (`N1`, jika ada) dan Sumber yang dipilih manual oleh pengguna (maksimal 8 sumber, total 64 KiB). Seluruh isinya dikirim ke model dengan penanda baris `[L1]`, `[L2]`, dan seterusnya.
- Format kutipan yang diterima hanya `[S1:L3]` dan `[S1:L3-L5]` untuk Sumber, serta `[N1:L3]` dan `[N1:L3-L5]` untuk Catatan aktif.
- `validateCitations` menolak jawaban (HTTP 422) jika tidak ada kutipan sama sekali, nomor sumber tidak dikenal, rentang baris di luar isi sumber, `end < start`, atau rentang lebih dari 50 baris.
- Satu kutipan yang salah membuat seluruh jawaban dibuang dan tidak disimpan.
- Validasi hanya memeriksa sintaks dan rentang, bukan apakah baris yang dikutip mendukung klaim.
- Perintah sistem meminta model menjawab "informasi tidak ditemukan dalam konteks" tanpa kutipan bila konteks tidak memuat jawaban, tetapi validator tetap menolak jawaban tanpa kutipan. Keduanya bertentangan.

### 1.2 Dugaan penyebab kegagalan

Dugaan dari pembacaan kode; **belum diukur** pada jawaban model yang sebenarnya.

1. Model yang patuh pada perintah sistem (menjawab "tidak ditemukan" tanpa kutipan) justru ditolak, karena validator menolak setiap jawaban tanpa kutipan.
2. Regex terlalu kaku terhadap variasi umum: `[S1:L3-5]`, `[S1: L3]`, tanda hubung panjang, `[S1, L3]`, `[S1:L3, L5]`. Kutipan yang salah format diabaikan diam-diam.
3. Referensi tidak dikenal (nomor sumber melebihi jumlah sumber terpilih, `N1` tanpa Catatan aktif, atau `N2` dan seterusnya), atau rentang melebihi 50 baris.
4. Riwayat memuat kutipan lama. Penomoran `S1..S8` dibuat ulang pada setiap permintaan, sehingga nomor lama dapat menunjuk sumber lain.
5. Model kecil lokal kurang patuh pada format yang persis.

### 1.3 Usulan, berurutan menurut prioritas

1. **Ukur dahulu.** Catat alasan penolakan sebagai kode (tanpa isi sumber dan tanpa isi jawaban) agar penyebab dominan diketahui sebelum perilaku diubah. Selaras dengan prinsip optimasi berdasarkan jalur nyata.
2. **Jawaban "tidak ditemukan" yang sah.** Izinkan jawaban tanpa kutipan bila model menyatakan informasi tidak ada dalam sumber. Perlu penanda yang tegas agar tidak menjadi celah bagi jawaban berklaim tanpa kutipan. Perintah sistem dan validator diubah bersamaan agar kontraknya sama.
3. **Normalisasi format kutipan.** Terima variasi yang tidak ambigu, lalu ubah ke bentuk baku sebelum disimpan.
4. **Kebijakan kutipan tidak valid.** Pilihan: (a) tolak seluruh jawaban seperti sekarang; (b) simpan jawaban dengan tanda "kutipan tidak terverifikasi" dan buang kutipan yang salah; (c) coba ulang satu kali dengan pesan koreksi. Opsi (c) menggandakan waktu tunggu. Putuskan setelah butir 1.
5. **Riwayat.** Buang atau ubah nomor kutipan lama pada riwayat yang dikirim ke model.
6. **Batas rentang.** Tinjau batas 50 baris; pertimbangkan menolak hanya kutipan berlebih, bukan seluruh jawaban.
7. **Kutip lalu cari.** Model menyalin cuplikan pendek kata demi kata; server mencari baris cuplikan itu. Nomor baris tidak lagi ditebak model, dan dukungan klaim dapat diperiksa secara langsung.
8. **Pemeriksaan dukungan leksikal.** Periksa tumpang tindih kata kunci antara kalimat klaim dan baris yang dikutip. Heuristik kasar; perlu diuji agar tidak menolak jawaban yang benar.
9. **Keluaran terstruktur.** Minta JSON atau tata bahasa terbatas dari penyedia. Memerlukan perluasan kontrak penyedia, yang saat ini sengaja kecil (lihat `AI-PROVIDER.md`).
10. **Antarmuka.** Pesan galat yang spesifik, tombol coba lagi, dan kutipan yang tampil sebagai penanda yang membuka lokasi sumber.

## 2. Apakah RAG perlu dirombak?

**Penilaian: belum perlu dirombak.**

- Percakapan saat ini belum RAG sungguhan. Tidak ada tahap pengambilan: konteks adalah isi penuh sumber yang dipilih manual. Pencarian FTS5 (`internal/db/search.go`) ada, tetapi terpisah dari percakapan.
- Kegagalan yang dilaporkan berada pada kontrak keluaran (format dan validasi kutipan), bukan pada pengambilan.
- Perombakan besar bertentangan dengan `AGENTS.md`: perubahan kecil dan mudah dibalik, ukur sebelum mengoptimalkan, skema berisiko.

Urutan yang disarankan: (1) perbaiki kontrak kutipan pada bagian 1; (2) pengambilan ringan pada 2.1; (3) penyematan pada 2.2 hanya jika langkah (2) terbukti kurang.

### 2.1 Pengambilan ringan berbasis FTS5

- Pecah sumber menjadi potongan berbasis baris atau paragraf, lalu pilih potongan teratas dengan `bm25` yang sudah ada. Tanpa penyematan, tetap offline.
- Kirim potongan dengan **nomor baris asli** dari sumber, bukan penomoran ulang, agar tetap cocok dengan `source_locations` dan checksum.
- Manfaat: tidak terbentur batas 64 KiB, konteks lebih pendek sehingga model kecil lebih patuh.
- Risiko: `searchExpression` menggabungkan semua kata dengan AND, sehingga pertanyaan berbentuk kalimat mudah tidak menemukan apa pun. Perlu ekstraksi kata kunci atau OR, dan daftar kata henti Bahasa Indonesia.
- Pertanyaan lanjutan ("maksudnya apa?") perlu ditulis ulang dari riwayat sebelum dicari.
- Tahap awal dapat berjalan tanpa migrasi (potongan dihitung saat permintaan). Penyimpanan potongan hanya dipertimbangkan bila pengukuran menunjukkan perlu.

### 2.2 Penyematan dan pencarian semantik

- Memerlukan kontrak penyedia untuk embeddings (saat ini di luar kontrak), tabel vektor baru melalui migrasi yang ditinjau cermat, dan pembatalan vektor saat checksum sumber berubah.
- Berdampak pada ukuran data dan waktu pengindeksan.
- Cocok dengan Tahap 4. Syarat masuk: bukti bahwa 2.1 tidak cukup.

### 2.3 Sumber besar

- Peta dokumen atau ringkasan per bagian sebagai Transformasi (Tahap 4), dipakai untuk memilih potongan pada sumber yang sangat panjang.

## 3. Evaluasi

- Pengujian tabel untuk `validateCitations` yang mencakup referensi Catatan (`N1`) dan Sumber (`S1..S8`): variasi format, rentang batas, referensi tidak dikenal, dan jawaban tanpa kutipan.
- Kumpulan uji manual kecil dengan sumber sintetis (bukan data nyata) pada runtime lokal sungguhan, sesuai `AI-PROVIDER.md`. Ukuran yang dicatat: persentase jawaban lolos validasi, kebenaran kutipan, dan ketepatan jawaban "tidak ditemukan".
- Jangan menulis isi sumber atau isi jawaban ke log (lihat bagian privasi di `KEPUTUSAN-DESAIN.md`).

## 4. Ide lain (belum dinilai)

- Tandai kutipan kedaluwarsa pada antarmuka. Untuk Sumber, bila `source_checksum` tidak lagi sama dengan checksum sumber saat ini. Untuk Catatan, yang dapat diedit, perlu dicek apakah versi atau checksum catatan tersedia sebagai pembanding.
- Catatan (`N`) dan Sumber (`S`) tetap satu kontrak (format, validator, perintah sistem). Perbedaannya hanya pada stabilitas target kutipan, bukan pada kontrak ke model.
- Simpan jawaban percakapan sebagai catatan beserta rujukan lokasi sumbernya.
- Batas konteks yang dapat diatur per penyedia atau model, karena model kecil berjendela kecil.
- Pilihan eksplisit di antarmuka: "hanya dari sumber" atau "boleh pengetahuan umum".
- Sertakan kutipan pada ekspor percakapan (lihat `EXPORT.md`).
