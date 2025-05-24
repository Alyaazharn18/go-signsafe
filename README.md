# SignSafe – Simulasi API E-Wallet


SignSafe adalah backend simulasi *e-wallet* berbasis Go yang menggunakan Gorilla Mux sebagai router. API ini mensimulasikan proses registrasi, top-up saldo, transfer antar pengguna, dan riwayat transaksi. Keamanan request dijamin dengan middleware berbasis tanda tangan digital.

## Fitur Utama

- Registrasi Pengguna: Pengguna baru dapat mendaftar dengan mengirimkan userID dan public key. Public key akan disimpan di database untuk verifikasi selanjutnya.
- Otentikasi Tanda Tangan Digital: Setiap permintaan (kecuali registrasi) harus ditandatangani secara digital menggunakan kunci privat pengguna. Server memverifikasi signature menggunakan public key pengguna, memastikan keaslian dan integritas data.
- Middleware SignSafe: Middleware ini memeriksa header khusus (X-UserID, X-Nonce, X-Signature, X-Timestamp) untuk setiap endpoint. Ia memastikan bahwa permintaan fresh dan belum pernah dipakai sebelumnya (mencegah replay attack).
- Transaksi E-Wallet: Endpoint untuk melakukan top-up (penambahan saldo) dan transfer antar pengguna.
- Riwayat Transaksi: Endpoint untuk melihat daftar transaksi (top-up/transfer) milik pengguna.
- Routing dengan Gorilla Mux: Menggunakan paket gorilla/mux yang populer sebagai HTTP router di Go.

## Otentikasi & Middleware SignSafe

SignSafe menggunakan middleware bernama SignSafeMiddleware pada semua endpoint (kecuali /auth/register). Middleware ini memastikan setiap permintaan valid dengan langkah sebagai berikut:

#### Header Wajib: Setiap request harus menyertakan header:

- X-UserID: ID pengguna yang melakukan permintaan.
- X-Timestamp: Stempel waktu pengiriman request (format Unix timestamp). Server menolak permintaan dengan timestamp di luar batas waktu wajar (misal lebih dari ±5 menit dari waktu server).
- X-Nonce: Angka acak unik sekali pakai. Menjamin bahwa setiap permintaan bersifat unik dan mencegah penggunaan ulang permintaan lama (replay attack).
- X-Signature: Tanda tangan digital dari pesan.

#### Normalisasi Body: Body request dinormalisasi ke format JSON kanonik (canonical JSON) agar konsisten. Ini memastikan tanda tangan dapat diverifikasi dengan benar meskipun urutan atau format JSON berubah.

#### Format Pesan Tertandatangani: Middleware membentuk pesan yang akan diverifikasi dengan format:

```
"{userID}|{timestamp}|{nonce}|{canonicalBody}"
```

dimana setiap bagian dipisahkan oleh karakter |. Pesan ini mencakup ID pengguna, timestamp, nonce, dan body request yang sudah dinormalisasi

#### Verifikasi Public Key: Server mengambil public key pengguna dari database (hasil registrasi). Kemudian X-Signature diverifikasi terhadap pesan di atas menggunakan algoritma kriptografi (misalnya RSA/ECDSA). Verifikasi ini memastikan pengirim benar-benar pemilik kunci privat terkait, serta data permintaan tidak diubah.

#### Cek Timestamp: Middleware memeriksa X-Timestamp untuk memastikan permintaan masih segar. Sebagai guideline, permintaan dengan timestamp lebih dari 5 menit yang lalu harus ditolak. Langkah ini mencegah replay attack dalam jangka waktu lama.

#### Cek Nonce: Jika signature valid, server menyimpan X-Nonce di Redis dengan time-to-live. Jika pengguna mencoba mengirim ulang nonce yang sudah pernah dipakai, request ditolak. Dengan cara ini, setiap nonce hanya dapat digunakan sekali saja.

Secara keseluruhan, SignSafeMiddleware menjamin otentikasi dan integritas permintaan. Signature digital memastikan pesan belum diubah dan benar-benar dari pengguna terotentikasi. Kombinasi nonce dan timestamp mencegah serangan replay, yaitu mencegah request lama yang sah dikirim ulang oleh pihak jahat

## Endpoint

Seluruh endpoint diakses dengan prefix /api. Semua request (kecuali register) harus melalui SignSafeMiddleware.

- POST /api/auth/register: Mendaftarkan pengguna baru. Endpoint ini tidak membutuhkan autentikasi karena bertugas menyimpan public key awal.

- GET /api/user: Mendapatkan informasi pengguna saat ini (berdasarkan userID di header).

- GET /api/users/{id}: Mendapatkan informasi pengguna berdasarkan path parameter {id}.

- POST /api/topup: Menambahkan saldo kepada pengguna sendiri.

- POST /api/transfer: Mengirim saldo dari pengguna saat ini ke pengguna lain.

- GET /api/history: Menampilkan riwayat transaksi pengguna saat ini (top-up dan transfer).

## Instalasi dan Penggunaan

```bash
docker build -t signsafe-backend .

docker run -p 8080:8080 --env-file .env signsafe-backend
```

pastikan Redis dan postgresql sudah berjalan
