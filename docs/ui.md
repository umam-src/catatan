# Desain antarmuka

Dokumen ini adalah acuan struktur antarmuka Catatan. Perubahan struktur utama harus mengikuti keputusan di dokumen ini atau diperbarui bersamaan dengan keputusan baru.

## Struktur utama

Pada layar lebar, antarmuka mengikuti urutan:

```
N — Catatan  |  C — Ruang Kerja  |  S/A — Konteks
```

### N — Catatan

N adalah daftar pekerjaan utama, bukan sekadar bilah navigasi.

- Menampilkan daftar catatan pada buku aktif.
- Memungkinkan memilih catatan.
- Menyediakan tindakan untuk membuat catatan.
- Tetap menjadi area kiri utama pada layar lebar.

### C — Ruang Kerja

C adalah ruang kerja universal. Ia tidak dibuat sebagai editor khusus untuk satu jenis isi.

C dapat mengambil keadaan:

- **Note** — melihat dan mengedit catatan.
- **Chat** — percakapan dengan konteks yang dipilih.
- **Note + Chat** — membagi ruang kerja untuk menulis sambil bertanya.
- **Artifact** — melihat atau mengedit artefak yang dipilih.
- **Preview** — melihat isi yang sedang dibuka.

Mode split berada di dalam C, bukan menjadi kolom keempat.

Contoh:

```
┌──────────────────────────────┐
│ Note            │ Chat       │
│ isi catatan     │ tanya...   │
│                 │ jawaban    │
└──────────────────────────────┘
```

### S/A — Konteks

S/A adalah satu panel yang dapat berganti antara **Sumber** dan **Artefak**.

Sumber menyediakan:

- daftar dokumen yang dipilih sebagai referensi;
- tambah dan hapus sumber;
- pratinjau;
- tindakan Buka untuk membaca dalam tampilan penuh.

Sumber tetap merupakan sumber asli. Sumber tidak dipindahkan atau disalin menjadi artefak.

Artefak menyediakan daftar hasil kerja. Memilih artefak membuat C menjadi ruang kerja artefak.

S dan A tidak ditampilkan sebagai dua kolom permanen.

## Hubungan kerja

```
Sumber
  │
  ├── dipilih sebagai referensi
  ├── dibuka untuk dibaca
  └── digunakan Chat
          │
          ├── jawaban
          ├── Catatan
          └── Artefak

N → C → S/A
```

N adalah pekerjaan utama, C adalah tempat bekerja, dan S/A adalah konteks serta hasil.

## Responsif

Desktop:

```
N | C | S/A
```

Layar lebih sempit:

```
N | C
    └── S/A dibuka sebagai panel
```

Mobile menggunakan drawer atau tab untuk N, C, dan S/A. Struktur tidak dipaksa menjadi empat kolom kecil.

## Prinsip

- Alur utama tetap dapat digunakan tanpa jaringan.
- Pertahankan perilaku API yang sudah benar.
- Tidak menambah pustaka antarmuka berat tanpa kebutuhan yang jelas.
- Keadaan loading, kosong, galat, dan penyimpanan harus jelas.
- Fokus keyboard harus terlihat.
- Kontrol interaktif menggunakan elemen HTML yang sesuai.
- Teks antarmuka menggunakan Bahasa Indonesia baku.

## Batas implementasi saat ini

Backend sumber sudah tersedia untuk daftar, impor, pratinjau, lokasi, perubahan, dan penghapusan sumber. Chat dan artefak dapat diperkenalkan bertahap; fondasi UI tidak boleh membuat editor terpisah untuk masing-masing jenis pekerjaan.

Desain ini menjadi fondasi untuk pengembangan Chat, Sumber, dan Artefak berikutnya tanpa mengubah pola N → C → S/A.
