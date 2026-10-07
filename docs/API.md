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
