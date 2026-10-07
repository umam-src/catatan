# Skema basis data

Dokumen ini menjelaskan bentuk data yang disimpan Catatan, hubungan antar data, serta aturan perubahan skema. Sumber kebenaran untuk perubahan struktur tetap berada pada berkas migrasi di `internal/db/migrations/`.

## Gambaran umum

Catatan menggunakan basis data lokal. Data utama dikelompokkan menjadi:

- **identitas**: pengguna, kredensial, dan sesi;
- **buku**: wadah catatan dan sumber;
- **catatan**: isi yang ditulis pengguna;
- **sumber**: bahan teks yang disimpan secara lokal;
- **percakapan**: struktur data yang disiapkan untuk alur percakapan;
- **lokasi sumber**: penanda rentang baris pada sumber.

Semua ID disimpan sebagai `TEXT`. Waktu disimpan sebagai `TEXT` dalam format waktu yang digunakan aplikasi.

## Tabel

### `metadata`

Menyimpan pasangan kunci-nilai untuk informasi format basis data.

| Kolom | Tipe | Aturan | Keterangan |
|---|---|---|---|
| `key` | TEXT | PRIMARY KEY | Nama metadata. |
| `value` | TEXT | NOT NULL | Nilai metadata. |

Saat basis data baru dibuat, `format` diisi dengan `buku-catatan/v1`.

### `users`

Identitas pengguna lokal.

| Kolom | Tipe | Aturan | Keterangan |
|---|---|---|---|
| `id` | TEXT | PRIMARY KEY | ID pengguna. |
| `username` | TEXT | NOT NULL, UNIQUE | Nama pengguna. |
| `email` | TEXT | NOT NULL, default `''` | Alamat surel jika digunakan. |
| `display_name` | TEXT | NOT NULL, default `''` | Nama tampilan. |
| `created_at` | TEXT | NOT NULL | Waktu pembuatan. |
| `updated_at` | TEXT | NOT NULL | Waktu perubahan. |

Pengguna bawaan menggunakan ID `local`.

### `auth_credentials`

Kredensial autentikasi pengguna.

| Kolom | Tipe | Aturan | Keterangan |
|---|---|---|---|
| `user_id` | TEXT | PRIMARY KEY, FK → `users.id` | Satu kredensial per pengguna. |
| `password_hash` | TEXT | NOT NULL | Nilai hash kata sandi. |
| `updated_at` | TEXT | NOT NULL | Waktu perubahan kredensial. |

Jika pengguna dihapus, kredensialnya ikut dihapus.

### `sessions`

Sesi autentikasi lokal.

| Kolom | Tipe | Aturan | Keterangan |
|---|---|---|---|
| `id` | TEXT | PRIMARY KEY | ID sesi. |
| `user_id` | TEXT | NOT NULL, FK → `users.id` | Pemilik sesi. |
| `token_hash` | TEXT | NOT NULL, UNIQUE | Hash token sesi. |
| `expires_at` | TEXT | NOT NULL | Waktu kedaluwarsa. |
| `created_at` | TEXT | NOT NULL | Waktu pembuatan. |
| `revoked_at` | TEXT | nullable | Waktu pencabutan jika dicabut. |

Jika pengguna dihapus, sesinya ikut dihapus.

### `notebooks`

Buku sebagai wadah utama catatan dan sumber.

| Kolom | Tipe | Aturan | Keterangan |
|---|---|---|---|
| `id` | TEXT | PRIMARY KEY | ID buku. |
| `title` | TEXT | NOT NULL | Judul buku. |
| `description` | TEXT | NOT NULL, default `''` | Deskripsi buku. |
| `archived` | INTEGER | NOT NULL, `0` atau `1` | Status arsip. |
| `created_at` | TEXT | NOT NULL | Waktu pembuatan. |
| `updated_at` | TEXT | NOT NULL | Waktu perubahan. |
| `owner_id` | TEXT | FK → `users.id` | Pemilik buku. |

Buku yang dimiliki pengguna ikut terhapus ketika pemilik dihapus. Data lama memperoleh pemilik `local` saat migrasi autentikasi dijalankan.

### `sources`

Sumber yang disimpan secara lokal dan dikaitkan dengan buku.

