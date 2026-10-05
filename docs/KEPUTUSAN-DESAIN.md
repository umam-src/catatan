# Keputusan Desain

## Identitas dan autentikasi

Fase 0 menggunakan identitas pengguna internal yang stabil. Identitas tersebut dipisahkan dari cara pengguna masuk sehingga metode autentikasi dapat diganti tanpa mengubah pemilik data.

### Implementasi awal

- Akun lokal disimpan di basis data SQLite.
- Kata sandi tidak pernah disimpan dalam bentuk mentah.
- Hash kata sandi menggunakan PBKDF2-HMAC-SHA-256 dengan salt acak dan 600.000 iterasi.
- Session menggunakan token acak yang hanya disimpan dalam basis data dalam bentuk hash SHA-256.
- Cookie session memiliki HttpOnly dan SameSite=Strict; Secure aktif saat koneksi menggunakan TLS.
- Session berlaku 24 jam dan dapat dicabut melalui logout.
- Percobaan login dibatasi sementara setelah kegagalan berulang.
- Data buku memiliki owner_id; catatan memperoleh kepemilikan melalui buku yang dimilikinya.
- Endpoint data selalu memeriksa session dan kepemilikan sebelum membaca atau mengubah data.

## Batas provider autentikasi

Metode autentikasi lokal diperlakukan sebagai provider, sedangkan identitas pengguna tetap menjadi entitas internal. Desain ini sengaja tidak menyimpan atribut LDAP, Active Directory, Synology Directory Server, OIDC, atau SAML sebagai identitas utama.

Integrasi provider eksternal pada fase berikutnya harus memetakan akun eksternal ke pengguna internal. Jika provider eksternal tidak tersedia, instalasi yang mengaktifkan login lokal tetap dapat menggunakan provider lokal.

LDAP/AD/Synology, OIDC, dan SAML bukan bagian implementasi Fase 0. Fase ini hanya menetapkan batas agar integrasinya tidak memerlukan perubahan mendasar pada identitas dan kepemilikan data.

## Privasi dan log

Isi catatan, kata sandi, token session, cookie session, dan kredensial provider tidak ditulis ke log. Kesalahan API yang dikirim kepada klien juga tidak mengekspos pesan kesalahan basis data.

## Migrasi

Perubahan autentikasi masuk melalui migrasi baru dan tidak mengubah migrasi yang sudah ada. Data buku awal diberi pemilik internal local agar data lama tetap memiliki pemilik setelah migrasi.

## Authorization

Model otorisasi Catatan menggunakan ACL sebagai fondasi.

### Keputusan

- Pola dasar: User/Group → ACL → Resource.
- ACL menyimpan hanya pemberian izin yang benar-benar diperlukan pada sumber daya; tidak membuat matriks massal pengguna × sumber daya × permission.
- Subjek ACL dapat berupa pengguna atau grup melalui subject_type dan subject_id.
- Permission menggunakan pengenal kecil dan terstandar, misalnya note.read, note.write, note.delete, notebook.read, dan notebook.write.
- Hierarki sumber daya dapat ditambahkan kemudian melalui parent resource dan inheritance. Inheritance bukan persyaratan fondasi Fase 0.
- Role tidak menjadi lapisan wajib. Role hanya dipertimbangkan jika pola penggunaan nyata menunjukkan manfaat yang jelas.
- ABAC, policy kompleks, dan policy engine eksternal tidak menjadi bagian fondasi Catatan.

### Prinsip performa

Beban utama authorization bukan ukuran tabel ACL, melainkan jumlah relasi dan evaluasi yang harus dilakukan saat menentukan effective permission. Karena itu lookup harus didukung indeks yang tepat dan tidak menelusuri hierarki berulang tanpa kebutuhan.

Jika inheritance digunakan, effective permission dapat di-cache dan cache dibatalkan ketika membership grup, ACL, atau hierarki sumber daya berubah.

Desain ini dipilih agar authorization tetap ringan untuk SQLite, sesuai dengan prinsip offline-first, dan dapat dikembangkan tanpa mengganti identitas internal atau kepemilikan data.
