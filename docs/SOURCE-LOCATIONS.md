# Penanda lokasi sumber

Catatan menyimpan penanda lokasi untuk sumber teks dalam bentuk rentang nomor baris.

## Bentuk

Satu penanda berisi:

- `source_id`: sumber yang dirujuk.
- `start_line`: nomor baris awal, dimulai dari 1.
- `end_line`: nomor baris akhir, termasuk batas akhir.
- `source_checksum`: checksum SHA-256 sumber saat penanda dibuat.
- `created_at`: waktu penanda dibuat.

Rentang ini tidak bergantung pada tampilan antarmuka dan dapat dipakai kembali oleh fitur lain.

## Verifikasi

Sebelum penanda dibuat, isi sumber harus lulus pemeriksaan UTF-8 dan SHA-256. Rentang baris juga harus berada di dalam jumlah baris sumber.

Saat penanda digunakan, checksum penanda harus sama dengan checksum sumber saat ini. Dengan begitu, lokasi yang dibuat untuk isi lama tidak dianggap merujuk ke isi baru secara diam-diam.

API menyediakan:

- `POST /api/sources/{id}/locations` untuk menyimpan rentang baris.
- `GET /api/sources/{id}/locations` untuk membaca penanda sumber.

## Batas stabilitas

Nomor baris stabil hanya terhadap versi isi sumber yang memiliki checksum yang sama. Jika isi sumber berubah, penanda lama tidak otomatis dipindahkan atau ditebak ulang. Penanda tersebut harus dianggap tidak berlaku untuk versi baru sampai dibuat penanda baru.

Pendekatan ini sengaja sederhana dan tidak menyimpan salinan kutipan. Penyajian kutipan pada jawaban AI dan pengelolaan kutipan yang lebih kaya tetap berada di fase berikutnya.