| Kolom | Tipe | Aturan | Keterangan |
|---|---|---|---|
| `id` | TEXT | PRIMARY KEY | ID sumber. |
| `notebook_id` | TEXT | NOT NULL, FK → `notebooks.id` | Buku pemilik sumber. |
| `title` | TEXT | NOT NULL | Judul sumber. |
| `kind` | TEXT | NOT NULL, default `text` | Jenis sumber. |
| `content` | TEXT | NOT NULL, default `''` | Isi sumber teks. |
| `locator` | TEXT | NOT NULL, default `''` | Lokasi tambahan jika diperlukan. |
| `checksum` | TEXT | NOT NULL, default `''` | Checksum isi sumber. |
| `metadata_json` | TEXT | NOT NULL, default `{}` | Metadata sumber dalam JSON. |
| `created_at` | TEXT | NOT NULL | Waktu pembuatan. |
| `updated_at` | TEXT | NOT NULL | Waktu perubahan. |

Sumber ikut terhapus ketika bukunya dihapus.

### `notes`

Catatan pengguna.

| Kolom | Tipe | Aturan | Keterangan |
|---|---|---|---|
| `id` | TEXT | PRIMARY KEY | ID catatan. |
| `notebook_id` | TEXT | NOT NULL, FK → `notebooks.id` | Buku pemilik catatan. |
| `title` | TEXT | NOT NULL | Judul catatan. |
| `content` | TEXT | NOT NULL, default `''` | Isi catatan. |
| `note_type` | TEXT | NOT NULL, default `human` | Jenis catatan. |
| `metadata_json` | TEXT | NOT NULL, default `{}` | Metadata dalam JSON. |
| `created_at` | TEXT | NOT NULL | Waktu pembuatan. |
| `updated_at` | TEXT | NOT NULL | Waktu perubahan. |
| `deleted_at` | TEXT | nullable | Waktu penghapusan lunak. |

Catatan ikut terhapus ketika bukunya dihapus.

### `conversations`

Percakapan yang terkait dengan buku.

| Kolom | Tipe | Aturan | Keterangan |
|---|---|---|---|
| `id` | TEXT | PRIMARY KEY | ID percakapan. |
| `notebook_id` | TEXT | NOT NULL, FK → `notebooks.id` | Buku pemilik percakapan. |
| `title` | TEXT | NOT NULL, default `''` | Judul percakapan. |
| `created_at` | TEXT | NOT NULL | Waktu pembuatan. |
| `updated_at` | TEXT | NOT NULL | Waktu perubahan. |

Percakapan ikut terhapus ketika bukunya dihapus.

### `messages`

Pesan dalam percakapan.

| Kolom | Tipe | Aturan | Keterangan |
|---|---|---|---|
| `id` | TEXT | PRIMARY KEY | ID pesan. |
| `conversation_id` | TEXT | NOT NULL, FK → `conversations.id` | Percakapan pemilik pesan. |
| `role` | TEXT | NOT NULL | Salah satu dari `system`, `user`, `assistant`, `tool`. |
| `content` | TEXT | NOT NULL | Isi pesan. |
| `metadata_json` | TEXT | NOT NULL, default `{}` | Metadata dalam JSON. |
| `created_at` | TEXT | NOT NULL | Waktu pembuatan. |

Pesan ikut terhapus ketika percakapannya dihapus.

### `search_documents`

Cermin lokal untuk kebutuhan pencarian. Data utama tetap berada pada `notes` dan `sources`; tabel ini hanya menjaga satu sumber dokumen untuk indeks pencarian dan diperbarui otomatis melalui pemicu basis data.

| Kolom | Tipe | Aturan | Keterangan |
|---|---|---|---|
| `id` | TEXT | PRIMARY KEY | ID gabungan sumber pencarian. |
| `notebook_id` | TEXT | NOT NULL, FK → `notebooks.id` | Buku pemilik dokumen. |
| `kind` | TEXT | NOT NULL | `note` atau `source`. |
| `title` | TEXT | NOT NULL | Judul yang dicari. |
| `content` | TEXT | NOT NULL | Isi yang dicari. |

Indeks FTS5 menggunakan tabel ini untuk pencarian lokal. Pengguna tidak mengakses tabel ini secara langsung.

### `source_locations`

Penanda rentang baris pada sumber teks.

| Kolom | Tipe | Aturan | Keterangan |
|---|---|---|---|
| `id` | TEXT | PRIMARY KEY | ID penanda. |
| `source_id` | TEXT | NOT NULL, FK → `sources.id` | Sumber yang ditandai. |
| `start_line` | INTEGER | NOT NULL, ≥ 1 | Baris awal. |
| `end_line` | INTEGER | NOT NULL, ≥ `start_line` | Baris akhir. |
| `source_checksum` | TEXT | NOT NULL | Checksum sumber saat penanda dibuat. |
| `created_at` | TEXT | NOT NULL | Waktu pembuatan. |

