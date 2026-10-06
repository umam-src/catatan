# Desain antarmuka

Dokumen ini menjadi acuan dasar untuk antarmuka Catatan yang dibangun ulang dari nol.

## Arah visual

Antarmuka dibuat tenang, terang, ringkas, dan berfokus pada pekerjaan. Tidak ada hiasan yang tidak membantu pengguna memahami keadaan atau mengambil tindakan.

Prinsip utama:
- satu dasar visual untuk seluruh halaman;
- hierarki teks jelas;
- tombol utama mudah ditemukan;
- keadaan kosong dan galat selalu memiliki penjelasan singkat;
- formulir masuk terpisah dari ruang kerja;
- tidak ada pustaka antarmuka tambahan;
- tetap nyaman digunakan tanpa jaringan.

## Halaman masuk

Halaman masuk dan penyiapan akun memakai kartu sederhana di tengah layar. Keduanya berbagi struktur visual yang sama tetapi tidak mencampur fungsi.

Alur:
1. pengguna masuk jika akun sudah tersedia;
2. pengguna dapat beralih ke penyiapan akun pada perangkat baru;
3. setelah berhasil, pengguna langsung masuk ke ruang kerja.

## Ruang kerja

Pada layar lebar, susunan utama adalah:

`N — Catatan | C — Ruang kerja | S/A — Konteks`

### N — Catatan

N berisi buku dan daftar catatan. N adalah daftar pekerjaan utama, bukan sekadar menu.

### C — Ruang kerja

C adalah tempat kerja utama:
- menulis dan mengedit catatan;
- membaca sumber yang dipilih;
- kelak menjadi tempat Chat;
- kelak menjadi tempat editor artefak.

Tidak dibuat halaman editor terpisah untuk setiap jenis hasil kerja.

### S/A — Konteks

S/A adalah panel konteks yang berganti antara:
- **Sumber**, untuk memilih, menambah, menghapus, dan membuka sumber;
- **Artefak**, untuk menampilkan hasil kerja.

Sumber tetap merupakan sumber asli. Isi sumber tidak disalin ke catatan hanya karena dibuka sebagai referensi.

## Responsif

- desktop: N | C | S/A;
- layar menengah: S/A menjadi panel dari sisi kanan;
- ponsel: N dan C menjadi alur utama, S/A tetap tersedia sebagai panel.

## Keadaan

Semua halaman harus memiliki pola yang jelas untuk:
- memuat;
- kosong;
- berhasil;
- menyimpan;
- gagal;
- tindakan yang tidak tersedia.

Pesan menggunakan Bahasa Indonesia yang singkat dan tidak menyalahkan pengguna.

## Batasan

Pekerjaan ini hanya membangun ulang antarmuka. API dan penyimpanan yang sudah benar dipertahankan. Chat dan Artefak belum menambahkan mesin backend baru.
