# Roadmap

## Fase 0 — Fondasi

- [x] Satu proses Go.
- [x] SQLite lokal.
- [x] Migrasi berurutan.
- [x] UI SvelteKit tertanam.
- [x] Model buku, sumber, catatan, percakapan, pesan.
- [ ] Uji otomatis dan pengukuran ukuran rilis.

## Fase 1 — Catatan teks

- [x] Buku.
- [x] Catatan teks.
- [ ] Pengeditan nyaman dengan pintasan papan ketik.
- [ ] Pencarian lokal cepat.
- [ ] Penyimpanan otomatis dengan indikator status.
- [ ] Riwayat perubahan sederhana.
- [ ] Ekspor Markdown dan JSON.

## Fase 2 — Sumber

- [ ] Sumber teks sebagai objek terpisah dari catatan.
- [ ] Tautan sumber ke buku.
- [ ] Impor Markdown/TXT.
- [ ] Impor URL sebagai fitur opsional.
- [ ] Pratinjau dan metadata sumber.

## Fase 3 — AI lokal/eksternal

- [ ] Kontrak model OpenAI-compatible.
- [ ] Dukungan llama-server.
- [ ] Dukungan Ollama.
- [ ] Percakapan per buku.
- [ ] Rujukan jawaban ke sumber.
- [ ] Pencarian semantik opsional.

## Fase 4 — Adopsi fitur Open Notebook

Prioritas berdasarkan nilai dan ukuran, bukan jumlah fitur.

- [ ] Transformasi teks.
- [ ] Konteks sumber yang dapat dipilih.
- [ ] Beberapa sesi percakapan.
- [ ] Penyedia model jamak.
- [ ] Impor/ekspor yang kompatibel secara konseptual.
- [ ] Fitur lanjutan hanya setelah ukuran dan pemeliharaan tetap sehat.

## Fase 5 — Media

- [ ] Lampiran dokumen.
- [ ] Gambar.
- [ ] Audio/video bila manfaatnya jelas.
- [ ] Ekstraksi isi melalui modul opsional agar program dasar tetap kecil.

## Pagar kualitas

Setiap fase harus menjaga:

- data lama tetap dapat dibuka;
- migrasi dapat diuji dan dipulihkan;
- ukuran rilis di bawah 30 MB;
- target normal di bawah 20 MB;
- tidak ada rahasia atau data sensitif di repositori;
- CI tetap singkat;
- dokumentasi diperbarui bersamaan dengan perubahan.
