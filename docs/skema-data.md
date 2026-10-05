# Skema Data

## Prinsip

1. Data pengguna tidak bercampur dengan berkas program.
2. Perubahan skema hanya melalui migrasi baru.
3. Migrasi tidak menghapus data secara default.
4. Kolom `metadata_json` menyediakan ruang kecil untuk informasi tambahan tanpa mengubah tabel setiap kali fitur kecil ditambahkan.
5. ID berupa teks sehingga impor/ekspor dari sumber lain tidak memerlukan konversi angka.
6. `sources` sengaja hadir sejak awal walaupun antarmuka tahap pertama hanya membuat catatan teks.

## Entitas

### notebooks

Wadah kerja. Satu buku dapat memiliki banyak sumber, catatan, dan percakapan.

### sources

Bahan mentah yang nantinya dapat berupa teks, URL, dokumen, atau jenis lain. Tahap awal belum menyediakan impor selain teks.

### notes

Hasil catatan pengguna atau keluaran AI. `note_type` membedakan asal tanpa membuat tabel terpisah.

Catatan aktif tidak memiliki nilai pada `deleted_at`. Saat dihapus, kolom tersebut diisi dengan waktu penghapusan. Daftar catatan aktif hanya membaca baris dengan `deleted_at IS NULL`.

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

Migrasi `002_note_deletion.sql` menambahkan kolom `deleted_at` dan indeks parsial untuk daftar catatan aktif. Migrasi ini mempertahankan data lama karena catatan yang sudah ada tetap aktif hingga pengguna menghapusnya.

Pola ini mengikuti praktik Open Notebook yang menggunakan migrasi berurutan dan tidak menggabungkan ulang migrasi lama setelah masuk ke cabang utama.