Penanda ikut terhapus ketika sumber dihapus.

## Relasi utama

```text
users
 ├── auth_credentials
 ├── sessions
 └── notebooks
      ├── notes
      ├── sources
      │    └── source_locations
      └── conversations
           └── messages
```

Kepemilikan buku menjadi batas utama isolasi data. Catatan, sumber, percakapan, dan data turunannya harus diakses melalui buku atau pemilik yang berwenang.

## Indeks

Indeks saat ini mendukung daftar data berdasarkan waktu dan pemeriksaan kepemilikan/sesi:

- `idx_notebooks_updated`
- `idx_sources_notebook_updated`
- `idx_notes_notebook_updated`
- `idx_conversations_notebook_updated`
- `idx_messages_conversation_created`
- `idx_sessions_token_hash`
- `idx_sessions_user_active`
- `idx_notebooks_owner_updated`
- `idx_notes_notebook_deleted`
- `idx_notes_notebook_updated_active`
- `idx_notes_notebook_deleted_updated`
- `idx_source_locations_source`
- `idx_search_documents_notebook`

Jangan menambah indeks hanya berdasarkan perkiraan. Ukur jalur baca yang nyata terlebih dahulu dan pastikan biaya penyimpanan serta perubahan tulis sepadan dengan manfaatnya.

## Migrasi

Migrasi diberi nomor berurutan dan dijalankan oleh sistem migrasi aplikasi. Nomor migrasi tidak boleh digunakan dua kali.

Urutan yang tercatat pada pengembangan saat ini:

| Nomor | Berkas | Tujuan |
|---|---|---|
| 001 | `001_init.sql` | Struktur dasar buku, sumber, catatan, percakapan, pesan, metadata, dan indeks awal. |
| 002 | `002_note_deletion.sql` | Penghapusan lunak catatan dan indeks terkait. |
| 003 | `003_auth.sql` | Identitas, kredensial, sesi, serta kepemilikan buku. |
| 004 | `004_note_list_index.sql` | Indeks daftar catatan. |
| 005 | `005_source_locations.sql` | Penanda lokasi pada sumber. |
| 006 | `006_search.sql` | Indeks pencarian lokal FTS5 dan sinkronisasi catatan/sumber. |

Pada pengembangan sebelum `v1.0.0`, migrasi dapat dirapikan jika riwayat pengembangan memang mengandung kesalahan. Namun setiap perubahan nomor harus disertai pengujian migrasi dari basis data kosong dan basis data yang sudah memiliki migrasi sebelumnya.

## Aturan perubahan skema

1. Periksa dokumen ini dan migrasi yang sudah ada sebelum mengubah struktur.
2. Gunakan nomor migrasi baru; jangan mendaur ulang nomor yang sudah digunakan.
3. Pertahankan `FOREIGN KEY`, `CHECK`, dan indeks yang memang diperlukan untuk menjaga konsistensi dan jalur baca.
4. Uji pembukaan basis data baru dan pembaruan basis data yang sudah ada.
5. Perbarui dokumen ini jika tabel, kolom, relasi, indeks, atau aturan penting berubah.
6. Jangan memasukkan data nyata, rahasia, token, atau data pribadi ke contoh migrasi maupun pengujian.
7. Setelah `v1.0.0`, perlakukan skema sebagai kontrak stabil dan gunakan migrasi yang kompatibel dengan data yang sudah ada.

## Backup dan pemulihan

Format cadangan harus mempertahankan data buku, catatan, sumber, serta metadata penting tanpa menyimpan rahasia sistem. Pemeriksaan integritas dilakukan sebelum data aktif ditimpa. Detail format cadangan mengikuti keputusan implementasi Fase 2 dan harus diperbarui bersama dokumen ini ketika format tersebut ditetapkan.


## Pengukuran pencarian lokal

Pengukuran satu kali pada runner CI Linux, menggunakan dataset sintetis dan 100 pencarian setelah pemanasan:

| Data uji | Ukuran basis data | Rata-rata pencarian |
| ---: | ---: | ---: |
| 1.000 catatan | 720.896 byte (≈0,69 MiB) | 2,73 ms |
| 10.000 catatan | 5.386.240 byte (≈5,14 MiB) | 25,20 ms |

Hasil ini menjadi garis dasar untuk optimasi berikutnya. Belum ada bukti yang membenarkan indeks tambahan atau struktur pencarian yang lebih kompleks. Pengukuran tidak dijalankan pada CI rutin agar waktu CI tetap hemat.
