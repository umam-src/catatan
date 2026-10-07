# API lokal

API dipakai oleh antarmuka Buku Catatan melalui alamat lokal aplikasi. API tidak membutuhkan layanan jaringan eksternal.

## Pencarian

### GET /api/search

Mencari judul dan isi catatan atau sumber yang dimiliki pengguna yang sedang masuk.

Parameter kueri:

| Parameter | Wajib | Keterangan |
|---|---|---|
| `q` | Ya | Teks yang dicari. Kueri kosong menghasilkan daftar kosong. |
| `notebook_id` | Tidak | Membatasi hasil ke satu buku yang dapat diakses pengguna. |
| `limit` | Tidak | Jumlah hasil yang diminta. Nilai bawaan dan batas maksimum ditetapkan oleh mesin pencarian. |

Contoh: `GET /api/search?q=rapat&notebook_id=...&limit=10`

Respons berhasil:

```json
{"results":[{"id":"note:...","kind":"note","notebook_id":"...","title":"Catatan rapat","relevance":-1.23}]}
```

Hasil hanya memuat metadata yang diperlukan untuk daftar pencarian. Isi lengkap catatan atau sumber tidak dikirim oleh endpoint ini.

### Keamanan

- Endpoint memerlukan sesi pengguna aktif.
- Identitas pemilik diambil dari sesi; klien tidak dapat menentukan `owner_id`.
- Hasil selalu dibatasi pada data milik pengguna.
- `notebook_id` hanya berfungsi sebagai penyaring, bukan sebagai sumber otorisasi.
- Kueri kosong tidak dianggap sebagai kesalahan.
- Kueri yang tidak valid dikembalikan sebagai `400 Bad Request`.
- Kesalahan internal dikembalikan sebagai `500 Internal Server Error`.

API pencarian tetap lokal dan tidak mengirim kueri atau isi data ke layanan eksternal.


## Status penyedia model

### GET /api/ai/status

Memeriksa keadaan penyedia model lokal yang dikonfigurasi untuk aplikasi.

Endpoint memerlukan sesi pengguna aktif.

Jika AI tidak dikonfigurasi:

```json
{"enabled":false,"status":"disabled"}
```

Jika AI dikonfigurasi:

```json
{"enabled":true,"status":"ready"}
```

Status dapat berupa:

- `disabled`: tidak ada penyedia model yang dikonfigurasi;
- `ready`: endpoint kompatibel OpenAI dapat dijangkau dan merespons JSON yang sah;
- `unavailable`: penyedia tidak dapat dijangkau atau mengembalikan HTTP 5xx;
- `invalid`: penyedia merespons tetapi kontraknya tidak sesuai.

Pemeriksaan menggunakan `GET /v1/models` dan tidak mengirim isi catatan.

### Konfigurasi lokal

Provider diaktifkan hanya jika lingkungan proses memiliki `CATATAN_AI_URL`.

Variabel yang didukung:

| Variabel | Wajib | Keterangan |
|---|---|---|
| `CATATAN_AI_URL` | Ya untuk mengaktifkan | Alamat dasar server kompatibel OpenAI |
| `CATATAN_AI_API_KEY` | Tidak | Kunci runtime bila server membutuhkannya |

Contoh Ollama: `CATATAN_AI_URL=http://127.0.0.1:11434/v1`.

Kunci API hanya dibaca dari lingkungan proses dan tidak disimpan oleh aplikasi.
