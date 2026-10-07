# Penyedia model lokal

## Tujuan

Catatan memakai batas kecil antara logika aplikasi dan server model. Dukungan AI bersifat opsional; fungsi dasar tidak membutuhkan penyedia model.

## Kontrak internal

Paket `internal/ai` menyediakan:

- `Provider` untuk batas pemanggilan model;
- `Request` berisi model dan pesan;
- `Response` berisi teks dan nama model;
- jenis pesan system, user, dan assistant;
- galat terklasifikasi untuk ketidaktersediaan, penolakan, waktu habis, respons tidak sah, dan batas ukuran.

Kontrak sengaja lebih kecil daripada seluruh spesifikasi OpenAI. Embeddings, streaming, dan tool calling belum menjadi bagian kontrak awal.

## Server kompatibel OpenAI

Implementasi awal menggunakan HTTP standar Go dan endpoint:

`POST /v1/chat/completions`

Permintaan minimum:

```json
{
  "model": "nama-model",
  "messages": [
    {"role": "user", "content": "Halo"}
  ]
}
```

Respons yang diterima harus memiliki setidaknya satu pilihan dengan isi pesan teks.

Alamat dasar dapat menunjuk ke server lokal, misalnya `http://127.0.0.1:8080`. Tidak ada alamat awan yang diwajibkan.

## Batas bawaan

- waktu permintaan: 30 detik;
- ukuran permintaan: 1 MiB;
- ukuran respons: 1 MiB.

Nilai tersebut dapat diubah oleh pemanggil saat diperlukan. Batas diterapkan sebelum data diteruskan atau dibaca penuh.

## Konfigurasi dan rahasia

Konfigurasi provider dipisahkan dari logika domain. Kunci API, bila diperlukan oleh server yang dipakai, hanya diberikan ke client saat runtime.

Kunci API:

- tidak disimpan di SQLite;
- tidak ditulis ke berkas konfigurasi oleh paket provider;
- tidak ditulis ke log;
- tidak dimasukkan ke model data aplikasi.

Catatan tidak menganggap kunci API sebagai bagian dari identitas atau otorisasi pengguna.

## Galat

| Kondisi | Galat |
|---|---|
| Server tidak dapat dihubungi | `ErrProviderUnavailable` |
| HTTP 5xx | `ErrProviderUnavailable` |
| HTTP 4xx | `ErrProviderRejected` |
| Waktu habis | `ErrTimeout` |
| JSON/struktur respons tidak sesuai | `ErrInvalidResponse` |
| Permintaan melewati batas | `ErrRequestTooLarge` |
| Respons melewati batas | `ErrResponseTooLarge` |

Galat dapat diperiksa dengan `errors.Is` tanpa bergantung pada teks pesan.

## Pengujian

Pengujian menggunakan `httptest` dan tidak mengunduh model atau menghubungi layanan eksternal. Integrasi runtime model sungguhan ditangani pada #85.

## Batas fase

Fase ini tidak:

- membundel model;
- membundel runtime inference;
- menambahkan SDK Ollama atau llama.cpp;
- mewajibkan embeddings;
- mewajibkan streaming;
- menambahkan tool calling.

Pemilihan dan pengukuran runtime lokal dilakukan pada #85.
