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

## Tema

Seluruh warna, jarak, radius, bayangan, dan lebar kolom berasal dari token di `web/static/ui.css`. Komponen tidak boleh menulis warna langsung.

- Tema terang dan gelap tersedia; gelap dipilih otomatis mengikuti perangkat.
- Huruf antarmuka dan huruf isi catatan memakai huruf bawaan sistem. Isi catatan berserif; ganti lewat token `--huruf-catatan`.
- Ukuran teks memakai `rem` dan tidak ada teks di bawah 12 px. Isi catatan 17 px (18 px di ponsel) dengan tinggi baris 1,8.
- Kontras teks minimal 4,5:1 pada kedua tema.
- Target sentuh minimal 44 px pada butir yang dapat diketuk.
- Gerak hanya menjawab tindakan pengguna dan mengikuti `prefers-reduced-motion`.

## Dialog

Dialog memakai elemen `<dialog>` modal sehingga fokus terkunci dan Esc menutupnya. Dialog bawaan peramban (`prompt`, `confirm`) tidak dipakai.

- Pengaturan: ukuran tetap agar tidak melompat saat isi bertambah.
- Aksi (buat, ubah nama, konfirmasi): tombol utama di kanan, Batal di kiri. Konfirmasi hapus memfokuskan Batal.
- Di ponsel, dialog tampil sebagai lembar bawah.

## Responsif

- desktop: N | C | S/A;
- layar menengah (hingga 1120 px): S/A menjadi panel dari sisi kanan;
- ponsel (hingga 760 px): satu panel pada satu waktu.
  - daftar catatan adalah layar awal; memilih catatan membuka editor layar penuh dengan tombol kembali;
  - bilah bawah Catatan / Sumber / Artefak; Sumber dan Artefak naik sebagai lembar bawah;
  - tombol catatan baru mengambang di jangkauan jempol;
  - daftar buku berupa laci dari kiri;
  - membuka sumber menutup lembar dan menampilkan isinya di ruang kerja.

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

## Pencarian lokal

Antarmuka utama menyediakan pencarian catatan dan sumber dari buku yang sedang aktif. Kueri dikirim ke API lokal setelah penundaan singkat agar perubahan setiap karakter tidak langsung menghasilkan permintaan baru.

- Hasil selalu dibatasi pada buku aktif.
- Permintaan pencarian lama dibatalkan ketika kueri berubah.
- Kueri tidak disimpan sebagai data pengguna.
- Daftar hasil hanya menampilkan judul, jenis, dan buku; isi catatan atau sumber tidak dikirim ke antarmuka pencarian.
- Antarmuka menampilkan keadaan memuat, tidak ada hasil, dan kesalahan.
- Pemilihan hasil membuka catatan atau sumber yang sudah dimuat pada buku aktif.
