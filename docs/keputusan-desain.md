# Keputusan Desain

## 1. Satu program

**Keputusan:** Go menjadi program utama dan aset web ditanam.

**Alasan:** distribusi sederhana, tidak membutuhkan Node.js saat dipakai, dan cocok untuk aplikasi lokal.

## 2. SQLite

**Keputusan:** SQLite sebagai penyimpanan utama.

**Alasan:** satu berkas, matang, mudah dicadangkan, dan tidak membutuhkan layanan basis data terpisah.

**Konsekuensi:** kemampuan graf dan pencarian vektor tidak tersedia secara bawaan seperti pada SurrealDB. Jika dibutuhkan, tambahkan lapisan kemampuan atau indeks opsional tanpa mengubah data inti.

## 3. SvelteKit

**Keputusan:** SvelteKit untuk antarmuka.

**Alasan:** komponen konsisten, pengalaman pengembangan baik, dan hasil produksi dapat ditanam. Pendekatan ini mirip pola llama.cpp yang menanam UI SvelteKit ke servernya.

## 4. Tidak menyalin seluruh Open Notebook

**Keputusan:** adopsi konsep dan kontrak, bukan seluruh tumpukan.

**Alasan:** target ukuran dan satu eksekusi tidak cocok dengan membawa frontend, API, basis data, dan orkestrasi AI Open Notebook sekaligus.

## 5. AI sebagai modul opsional

**Keputusan:** aplikasi dasar tidak bergantung pada model AI.

**Alasan:** pengguna dapat memakai catatan tanpa mengunduh model atau mengatur kunci layanan.

## 6. Kompatibilitas data

**Keputusan:** migrasi hanya maju dan perubahan destruktif dihindari.

**Alasan:** data pengguna adalah aset utama. Fitur boleh ditunda; kerusakan data tidak boleh diterima.
