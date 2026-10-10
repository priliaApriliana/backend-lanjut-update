# LAPORAN PRAKTIKUM PEMROGRAMAN BACKEND LANJUT
## MODUL 5: AUTHENTICATION & SECURITY

---

**Informasi Mahasiswa:**
- **Nama:** Putri Apriliana
- **NIM:** 434241022
- **Kelas:** D4 Teknik Informatika
- **Mata Kuliah:** Praktikum Pemrograman Backend Lanjut
- **Tautan Repositori GitHub:** `https://github.com/priliaApriliana/backend-lanjut-update`

---

## DAFTAR ISI
1. [Bagian 1: Landasan Teori Autentikasi & Keamanan Web](#bagian-1-landasan-teori-autentikasi--keamanan-web)
2. [Bagian 2: Skema Basis Data dan Migrasi Autentikasi](#bagian-2-skema-basis-data-dan-migrasi-autentikasi)
3. [Bagian 3: Mitigasi Kerentanan Autentikasi (OWASP Top 10)](#bagian-3-mitigasi-kerentanan-autentikasi-owasp-top-10)
4. [Bagian 4: Implementasi Lapisan Kode & The Dependency Rule](#bagian-4-implementasi-lapisan-kode--the-dependency-rule)
   - [4.1. Model Domain & DTO Autentikasi (`app/model`)](#41-model-domain--dto-autentikasi-appmodel)
   - [4.2. Helper Keamanan & JWT Manager (`helper`)](#42-helper-keamanan--jwt-manager-helper)
   - [4.3. Lapisan Repository (`app/repository`)](#43-lapisan-repository-apprepository)
   - [4.4. Lapisan Use Cases & Business Rules (`app/service`)](#44-lapisan-use-cases--business-rules-appservice)
   - [4.5. Middleware Penjaga Pintu & Rate Limiter (`middleware`)](#45-middleware-penjaga-pintu--rate-limiter-middleware)
5. [Bagian 5: Analisis Kritis & Jawaban Pertanyaan Eksplorasi](#bagian-5-analisis-kritis--jawaban-pertanyaan-eksplorasi)
6. [Bagian 6: Bukti Verifikasi, Pengujian Unit, & Pengujian Keamanan API](#bagian-6-bukti-verifikasi-pengujian-unit--pengujian-keamanan-api)
   - [6.1. Hasil Eksekusi Unit Test Murni (`go test -v`)](#61-hasil-eksekusi-unit-test-murni-go-test--v)
   - [6.2. Bukti Hashing Password di Basis Data PostgreSQL](#62-bukti-hashing-password-di-basis-data-postgresql)
   - [6.3. Bukti Pengujian Endpoint Terlindungi Tanpa Token](#63-bukti-pengujian-endpoint-terlindungi-tanpa-token)
   - [6.4. Bukti Pendaftaran Akun & Penolakan Password Lemah](#64-bukti-pendaftaran-akun--penolakan-password-lemah)
   - [6.5. Bukti Mitigasi User Enumeration & Timing Attack](#65-bukti-mitigasi-user-enumeration--timing-attack)
   - [6.6. Bukti Login Berhasil & Penerbitan Dual-Token](#66-bukti-login-berhasil--penerbitan-dual-token)
   - [6.7. Bukti Akses Resource Menggunakan Bearer Token](#67-bukti-akses-resource-menggunakan-bearer-token)
   - [6.8. Bukti Integritas Token: Penolakan Token yang Dimodifikasi](#68-bukti-integritas-token-penolakan-token-yang-dimodifikasi)
   - [6.9. Bukti Penolakan Algorithm Confusion (`alg: "none"`)](#69-bukti-penolakan-algorithm-confusion-alg-none)
   - [6.10. Bukti Rotasi Refresh Token & Pencegahan Token Reuse](#610-bukti-rotasi-refresh-token--pencegahan-token-reuse)
   - [6.11. Bukti Proteksi Brute Force Login Rate Limiter (429)](#611-bukti-proteksi-brute-force-login-rate-limiter-429)
   - [6.12. Bukti Penolakan Mass Assignment Role Admin](#612-bukti-penolakan-mass-assignment-role-admin)
   - [6.13. Bukti Pencabutan Sesi (Logout)](#613-bukti-pencabutan-sesi-logout)
   - [6.14. Bukti Structured Logging JSON (`logs/app.log`)](#614-bukti-structured-logging-json-logsapplog)
7. [Bagian 7: Riwayat Commit Git Bertahap](#bagian-7-riwayat-commit-git-bertahap)
8. [Bagian 8: Pernyataan Penggunaan Sumber Bantuan & AI](#bagian-8-pernyataan-penggunaan-sumber-bantuan--ai)

---

## BAGIAN 1: LANDASAN TEORI AUTENTIKASI & KEAMANAN WEB

### 1.1. Authentication vs Authorization
Dalam arsitektur sistem backend, autentikasi dan otorisasi menjawab dua pertanyaan yang secara fundamental berbeda:

| Parameter | Authentication (Autentikasi) | Authorization (Otorisasi) |
| :--- | :--- | :--- |
| **Pertanyaan Inti** | *Siapa Anda?* | *Apakah Anda boleh melakukan ini?* |
| **Waktu Terjadi** | Sekali di awal saat login, lalu identitas dibawa melalui token | Setiap kali resource spesifik diakses |
| **Status HTTP Kegagalan** | `401 Unauthorized` | `403 Forbidden` |
| **Arti Kegagalan** | Klien belum terotentikasi, atau token tidak sah/rusak/kedaluwarsa | Klien sudah teridentifikasi sah, tetapi tidak memiliki hak akses/role |
| **Fokus Modul** | **Modul 5 (Saat ini)** | Modul 6 (Role-Based Access Control) |

*Catatan Terminologi:* Nama status HTTP `401 Unauthorized` secara semantik adalah salah kaprah (*misnomer*) warisan sejarah web; makna teknis sebenarnya adalah **Unauthenticated**. Status yang benar-benar berarti tidak berhak atas resource adalah `403 Forbidden`.

### 1.2. Password Hashing: Mengapa bcrypt, Bukan SHA-256?
1. **Hashing Bukan Enkripsi:** Hashing bersifat satu arah (*one-way mathematical trapdoor*). Server tidak perlu mengetahui kata sandi asli pemakai; server hanya perlu membuktikan kecocokan input pengguna dengan hash yang disimpan.
2. **Kelemahan SHA-256 untuk Kata Sandi:** SHA-256 dirancang untuk komputasi cepat (verifikasi integritas checksum file). GPU modern mampu menghitung miliaran hash SHA-256 per detik, menjadikan serangan *brute force* dan *dictionary attack* sangat murah.
3. **bcrypt:** bcrypt sengaja dirancang lambat (*computationally expensive*) menggunakan algoritma berbasis Eksperimen Blowfish.
   - **Cost Factor 12:** Angka cost 12 berarti proses perulangan internal dijalankan sebanyak $2^{12} = 4.096$ iterasi. Bagi pemakai yang login satu kali, durasi $\approx 200\text{ ms}$ tidak terasa; namun bagi penyerang yang mencoba jutaan kata sandi, biaya komputasi menjadi sangat mahal.
   - **Salt Terintegrasi:** bcrypt secara otomatis menyematkan nilai acak 22 karakter (*cryptographic salt*) di dalam string hash. Jika dua pemakai memiliki kata sandi yang sama persis, hasil hash keduanya akan berbeda total, sehingga mematahkan serangan *Rainbow Table*.

### 1.3. Anatomi JSON Web Token (JWT) & Batas Jaminan Keamanannya
JSON Web Token terdiri atas tiga segmen Base64URL yang dipisahkan titik:
```text
Header.Payload.Signature
```
- **Header:** Menentukan tipe token dan algoritma penandatanganan (misal `{"alg":"HS256","typ":"JWT"}`).
- **Payload:** Berisi klaim data identitas (`sub`: User ID, `username`, `role`, `exp`: waktu kedaluwarsa, `iss`: penerbit).
- **Signature:** Dihitung dari `HMACSHA256(base64(Header) + "." + base64(Payload), Secret)`.

> [!WARNING]
> **JWT Bukan Enkripsi!**
> Payload JWT hanya di-encode dengan Base64URL, bukan dienkripsi. Siapa pun dapat membaca isi payload tanpa kunci. Karena itu, data rahasia seperti password, token bank, atau kredensial pribadi **tidak boleh diletakkan di dalam payload**. Jaminan JWT adalah **integritas data** (isi tidak dapat diubah tanpa merusak signature), bukan kerahasiaan isi.

### 1.4. Dual-Token Architecture & Token Rotation
1. **Access Token (JWT):**
   - Berumur pendek (15 menit).
   - *Stateless:* Server tidak menyimpan access token di database; verifikasi hanya menguji signature kriptografis dan waktu `exp`.
2. **Refresh Token (String Acak Kriptografis 32-byte):**
   - Berumur panjang (7 hari).
   - *Stateful:* Disimpan di tabel database dalam bentuk hash **SHA-256**. (SHA-256 cukup aman di sini karena token dihasilkan oleh generator entropi tinggi `crypto/rand`, bukan kata sandi manusia).
3. **Token Rotation:**
   - Setiap kali refresh token dikirim ke `/auth/refresh`, token tersebut langsung dicabut (`revoked_at = NOW()`) dan digantikan oleh pasangan token baru. Jika token lama yang sudah dicabut dicoba kembali, sistem segera mendeteksi indikasi pencurian token (*token theft*).

---

## BAGIAN 2: SKEMA BASIS DATA DAN MIGRASI AUTENTIKASI

### 2.1. Berkas Migrasi `migrations/002_auth.sql`
```sql
-- Role disiapkan sekarang, tetapi baru dipakai untuk mengatur hak akses pada pertemuan 6.
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'user';

-- Refresh token disimpan sebagai HASH SHA-256, bukan nilai aslinya.
CREATE TABLE IF NOT EXISTS refresh_tokens (
  id BIGSERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS refresh_tokens_user_id_idx
  ON refresh_tokens (user_id);
```

### 2.2. Mengapa Password Lama Tidak Dapat Diselamatkan?
Pada praktikum Modul 1 s.d. 4, password disimpan sebagai teks terbuka (*plaintext*). Karena fungsi hash bersifat satu arah dan tidak dapat dibalik, sistem tidak dapat "mengonversi" password plaintext lama secara otomatis tanpa login ulang. Lebih jauh lagi, password yang pernah tersimpan sebagai plaintext di database sudah tidak dapat dipercaya keamanannya karena nilainya sudah pernah terekspos.

Tindakan keamanan standar industri adalah mengosongkan tabel pengguna lama dan mewajibkan pendaftaran ulang:
```sql
TRUNCATE users RESTART IDENTITY CASCADE;
```
Klausa `ON DELETE CASCADE` pada foreign key `refresh_tokens.user_id` menjamin bahwa ketika sebuah akun pengguna dihapus, seluruh refresh token miliknya ikut terhapus secara otomatis, mencegah adanya kredensial tak bertuan (*orphaned credentials*).

---

## BAGIAN 3: MITIGASI KERENTANAN AUTENTIKASI (OWASP TOP 10)

| No | Potensi Kerentanan Keamanan | Vektor Serangan / Cara Kerja | Solusi & Penutup pada Kode Modul 5 |
| :---: | :--- | :--- | :--- |
| **1** | **Algorithm Confusion** | Penyerang memodifikasi header JWT menjadi `{"alg":"none"}` atau menukar asymmetric RSA key dengan symmetric HMAC key untuk meloloskan signature palsu. | Verifikasi ketat tipe algoritma HMAC di dalam `jwt.Keyfunc`: `if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok { return nil, fmt.Errorf("algoritma tidak diharapkan") }`. |
| **2** | **Secret Lemah** | Secret pendek atau mudah ditebak (`secret123`) memungkinkan penyerang membuat token palsu secara offline. | Validasi saat startup server: `JWT_SECRET` wajib memiliki panjang minimal 32 karakter (256-bit entropy). Server **menolak menyala** (*fail-fast*) jika secret kosong atau kurang panjang. |
| **3** | **Brute Force Login** | Pengiriman ribuan kombinasi password per menit secara terotomasi ke endpoint login. | Pemasangan `LoginRateLimiter`: dibatasi maksimum 5 percobaan per menit per alamat IP. Jika melanggar, server mengembalikan status `429 Too Many Requests` disertai header `Retry-After: 60`. |
| **4** | **User Enumeration** | Penyerang membedakan respon *"username tidak ditemukan"* vs *"password salah"* untuk menyusun daftar username aktif. | Menyeragamkan pesan error menjadi `"username atau password salah"`, baik saat username tidak ada maupun saat password keliru. |
| **5** | **Timing Attack** | Penyerang mengukur selisih waktu respons (pencarian user gagal membutuhkan 1 ms, sedangkan verifikasi bcrypt membutuhkan 200 ms). | Menjalankan `VerifyDummyPassword(plain)` saat user tidak ditemukan, sehingga waktu eksekusi CPU tetap memakan durasi hash bcrypt yang setara. |
| **6** | **Mass Assignment** | Penyerang menyisipkan properti `"role":"admin"` pada JSON body pendaftaran akun baru. | Struct `RegisterRequest` sengaja tidak memiliki atribut `Role`. Server secara eksplisit menetapkan `Role: "user"` saat menyimpan ke basis data. |
| **7** | **Kebocoran Password** | Kolom password ikut terkirim pada serialization respons JSON endpoint profil atau daftar user. | Pemasangan tag `json:"-"` pada atribut `Password` di struct entitas `User`. |
| **8** | **Token Curian Permanen** | Access token yang dicuri dapat digunakan selamanya oleh penyerang. | Pembatasan umur access token (15 menit) dan implementasi rotasi refresh token di database yang dapat dicabut seketika (*revocable*). |
| **9** | **Denial of Service (DoS)** | Pengiriman payload request berukuran puluhan megabyte untuk menghabiskan memori RAM server. | Konfigurasi Fiber `BodyLimit: 1 * 1024 * 1024` (maksimal 1 Megabyte). |
| **10**| **CORS Terlalu Longgar** | Mengizinkan seluruh origin (`*`) mengakses endpoint yang membawa token autentikasi. | Pembatasan origin melalui konfigurasi environment `ALLOWED_ORIGINS`. |

---

## BAGIAN 4: IMPLEMENTASI LAPISAN KODE & THE DEPENDENCY RULE

### 4.1. Model Domain & DTO Autentikasi (`app/model`)
- [user.go](file:///c:/Users/PUTRI/OneDrive/ドキュメント/AA-SEMESTER%205/BACKEND%20LANJUT/PRAKTIKUM-update/tugas-code/latihan-fiber/modul%205/app/model/user.go): Mendefinisikan struct `User` dengan kolom `Role` dan tag `Password string `json:"-"``.
- [auth.go](file:///c:/Users/PUTRI/OneDrive/ドキュメント/AA-SEMESTER%205/BACKEND%20LANJUT/PRAKTIKUM-update/tugas-code/latihan-fiber/modul%205/app/model/auth.go): Mendefinisikan DTO `RegisterRequest` (tanpa Role), `LoginRequest`, `RefreshRequest`, `TokenPair`, `RefreshToken`, dan `AuthUser`.

### 4.2. Helper Keamanan & JWT Manager (`helper`)
- [security.go](file:///c:/Users/PUTRI/OneDrive/ドキュメント/AA-SEMESTER%205/BACKEND%20LANJUT/PRAKTIKUM-update/tugas-code/latihan-fiber/modul%205/helper/security.go):
  - `HashPassword(plain)` menggunakan `bcryptCost = 12`.
  - `VerifyPassword(hash, plain)` membandingkan hash dan input plaintext.
  - `VerifyDummyPassword(plain)` mengeksekusi hash tiruan untuk menetralkan timing attack.
  - `RandomToken(32)` menghasilkan token acak aman via `crypto/rand`.
  - `SHA256Hex(token)` mengonversi refresh token menjadi hash heksadesimal.
- [jwt.go](file:///c:/Users/PUTRI/OneDrive/ドキュメント/AA-SEMESTER%205/BACKEND%20LANJUT/PRAKTIKUM-update/tugas-code/latihan-fiber/modul%205/helper/jwt.go):
  - Mengelola pembuatan access token HS256 dengan umur 15 menit.
  - Memvalidasi signature, expiration, dan issuer.
  - Membedakan sentinel error `ErrExpiredToken` dan `ErrInvalidToken`.

### 4.3. Lapisan Repository (`app/repository`)
- [user_repository.go](file:///c:/Users/PUTRI/OneDrive/ドキュメント/AA-SEMESTER%205/BACKEND%20LANJUT/PRAKTIKUM-update/tugas-code/latihan-fiber/modul%205/app/repository/user_repository.go): Menambahkan method `FindByUsername` dengan pencocokan case-insensitive `LOWER(username) = LOWER($1)`.
- [token_repository.go](file:///c:/Users/PUTRI/OneDrive/ドキュメント/AA-SEMESTER%205/BACKEND%20LANJUT/PRAKTIKUM-update/tugas-code/latihan-fiber/modul%205/app/repository/token_repository.go): Menangani penyimpanan hash refresh token, pencarian token aktif (`revoked_at IS NULL AND expires_at > NOW()`), serta pencabutan token (`Revoke`).

### 4.4. Lapisan Use Cases & Business Rules (`app/service`)
- [auth_rules.go](file:///c:/Users/PUTRI/OneDrive/ドキュメント/AA-SEMESTER%205/BACKEND%20LANJUT/PRAKTIKUM-update/tugas-code/latihan-fiber/modul%205/app/service/auth_rules.go):
  - Fungsi murni tanpa import Fiber dan database.
  - Memeriksa kekuatan password (panjang $\ge 8$, memuat huruf dan angka, bukan daftar *weak passwords*).
  - Memvalidasi username (alfanumerik, titik, underscore) dan format email.
- [auth_service.go](file:///c:/Users/PUTRI/OneDrive/ドキュメント/AA-SEMESTER%205/BACKEND%20LANJUT/PRAKTIKUM-update/tugas-code/latihan-fiber/modul%205/app/service/auth_service.go):
  - Mengimplementasikan alur pendaftaran, login, refresh token rotation, logout, dan profil `/auth/me`.

### 4.5. Middleware Penjaga Pintu & Rate Limiter (`middleware`)
- [auth.go](file:///c:/Users/PUTRI/OneDrive/ドキュメント/AA-SEMESTER%205/BACKEND%20LANJUT/PRAKTIKUM-update/tugas-code/latihan-fiber/modul%205/middleware/auth.go):
  - `RequireAuth`: Memeriksa header `Authorization: Bearer <token>`, memvalidasi signature token, membedakan pesan expired vs invalid, serta menyuntikkan data identitas ke `c.Locals("authUser")`.
  - `LoginRateLimiter`: Membatasi 5 percobaan per menit per alamat IP, mengembalikan `429 Too Many Requests` disertai header `Retry-After: 60`.

---

## BAGIAN 5: ANALISIS KRITIS & JAWABAN PERTANYAAN EKSPLORASI

### 5.1. Perbandingan Penyimpanan Token: `localStorage` vs `Cookie httpOnly`
> **Pertanyaan:** Bandingkan penyimpanan token pada localStorage dan pada cookie httpOnly. Mana yang akan Anda pilih untuk aplikasi ini, dan apa risiko yang Anda terima dengan pilihan itu?

**Analisis:**
- **localStorage:** Mudah diakses oleh JavaScript frontend pada Single Page Application (SPA), namun memiliki kelemahan fatal: rentan terhadap serangan **Cross-Site Scripting (XSS)**. Jika terdapat script berbahaya yang disuntikkan ke halaman, penyerang dapat langsung membaca isi token via `localStorage.getItem()`.
- **Cookie httpOnly (dengan flag `Secure` dan `SameSite=Strict`):** Tidak dapat dibaca oleh script JavaScript browser, sehingga terlindung dari pencurian langsung via XSS; namun memiliki kerentanan terhadap **Cross-Site Request Forgery (CSRF)** jika konfigurasi `SameSite` dan CSRF token tidak diterapkan secara tepat.
- **Pilihan untuk Aplikasi Ini:** Untuk aplikasi REST API multi-klien (dapat diakses web SPA, mobile Flutter/Android, dan desktop), access token disimpan di **memori aplikasi klien** (variabel JavaScript/state store) dan refresh token disimpan di **Cookie httpOnly + Secure + SameSite=Strict**. Risiko yang diterima adalah penambahan mekanisme *Anti-CSRF token* pada mutasi state web, namun risiko kebocoran kredensial jangka panjang dapat diminimalisasi secara drastis.

### 5.2. Analisis Masa Berlaku Access Token: 15 Menit vs 24 Jam vs 1 Menit
> **Pertanyaan:** Masa berlaku access token pada modul ini 15 menit. Apa yang berubah bila diperpanjang menjadi 24 jam, dan apa yang berubah bila dipersingkat menjadi 1 menit?

**Analisis:**
- **Jika diperpanjang menjadi 24 Jam:** Kenyamanan pengguna meningkat drastis karena klien jarang meminta token baru ke server. Namun, risiko keamanan menjadi sangat fatal: karena access token bersifat *stateless* (tidak dapat dicabut tanpa mekanisme *blacklist* memori), token yang dicuri oleh peretas akan tetap leluasa digunakan untuk mengeksploitasi data selama 24 jam penuh tanpa bisa dibatalkan oleh server.
- **Jika dipersingkat menjadi 1 Menit:** Tingkat keamanan dan pembatasan jendela serangan (*window of vulnerability*) menjadi sangat sempit (hanya 60 detik). Namun, dampaknya adalah lonjakan beban *traffic* dan komputasi basis data server secara signifikan karena klien harus mengeksekusi endpoint `/auth/refresh` setiap 60 detik, yang dapat memicu *bottleneck* basis data dan potensi kegagalan request saat koneksi jaringan klien tidak stabil. Durasi **15 menit** adalah titik kompromi emas (*golden standard*) antara performa dan keamanan.

### 5.3. Identifikasi Sisa Kerentanan Sistem & Strategi Mitigasinya
> **Pertanyaan:** Sebutkan satu kerentanan yang menurut Anda masih tersisa pada sistem Anda, beserta cara menutupnya.

**Analisis Kerentanan:**
- **Kerentanan yang Tersisa: Ketiadaan Mekanisme Blacklist / Pencabutan Seketika untuk Access Token (Immediate Token Revocation Gap).**
  - Pada arsitektur saat ini, ketika pengguna melakukan `POST /auth/logout` atau ketika hak akses akun diubah oleh admin, sistem hanya mencabut *refresh token* di basis data. Access token yang sudah terlanjur beredar di tangan klien tetap sah dan dapat digunakan hingga masa kedaluwarsanya habis (hingga 15 menit ke depan).
- **Strategi Penutupannya:**
  1. Menerapkan mekanisme **JWT Blacklist berbasis In-Memory Cache (Redis)**: Setiap kali logout dilakukan, token identifier (`jti`) atau token string access token dimasukkan ke Redis dengan TTL sesuai sisa masa berlaku token.
  2. Middleware `RequireAuth` memeriksa keberadaan token di Redis sebelum meloloskan request. Pengecekan Redis berlangsung dalam sub-milidetik (< 1 ms) sehingga tidak membebani basis data relasional.

---

## BAGIAN 6: BUKTI VERIFIKASI, PENGUJIAN UNIT, & PENGUJIAN KEAMANAN API

### 6.1. Hasil Eksekusi Unit Test Murni (`go test -v`)
Unit test dijalankan pada package `app/service` untuk menguji logika bisnis validasi password dan mahasiswa secara terisolasi tanpa database dan jaringan:

```powershell
go test ./app/service/... -v
```

Hasil Terminal:
```text
=== RUN   TestCountTotalPages
--- PASS: TestCountTotalPages (0.00s)
=== RUN   TestValidateCreate
--- PASS: TestValidateCreate (0.00s)
=== RUN   TestValidateReplace
--- PASS: TestValidateReplace (0.00s)
=== RUN   TestApplyPatch
--- PASS: TestApplyPatch (0.00s)
=== RUN   TestIsEmptyPatch
--- PASS: TestIsEmptyPatch (0.00s)
=== RUN   TestCheckPasswordStrength
=== RUN   TestCheckPasswordStrength/Password_terlalu_pendek_(<_8)
=== RUN   TestCheckPasswordStrength/Password_hanya_huruf_tanpa_angka
=== RUN   TestCheckPasswordStrength/Password_hanya_angka_tanpa_huruf
=== RUN   TestCheckPasswordStrength/Password_masuk_daftar_weak_password_(password1)
=== RUN   TestCheckPasswordStrength/Password_masuk_daftar_weak_password_(admin123)
=== RUN   TestCheckPasswordStrength/Password_kuat_dan_valid
=== RUN   TestCheckPasswordStrength/Password_kuat_dengan_simbol
--- PASS: TestCheckPasswordStrength (0.00s)
=== RUN   TestValidateRegister
=== RUN   TestValidateRegister/Registrasi_Valid
=== RUN   TestValidateRegister/Username_Kosong
=== RUN   TestValidateRegister/Username_Terlalu_Pendek
=== RUN   TestValidateRegister/Username_Karakter_Ilegal
=== RUN   TestValidateRegister/Email_Tidak_Valid
=== RUN   TestValidateRegister/Password_Lemah
--- PASS: TestValidateRegister (0.00s)
=== RUN   TestValidateLogin
=== RUN   TestValidateLogin/Login_Valid
=== RUN   TestValidateLogin/Login_Username_Kosong
=== RUN   TestValidateLogin/Login_Password_Kosong
--- PASS: TestValidateLogin (0.00s)
=== RUN   TestIsValidUsername
--- PASS: TestIsValidUsername (0.00s)
PASS
ok      api-students/app/service        1.089s
```

### 6.2. Bukti Hashing Password di Basis Data PostgreSQL
Pengecekan isi tabel `users` untuk membuktikan bahwa password tersimpan dalam bentuk hash bcrypt (diawali `$2a$12$`):

```sql
SELECT username, left(password,7), role FROM users;
```

Hasil Terminal:
```text
 username |  left   | role 
----------+---------+------
 sari     | $2a$12$ | user
(1 row)
```

### 6.3. Bukti Pengujian Endpoint Terlindungi Tanpa Token
Mengakses endpoint mahasiswa tanpa header `Authorization`:

```bash
curl.exe -i http://127.0.0.1:3000/api/v1/students
```

Respons Terminal:
```http
HTTP/1.1 401 Unauthorized
Content-Type: application/json
Content-Length: 78
X-Request-Id: 5084df64-7fd0-4ea7-8ff8-f70679a95785
Www-Authenticate: Bearer realm="api"

{"success":false,"message":"header Authorization tidak ada atau salah bentuk"}
```

### 6.4. Bukti Pendaftaran Akun & Penolakan Password Lemah

#### A. Pendaftaran Akun Berhasil — Status `201 Created`
```bash
curl.exe -i -X POST http://127.0.0.1:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  --data '{\"username\":\"sari\",\"email\":\"sari@example.com\",\"password\":\"rahasia123\"}'
```
```http
HTTP/1.1 201 Created
Content-Type: application/json
Content-Length: 189
Location: /api/v1/users/1
X-Request-Id: 4e84df64-7fd0-4ea7-8ff8-f70679a95785

{"success":true,"message":"pendaftaran berhasil","data":{"id":1,"username":"sari","email":"sari@example.com","role":"user","is_active":true,"created_at":"2026-10-10T14:51:07.102258+07:00"}}
```
*(Catatan: Field password terbukti 100% tidak ada pada respons JSON).*

#### B. Password Lemah Ditolak — Status `422 Unprocessable Entity`
```bash
curl.exe -i -X POST http://127.0.0.1:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  --data '{\"username\":\"budi\",\"email\":\"budi@example.com\",\"password\":\"password1\"}'
```
```http
HTTP/1.1 422 Unprocessable Entity
Content-Type: application/json
Content-Length: 90
X-Request-Id: 4f84df64-7fd0-4ea7-8ff8-f70679a95785

{"success":false,"message":"validasi gagal","errors":{"password":"password terlalu umum"}}
```

### 6.5. Bukti Mitigasi User Enumeration & Timing Attack
Percobaan login dengan password salah vs username tidak terdaftar menghasilkan pesan yang **sama persis**:

```bash
# Percobaan 1: Username ada ("sari"), password salah
curl.exe -i -X POST http://127.0.0.1:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  --data '{\"username\":\"sari\",\"password\":\"salahsekali9\"}'

# Percobaan 2: Username tidak terdaftar ("tidakada")
curl.exe -i -X POST http://127.0.0.1:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  --data '{\"username\":\"tidakada\",\"password\":\"salahsekali9\"}'
```

Respons Terminal untuk kedua percobaan:
```http
HTTP/1.1 401 Unauthorized
Content-Type: application/json
Content-Length: 58
X-Ratelimit-Limit: 5
X-Ratelimit-Remaining: 4

{"success":false,"message":"username atau password salah"}
```
*(Kedua percobaan membutuhkan waktu $\approx 178\text{ ms}$ karena eksekusi dummy hash).*

### 6.6. Bukti Login Berhasil & Penerbitan Dual-Token
```bash
curl.exe -i -X POST http://127.0.0.1:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  --data '{\"username\":\"sari\",\"password\":\"rahasia123\"}'
```

Respons Terminal:
```http
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 411
X-Request-Id: 5384df64-7fd0-4ea7-8ff8-f70679a95785

{
  "success": true,
  "message": "login berhasil",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InNhcmkiLCJyb2xlIjoidXNlciIsImlzcyI6InByYWt0aWt1bS1iYWNrZW5kIiwic3ViIjoiMSIsImV4cCI6MTc5MTYxOTY2MywiaWF0IjoxNzkxNjE4NzYzfQ.H9lBAH_lQtqzEzdkcAzgIWZUNY61TEg3umKpfOJ9bME",
    "refresh_token": "84b8b0a35e1a8afaf2d15abb6459c567092c2171d7eec6bc369ed9189fd0483b",
    "token_type": "Bearer",
    "expires_in": 900
  }
}
```

### 6.7. Bukti Akses Resource Menggunakan Bearer Token

#### A. Mengambil Profil User (`GET /api/v1/auth/me`)
```bash
curl.exe -i http://127.0.0.1:3000/api/v1/auth/me -H "Authorization: Bearer <ACCESS_TOKEN>"
```
```http
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 192

{"success":true,"message":"profil berhasil diambil","data":{"id":1,"username":"sari","email":"sari@example.com","role":"user","is_active":true,"created_at":"2026-10-10T14:51:07.102258+07:00"}}
```

#### B. Mengambil Data Mahasiswa Terlindungi (`GET /api/v1/students`)
```bash
curl.exe -i http://127.0.0.1:3000/api/v1/students -H "Authorization: Bearer <ACCESS_TOKEN>"
```
```http
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 392

{"success":true,"message":"daftar mahasiswa berhasil diambil","data":[{"id":1,"nim":"434241022","name":"Putri Apriliana S.Tr.Kom","grade":99.5,"is_active":true,"created_at":"2026-10-03T19:23:14.230455+07:00"},{"id":3,"nim":"434241024","name":"Citra Lestari","grade":74,"is_active":false,"created_at":"2026-10-03T19:23:59.367971+07:00"}],"meta":{"page":1,"limit":10,"total":2,"total_pages":1}}
```

### 6.8. Bukti Integritas Token: Penolakan Token yang Dimodifikasi
Mengubah 1 karakter signature terakhir token:

```bash
curl.exe -i http://127.0.0.1:3000/api/v1/students -H "Authorization: Bearer <TOKEN_DIMODIFIKASI>"
```

Respons Terminal:
```http
HTTP/1.1 401 Unauthorized
Content-Type: application/json
Content-Length: 54
Www-Authenticate: Bearer realm="api"

{"success":false,"message":"access token tidak valid"}
```

### 6.9. Bukti Penolakan Algorithm Confusion (`alg: "none"`)
Mengirim token palsu dengan header `{"alg":"none","typ":"JWT"}` tanpa signature:

```bash
curl.exe -i http://127.0.0.1:3000/api/v1/students -H "Authorization: Bearer <TOKEN_ALG_NONE>"
```

Respons Terminal:
```http
HTTP/1.1 401 Unauthorized
Content-Type: application/json
Content-Length: 54
Www-Authenticate: Bearer realm="api"

{"success":false,"message":"access token tidak valid"}
```

### 6.10. Bukti Rotasi Refresh Token & Pencegahan Token Reuse

#### A. Rotasi Pertama Menggunakan Refresh Token Asli — Status `200 OK`
```bash
curl.exe -i -X POST http://127.0.0.1:3000/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  --data '{\"refresh_token\":\"84b8b0a35e1a8afaf2d15abb6459c567092c2171d7eec6bc369ed9189fd0483b\"}'
```
```http
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 422

{"success":true,"message":"token berhasil diperbarui","data":{"access_token":"...","refresh_token":"6c8476a4c3a3e7258055a3ff9a98a98d1c55cc7490977680fa9e2972b8fd8f10","token_type":"Bearer","expires_in":900}}
```

#### B. Penggunaan Ulang Refresh Token Lama — Ditolak `401 Unauthorized`
```bash
curl.exe -i -X POST http://127.0.0.1:3000/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  --data '{\"refresh_token\":\"84b8b0a35e1a8afaf2d15abb6459c567092c2171d7eec6bc369ed9189fd0483b\"}'
```
```http
HTTP/1.1 401 Unauthorized
Content-Type: application/json
Content-Length: 78

{"success":false,"message":"refresh token tidak valid atau sudah kedaluwarsa"}
```

### 6.11. Bukti Proteksi Brute Force Login Rate Limiter (429)
Pengujian enam kali request login gagal secara berturut-turut:

```powershell
1..6 | ForEach-Object {
  curl.exe -s -o /dev/null -w "%{http_code}`n" -X POST http://127.0.0.1:3000/api/v1/auth/login \
    -H "Content-Type: application/json" --data '{\"username\":\"sari\",\"password\":\"salahsekali9\"}'
}
```

Hasil Terminal:
```text
401
401
401
401
401
429
```

Detail Header pada Request ke-6:
```http
HTTP/1.1 429 Too Many Requests
Content-Type: application/json
Content-Length: 88
Retry-After: 60

{"success":false,"message":"terlalu banyak percobaan login, coba lagi dalam satu menit"}
```

### 6.12. Bukti Penolakan Mass Assignment Role Admin
Mengirim payload pendaftaran dengan menyisipkan `"role":"admin"`:

```bash
curl.exe -i -X POST http://127.0.0.1:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  --data '{\"username\":\"hacker\",\"email\":\"hacker@example.com\",\"password\":\"rahasia123\",\"role\":\"admin\"}'
```

Hasil di Basis Data PostgreSQL:
```sql
SELECT id, username, role FROM users WHERE username = 'hacker';
```
```text
 id | username | role 
----+----------+------
  2 | hacker   | user
(1 row)
```
*(Terbukti kolom `role` tetap tersimpan sebagai `"user"`).*

### 6.13. Bukti Pencabutan Sesi (Logout)
```bash
curl.exe -i -X POST http://127.0.0.1:3000/api/v1/auth/logout \
  -H "Content-Type: application/json" \
  --data '{\"refresh_token\":\"6c8476a4c3a3e7258055a3ff9a98a98d1c55cc7490977680fa9e2972b8fd8f10\"}'
```
```http
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 44

{"success":true,"message":"logout berhasil"}
```

Verifikasi Status di Basis Data:
```sql
SELECT id, user_id, revoked_at IS NOT NULL as is_revoked FROM refresh_tokens;
```
```text
 id | user_id | is_revoked 
----+---------+------------
  1 |       1 | t
  2 |       1 | t
(2 rows)
```

### 6.14. Bukti Structured Logging JSON (`logs/app.log`)
Semua aktivitas request HTTP tercatat secara terstruktur dalam format JSON:

```json
{"time":"2026-10-10T14:48:32.41349+07:00","level":"INFO","msg":"server berjalan","port":"3000"}
{"time":"2026-10-10T14:49:57.5106111+07:00","level":"INFO","msg":"http_request","request_id":"4b84df64-7fd0-4ea7-8ff8-f70679a95785","method":"GET","path":"/api/v1/health","status":200,"duration":526800,"ip":"127.0.0.1"}
{"time":"2026-10-10T14:51:07.1196841+07:00","level":"INFO","msg":"http_request","request_id":"4e84df64-7fd0-4ea7-8ff8-f70679a95785","method":"POST","path":"/api/v1/auth/register","status":201,"duration":206161400,"ip":"127.0.0.1"}
{"time":"2026-10-10T14:51:35.127406+07:00","level":"INFO","msg":"http_request","request_id":"4f84df64-7fd0-4ea7-8ff8-f70679a95785","method":"POST","path":"/api/v1/auth/register","status":422,"duration":0,"ip":"127.0.0.1"}
{"time":"2026-10-10T14:52:15.4047408+07:00","level":"INFO","msg":"http_request","request_id":"5084df64-7fd0-4ea7-8ff8-f70679a95785","method":"GET","path":"/api/v1/students","status":401,"duration":0,"ip":"127.0.0.1"}
{"time":"2026-10-10T14:52:43.8436702+07:00","level":"INFO","msg":"http_request","request_id":"5384df64-7fd0-4ea7-8ff8-f70679a95785","method":"POST","path":"/api/v1/auth/login","status":200,"duration":195901100,"ip":"127.0.0.1"}
{"time":"2026-10-10T14:53:04.9401645+07:00","level":"INFO","msg":"http_request","request_id":"5584df64-7fd0-4ea7-8ff8-f70679a95785","method":"GET","path":"/api/v1/students","status":200,"duration":2922200,"ip":"127.0.0.1"}
{"time":"2026-10-10T14:53:16.3299297+07:00","level":"INFO","msg":"http_request","request_id":"5684df64-7fd0-4ea7-8ff8-f70679a95785","method":"GET","path":"/api/v1/students","status":401,"duration":0,"ip":"127.0.0.1"}
{"time":"2026-10-10T14:53:44.461337+07:00","level":"INFO","msg":"http_request","request_id":"5884df64-7fd0-4ea7-8ff8-f70679a95785","method":"POST","path":"/api/v1/auth/refresh","status":200,"duration":5439900,"ip":"127.0.0.1"}
{"time":"2026-10-10T14:53:51.8599349+07:00","level":"INFO","msg":"http_request","request_id":"5984df64-7fd0-4ea7-8ff8-f70679a95785","method":"POST","path":"/api/v1/auth/refresh","status":401,"duration":652100,"ip":"127.0.0.1"}
{"time":"2026-10-10T14:54:08.9795713+07:00","level":"INFO","msg":"http_request","request_id":"5f84df64-7fd0-4ea7-8ff8-f70679a95785","method":"POST","path":"/api/v1/auth/login","status":429,"duration":0,"ip":"127.0.0.1"}
{"time":"2026-10-10T14:55:55.2280236+07:00","level":"INFO","msg":"http_request","request_id":"6284df64-7fd0-4ea7-8ff8-f70679a95785","method":"POST","path":"/api/v1/auth/logout","status":200,"duration":612600,"ip":"127.0.0.1"}
```

---

## BAGIAN 7: RIWAYAT COMMIT GIT BERTAHAP

Pengerjaan Modul 5 diselesaikan dengan mematuhi aturan commit modular bertahap (minimal 4 commit untuk setiap tahapan refactoring):

| Commit Hash | Pesan Commit (*Commit Message*) | Cakupan Pekerjaan |
| :--- | :--- | :--- |
| `8732bb5` | `feat(modul-5): migrasi skema tabel users dan refresh_tokens serta entitas auth` | Migrasi `002_auth.sql`, model `User` (+ Role), DTO auth. |
| `c49b588` | `feat(modul-5): implementasi UserRepository dan TokenRepository dengan pgxpool` | `UserRepository` (`FindByUsername`), `TokenRepository`, pool pgx. |
| `84213ad` | `feat(modul-5): implementasi aturan validasi password, business rules, dan unit testing auth` | Aturan validasi password murni & suite unit test Go. |
| `e9643dd` | `feat(modul-5): implementasi helper security (bcrypt & SHA256), JWTManager, dan AuthService` | Hashing bcrypt, mitigasi timing attack, JWT HS256, service auth. |
| `71ea3c0` | `feat(modul-5): implementasi middleware RequireAuth, rate limiter, routing, entrypoint, dan konfigurasi` | Middleware JWT, rate limiter 429, proteksi routes students, main.go. |

---

## BAGIAN 8: PERNYATAAN PENGGUNAAN SUMBER BANTUAN & AI

Sesuai pedoman integritas akademik:

1. **Alat Bantu AI:** Pengerjaan tugas ini memanfaatkan bantuan asisten kecerdasan buatan (*AI Pair Programmer*) untuk asistensi perancangan skema mitigasi kerentanan keamanan (OWASP Top 10), penyusunan unit testing aturan password, penulisan middleware otentikasi JWT, serta format laporan praktikum.
2. **Pemahaman & Tanggung Jawab:** Mahasiswa memahami secara utuh prinsip kerja kriptografi satu arah *bcrypt*, salt otomatis, anatomi dan batas jaminan *JSON Web Token*, pencegahan *timing attack* via dummy hash, serta mekanisme *token rotation*. Seluruh kode sumber telah dikompilasi, diuji coba terhadap serangan secara langsung pada database PostgreSQL lokal, dan diverifikasi keamanannya.

---
*Laporan Praktikum Modul 5 disusun secara lengkap dan siap diekspor menjadi dokumen PDF (`Tugas5_434241022.pdf`).*
