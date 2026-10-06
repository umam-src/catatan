# Hubungan sumber dan catatan

## Keputusan Fase 2

Fase 2 tidak menambahkan tabel hubungan langsung antara sumber dan catatan.

Keputusan ini diambil setelah meninjau alur yang sudah tersedia:

- sumber dimiliki oleh buku;
- catatan dimiliki oleh buku;
- penanda lokasi saat ini melekat pada sumber dan dapat diverifikasi melalui checksum;
- belum ada alur kutipan yang menyimpan atau menampilkan hubungan sumber tertentu dengan catatan tertentu;
- menambah tabel hubungan sekarang hanya akan menambah migrasi, indeks, API, pengujian, dan biaya pemeliharaan tanpa kebutuhan pengguna yang sudah terbukti.

## Dampak

Tidak ada data sumber yang diduplikasi. Sumber tetap disimpan satu kali pada buku pemiliknya.

Isolasi tetap mengikuti batas kepemilikan buku. Jika pada fase berikutnya kutipan catatan membutuhkan hubungan eksplisit, relasi dapat ditambahkan sebagai perubahan skema tersendiri dengan aturan kepemilikan dan penghapusan berantai yang jelas.

## Batas keputusan

Keputusan ini bukan larangan permanen. Hubungan sumber-catatan perlu ditinjau kembali ketika salah satu alur berikut benar-benar diimplementasikan:

- kutipan sumber yang melekat pada catatan;
- daftar sumber yang digunakan oleh suatu catatan;
- penandaan lokasi sumber dari dalam editor catatan;
- kebutuhan lain yang memerlukan hubungan sumber dan catatan secara persisten.

Sampai kebutuhan tersebut terbukti, model yang lebih sederhana dipertahankan untuk menjaga ukuran, performa, dan pemeliharaan aplikasi.
