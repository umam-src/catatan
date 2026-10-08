# Issue #94 — Uji model Catatan Jaga untuk chat dengan sumber

Status: **Terbuka**

## Tujuan

Menguji model yang digunakan oleh Catatan Jaga pada alur chat dengan sumber sebelum menyimpulkan bahwa kegagalan disebabkan model terlalu kecil.

## Alur yang diuji

**Catatan → pencarian → sumber/context → model Catatan Jaga → jawaban + citation**

## Checklist

- [ ] Identifikasi konfigurasi model Catatan Jaga.
- [ ] Pastikan runtime lokal terhubung ke model.
- [ ] Siapkan catatan dengan fakta uji yang mudah dikenali.
- [ ] Pastikan catatan/sumber muncul dalam pencarian.
- [ ] Pilih sumber sebagai konteks chat.
- [ ] Ajukan pertanyaan yang jawabannya hanya tersedia dari sumber.
- [ ] Verifikasi jawaban menggunakan konteks catatan.
- [ ] Verifikasi citation menunjuk ke sumber yang benar.
- [ ] Catat titik kegagalan jika model tidak mengikuti konteks/instruksi.
- [ ] Jika gagal, bandingkan dengan model yang lebih besar.

## Kriteria selesai

Hasil pengujian menentukan apakah masalah berada pada integrasi Catatan atau kemampuan model.

> Jangan menyimpulkan model 0.5B terlalu kecil sebelum model Catatan Jaga diuji.

## Referensi

- GitHub Issue #94: https://github.com/umam-src/catatan/issues/94

## Verifikasi 2026-10-08

Lingkungan kerja agen diperiksa untuk runtime model lokal:

- ollama tidak tersedia.
- llama-server tidak tersedia.
- http://127.0.0.1:11434/v1/models tidak dapat diakses.
- http://127.0.0.1:8080/v1/models merespons HTTP 404, sehingga bukan endpoint provider Catatan yang siap diuji.

**Kesimpulan:** pengujian model Catatan Jaga end-to-end belum dapat dilakukan dari lingkungan ini. Issue #94 tetap terbuka dan tidak boleh dianggap lulus.
