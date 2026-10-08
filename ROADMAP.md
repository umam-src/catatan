# ROADMAP

Roadmap ini menetapkan arah pengembangan Catatan. Pekerjaan berjalan dan kriteria selesai dilacak melalui issue GitHub; keputusan dan catatan kerja rinci disimpan di Obsidian.

## Prinsip

- Stabilitas, kualitas, dan optimasi didahulukan daripada penambahan fitur.
- Offline-first: penggunaan dasar tidak bergantung pada layanan awan.
- Satu program utama dengan UI web tertanam.
- Target ukuran rilis normal <20 MiB; >=30 MiB menjadi peringatan dan >=50 MiB gagal.
- AI bersifat opsional.
- Struktur data berkembang bertahap tanpa mengunci proyek pada satu penyedia.
- Dokumentasi mengikuti perubahan penting.
- Minor naik setiap fase selesai; patch untuk perubahan kode yang dirilis di dalam fase; dokumentasi saja tidak menaikkan versi.

## Tahap 0 — Fondasi

Menyiapkan aplikasi lokal yang stabil, aman, teruji, terdokumentasi, dan siap dirilis sebagai `0.1.0`.

**Status:** berjalan.

## Tahap 1 — Catatan Teks

Menyediakan pengalaman mencatat yang solid: editor, penyimpanan otomatis, pencarian lokal, pengelolaan catatan, ekspor, dan pencadangan.

**Target:** `0.2.0`.

## Tahap 2 — Sumber

Menjadikan buku sebagai tempat mengumpulkan bahan dengan sumber lokal, metadata, impor, kutipan, ekspor, dan pencadangan.

**Target:** `0.3.0`.

## Tahap 3 — Pencarian dan AI Lokal

Menambahkan pencarian yang lebih kuat serta AI lokal/eksternal secara opsional melalui kontrak penyedia yang sederhana, tanpa menghilangkan kemampuan penggunaan tanpa AI.

**Target:** `0.4.0`.

## Tahap 4 — Adopsi Konsep Open Notebook

Mengadopsi konsep yang terbukti berguna—seperti transformasi, konteks, percakapan, dan beberapa penyedia—tanpa menyalin seluruh arsitektur atau kompleksitas Open Notebook.

**Target:** `0.5.0`.

## Tahap 5 — Media Tambahan

Menambahkan media hanya jika manfaatnya jelas dan tetap sesuai dengan ukuran, kesederhanaan distribusi, serta prinsip offline-first.

**Target:** `0.6.0`.

## Tahap 6 — Menuju 1.0.0

Membekukan API, skema data, format pertukaran data, kebijakan migrasi, pengujian, keamanan, performa, dan dokumentasi yang sudah matang.

**Target:** `1.0.0`.

## Setelah 1.0.0

Prioritas beralih ke:
1. perbaikan bug dan keamanan;
2. optimasi kinerja dan ukuran;
3. migrasi basis data yang terencana;
4. fitur baru yang memberi manfaat nyata;
5. pemeliharaan kompatibilitas data dan format publik.

## Kebijakan kualitas

Setiap tahap hanya dapat dilanjutkan jika:
- alur utama tidak merusak data;
- kesalahan ditangani dengan jelas;
- tidak ada rahasia atau data sensitif di repositori;
- aplikasi dasar tetap dapat digunakan secara lokal;
- dependensi baru memiliki alasan yang kuat;
- ukuran dan waktu CI terkendali;
- dokumentasi sesuai implementasi aktual.
