# Arsitektur

## Tujuan

Arsitektur Buku Catatan mengutamakan satu proses, satu port lokal, satu berkas data, dan pemisahan yang cukup jelas agar fitur baru tidak memaksa perubahan besar pada data lama.

```text
Browser
  │ HTTP lokal
  ▼
Buku Catatan (satu proses Go)
  ├── API
  ├── layanan domain
  ├── penyimpanan SQLite
  └── aset SvelteKit tertanam
```

## Batas tanggung jawab

- `cmd/buku-catatan`: siklus hidup program.
- `internal/http`: kontrak HTTP dan penyajian antarmuka.
- `internal/db`: koneksi, transaksi, dan migrasi.
- `internal/ui`: aset antarmuka yang ditanam.
- `web`: kode antarmuka yang dikembangkan terpisah dari program utama.

Logika bisnis berikutnya sebaiknya ditempatkan di paket domain/service, bukan di pengendali HTTP.

## Mengapa tidak menyalin Open Notebook

Open Notebook memiliki arsitektur tiga lapis dengan antarmuka, API, dan SurrealDB serta cakupan fitur AI yang jauh lebih luas. Buku Catatan mengambil batas domain dan pola migrasi yang baik, tetapi mempertahankan satu program dan satu basis data lokal agar biaya distribusi tetap rendah.

## Jalur menuju fitur AI

AI tidak menjadi dependensi wajib tahap awal. Ketika ditambahkan, gunakan antarmuka penyedia yang terpisah:

```text
Catatan/Sumber
    ↓
Konteks terpilih
    ↓
Antarmuka model
    ├── OpenAI-compatible
    ├── Ollama
    └── llama-server
```

Implementasi lokal tidak boleh mengunci skema ke satu penyedia.

## Antarmuka

SvelteKit dipakai karena hasil produksi dapat dipadatkan dan ditanam ke program. Pendekatan ini sejalan dengan contoh llama.cpp yang menggunakan SvelteKit untuk UI bawaan, tanpa membuat pengguna memasang server frontend terpisah.
