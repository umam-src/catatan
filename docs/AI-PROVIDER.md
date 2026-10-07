# Penyedia model lokal

## Tujuan

Catatan memakai batas kecil antara logika aplikasi dan server model. Dukungan AI bersifat opsional; fungsi dasar tidak membutuhkan penyedia model.

## Kontrak internal

Paket \`internal/ai\` menyediakan:

- \`Provider\` untuk batas pemanggilan model;
- \`Request\` berisi model dan pesan;
- \`Response\` berisi teks dan nama model;
- jenis pesan system, user, dan assistant;
- galat terklasifikasi untuk ketidaktersediaan, penolakan, waktu habis, respons tidak sah, dan batas ukuran.

Kontrak sengaja lebih kecil daripada seluruh spesifikasi OpenAI. Embeddings, streaming, dan tool calling belum menjadi bagian kontrak awal.

## Server kompatibel OpenAI

Implementasi menggunakan HTTP standar Go dan endpoint:

\`POST /v1/chat/completions\`

Permintaan minimum:

\`\`\`json
{
  "model": "nama-model",
  "messages": [
    {"role": "user", "content": "Halo"}
  ]
}
\`\`\`

Respons yang diterima harus memiliki setidaknya satu pilihan dengan isi pesan teks.

Alamat dasar dapat menunjuk ke server lokal, misalnya \`http://127.0.0.1:8080\`. Tidak ada alamat awan yang diwajibkan.

## Pemeriksaan kesiapan

\`Client.Probe\` melakukan pemeriksaan ringan ke \`GET /v1/models\`.

Probe:

- tidak mengirim prompt atau isi catatan;
- tidak menjalankan generasi model;
- memiliki batas waktu 2 detik;
- mengirim kunci API hanya jika konfigurasi memang memilikinya;
- tidak menyimpan atau mencatat respons provider.

Status yang dikembalikan:

| Status | Arti |
|---|---|
| \`ready\` | Endpoint kompatibel merespons dengan JSON yang sah |
| \`unavailable\` | Server tidak dapat dihubungi atau mengembalikan HTTP 5xx |
| \`invalid\` | Endpoint merespons tetapi kontraknya tidak sesuai |

Status \`ready\` berarti server API siap dijangkau, bukan jaminan bahwa model tertentu tersedia atau mampu memenuhi semua permintaan.

## Batas bawaan

- waktu permintaan: 30 detik;
- waktu probe: 2 detik;
- ukuran permintaan: 1 MiB;
- ukuran respons: 1 MiB.

Nilai tersebut dapat diubah oleh pemanggil saat diperlukan. Batas diterapkan sebelum data diteruskan atau dibaca penuh.

## Runtime lokal

Catatan tidak membundel runtime inference atau model.

Runtime seperti llama-server dan Ollama dapat digunakan melalui kontrak OpenAI-compatible yang sama. Catatan cukup mengetahui alamat dasar dan model yang dipilih; tidak perlu adapter atau SDK khusus runtime.

Contoh alamat lokal:

- llama-server: sesuai alamat dan port yang dijalankan pengguna;
- Ollama: \`http://127.0.0.1:11434/v1\`.

Pengujian dengan runtime sungguhan dilakukan di lingkungan pengembangan, bukan sebagai prasyarat CI.

## Konfigurasi dan rahasia

Untuk penggunaan biasa, alamat server dan nama model dapat disimpan di `config.json` pada folder data Catatan:

```json
{
  "ai": {
    "url": "http://127.0.0.1:11434/v1",
    "model": "llama3.2"
  }
}
```

Environment variable tetap tersedia untuk pengembangan dan CI. `CATATAN_AI_URL` dan `CATATAN_AI_MODEL` menimpa nilai dari berkas jika diisi. `CATATAN_AI_API_KEY` hanya dibaca dari lingkungan proses.

Kunci API:

- tidak disimpan di SQLite;
- tidak disimpan di `config.json`;
- tidak ditulis ke log;
- tidak dimasukkan ke model data aplikasi.

Catatan tidak menganggap kunci API sebagai bagian dari identitas atau otorisasi pengguna.

## Galat

| Kondisi | Galat |
|---|---|
| Server tidak dapat dihubungi | \`ErrProviderUnavailable\` |
| HTTP 5xx | \`ErrProviderUnavailable\` |
| HTTP 4xx | \`ErrProviderRejected\` |
| Waktu habis | \`ErrTimeout\` |
| JSON/struktur respons tidak sesuai | \`ErrInvalidResponse\` |
| Permintaan melewati batas | \`ErrRequestTooLarge\` |
| Respons melewati batas | \`ErrResponseTooLarge\` |

Galat dapat diperiksa dengan \`errors.Is\` tanpa bergantung pada teks pesan.

## Pengujian

Pengujian menggunakan \`httptest\` dan tidak mengunduh model atau menghubungi layanan eksternal. Probe diuji untuk status siap, tidak tersedia, respons tidak sah, JSON tidak sah, dan memastikan tidak ada prompt yang dikirim.

Integrasi runtime model sungguhan tetap menjadi pengujian manual pada fase integrasi.

## Batas fase

Fase ini tidak:

- membundel model;
- membundel runtime inference;
- menambahkan SDK Ollama atau llama.cpp;
- mewajibkan embeddings;
- mewajibkan streaming;
- menambahkan tool calling.

## Percakapan dan kutipan

Percakapan menggunakan CATATAN_AI_MODEL sebagai nama model runtime. Isi sumber hanya dikirim jika source_ids dipilih pengguna dan sumber tersebut berada pada buku percakapan.

Batas awal konteks adalah 8 sumber dan 64 KiB isi gabungan. Isi sumber diverifikasi dengan SHA-256 sebelum dikirim. Jawaban dengan konteks sumber harus menyertakan kutipan [S1:L1-L3]; server memeriksa ID sumber dan rentang baris sebelum menyimpannya.

Isi sumber diperlakukan sebagai data tidak tepercaya, bukan instruksi, untuk membatasi dampak injeksi prompt.
