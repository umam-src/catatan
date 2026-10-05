# Skema Data

## Prinsip

1. Data pengguna tidak bercampur dengan berkas program.
2. Perubahan skema hanya melalui migrasi baru.
3. Migrasi tidak menghapus data secara default.
4. Kolom `metadata_json` menyediakan ruang kecil untuk informasi tambahan tanpa mengubah tabel setiap kali fitur kecil ditambahkan.
5. ID berupa teks sehingga impor/ekspor dari sumber lain tidak memerlukan konversi angka.
6. `sources` sengaja hadir sejak awal walaupun antarmuka tahap pertama hanya membuat catatan teks.

## Entitas

### users

Identitas internal pengguna. Identitas ini tidak bergantung pada provider autentikasi tertentu.

### auth_credentials

Kredensial login lokal yang terpisah dari identitas pengguna. Kata sandi hanya disimpan sebagai hash.

### sessions

Session aktif yang dapat dicabut dan memiliki masa berlaku.

### notebooks

Wadah kerja. Setiap buku memiliki owner_id yang menunjuk ke pengguna internal. Satu buku dapat memiliki banyak sumber, catatan, dan percakapan.

### sources

Bahan mentah yang nantinya dapat berupa teks, URL, dokumen, atau jenis lain. Tahap awal belum menyediakan impor selain teks.

### notes

Hasil catatan pengguna atau keluaran AI. `note_type` membedakan asal tanpa membuat tabel terpisah. Kepemilikan catatan mengikuti buku yang dimiliki pengguna.

### conversations

Ruang percakapan yang terikat pada satu buku. Disiapkan sekarang agar penambahan chat tidak membutuhkan migrasi besar.

### messages

Pesan percakapan terpisah dari sesi. Bentuk ini lebih mudah dikembangkan menjadi streaming, alat, atau referensi sumber.

## Kompatibilitas

Skema internal **bukan salinan skema Open Notebook**. Kompatibilitas dijaga pada tingkat konsep dan melalui lapisan impor/ekspor yang akan datang. Ini sengaja dilakukan agar perubahan internal Buku Catatan tidak merusak data ketika fitur baru masuk.

## Aturan migrasi

- Nomor migrasi selalu naik.
- Tidak mengubah migrasi yang sudah dirilis.
- Perubahan destruktif harus menyediakan jalur pemulihan atau ekspor.
- Uji migrasi terhadap basis data yang sudah berisi data.
- Cadangan dibuat sebelum migrasi mayor.
- Migrasi autentikasi memberi buku awal kepada pengguna internal `local` agar data lama tetap memiliki pemilik.

Pola ini mengikuti praktik Open Notebook yang menggunakan migrasi berurutan dan tidak menggabungkan ulang migrasi lama setelah masuk ke cabang utama.


## Authorization dan ACL

Authorization menggunakan ACL sebagai fondasi. Pola awal adalah User/Group → ACL → Resource.

### acl

Tabel acl menyimpan izin yang benar-benar diberikan pada sumber daya.

Kolom konseptual:

- resource_id — identitas sumber daya yang dilindungi.
- subject_type — jenis subjek, minimal user atau group.
- subject_id — identitas pengguna atau grup.
- permission — pengenal izin terstandar, misalnya note.read atau notebook.write.
- effect — hasil aturan, dengan dukungan allow/deny bila kebutuhan evaluasi sudah ditetapkan.

Indeks awal yang direncanakan:

- (resource_id, subject_type, subject_id) untuk lookup ACL berdasarkan sumber daya dan subjek.
- (resource_id, permission) untuk evaluasi izin pada sumber daya.
- (subject_type, subject_id) untuk pencarian ACL milik subjek.

ACL tidak disalin massal ke setiap Note hanya untuk mewariskan izin dari parent resource. Jika inheritance diperlukan, parent resource ditelusuri atau effective permission dihitung secara terukur.

### Batas kompleksitas

Fondasi ini tidak mewajibkan Role, inheritance recursive, Allow/Deny kompleks, ABAC, atau policy engine eksternal. Kemampuan tersebut hanya ditambahkan jika kebutuhan nyata membenarkannya.

## Authorization dan ACL

Authorization menggunakan ACL sebagai fondasi. Pola awal adalah User/Group → ACL → Resource.

### acl

Tabel acl menyimpan izin yang benar-benar diberikan pada sumber daya.

Kolom konseptual:

- resource_id — identitas sumber daya yang dilindungi.
- subject_type — jenis subjek, minimal user atau group.
- subject_id — identitas pengguna atau grup.
- permission — pengenal izin terstandar, misalnya note.read atau notebook.write.
- effect — hasil aturan, dengan dukungan allow/deny bila kebutuhan evaluasi sudah ditetapkan.

Indeks awal yang direncanakan:

- (resource_id, subject_type, subject_id) untuk lookup ACL berdasarkan sumber daya dan subjek.
- (resource_id, permission) untuk evaluasi izin pada sumber daya.
- (subject_type, subject_id) untuk pencarian ACL milik subjek.

ACL tidak disalin massal ke setiap Note hanya untuk mewariskan izin dari parent resource. Jika inheritance diperlukan, parent resource ditelusuri atau effective permission dihitung secara terukur.

### Batas kompleksitas

Fondasi ini tidak mewajibkan Role, inheritance recursive, Allow/Deny kompleks, ABAC, atau policy engine eksternal. Kemampuan tersebut hanya ditambahkan jika kebutuhan nyata membenarkannya.
