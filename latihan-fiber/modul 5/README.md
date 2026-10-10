# API Students — Authentication & Security (Modul 5)

Proyek REST API untuk manajemen data mahasiswa dan sistem autentikasi pengguna menggunakan framework **Go Fiber v2**, driver PostgreSQL **pgx v5** (`pgxpool`), token **JSON Web Token (JWT)** via `golang-jwt/jwt/v5`, dan hashing kata sandi **bcrypt** via `golang.org/x/crypto/bcrypt`. Proyek ini dibuat untuk memenuhi penugasan **Modul Praktikum Pertemuan 5: Authentication & Security** pada mata kuliah Praktikum Pemrograman Backend Lanjut, Program Studi D4 Teknik Informatika, Fakultas Vokasi, Universitas Airlangga.

---

## 🔐 Fitur Keamanan Utama (Security Features)

1. **Password Hashing (bcrypt cost 12):**
   - Password disimpan dalam bentuk hash irreversible menggunakan bcrypt dengan salt otomatis dan cost factor 12.
   - Tag `json:"-"` pada struct entity memastikan password tidak pernah bocor ke respons JSON.
2. **Dual-Token System (Access Token & Refresh Token):**
   - **Access Token:** Berbentuk JWT (HS256) berumur pendek (15 menit) berisi identitas minimal (`sub`, `username`, `role`), diverifikasi oleh middleware `RequireAuth`.
   - **Refresh Token:** String acak 32-byte kriptografis berumur 7 hari. Disimpan di database dalam bentuk **SHA-256 hash**.
   - **Token Rotation:** Setiap kali refresh token digunakan, token lama dicabut (*revoked*) dan diterbitkan pasangan token baru.
3. **Mitigasi Kerentanan Autentikasi (OWASP Top 10):**
   - **Algorithm Confusion:** Verifikasi tipe algoritma HMAC secara eksplisit di dalam `jwt.Keyfunc` menolak token `alg: "none"` atau pergantian asymmetric algorithm.
   - **Brute Force Protection:** Rate limiter 5 request per menit per alamat IP pada endpoint login, mengembalikan `429 Too Many Requests` disertai header `Retry-After: 60`.
   - **User Enumeration & Timing Attacks:** Pesan error identik `username atau password salah` untuk akun tak terdaftar maupun password salah, ditambah verifikasi hash palsu (`VerifyDummyPassword`) agar durasi komputasi setara.
   - **Mass Assignment:** Struct registrasi tidak menyertakan field `role`; role selalu ditetapkan secara mutlak oleh server (`"user"`).
   - **Denial-of-Service (DoS):** Pembatasan ukuran payload HTTP request (`BodyLimit: 1 MB`).
   - **CORS Ketat:** Origin yang diizinkan dibatasi melalui konfigurasi `ALLOWED_ORIGINS`.

---

## 🏛️ Struktur Direktori Proyek

```text
api-students/
├── app/
│   ├── model/
│   │   ├── student.go              <- Entitas Student, DTO Request & Response
│   │   ├── user.go                 <- Entitas User (Role, json:"-" untuk Password)
│   │   └── auth.go                 <- DTO Registrasi, Login, Refresh, TokenPair
│   ├── repository/
│   │   ├── student_repository.go   <- Repository PostgreSQL untuk mahasiswa
│   │   ├── user_repository.go      <- Repository PostgreSQL untuk user (+ FindByUsername)
│   │   └── token_repository.go     <- Repository PostgreSQL untuk refresh token
│   └── service/
│       ├── student_rules.go        <- Aturan validasi domain mahasiswa murni
│       ├── student_rules_test.go   <- Unit test aturan mahasiswa
│       ├── student_service.go      <- HTTP Controller manajemen mahasiswa
│       ├── auth_rules.go           <- Aturan validasi password & registrasi murni
│       ├── auth_rules_test.go      <- Unit test aturan autentikasi (password strength)
│       └── auth_service.go         <- Controller login, register, refresh, logout, me
├── config/
│   ├── app.go                      <- Konfigurasi Fiber, BodyLimit (1MB), ErrorHandler
│   ├── env.go                      <- Pemuat konfigurasi .env
│   └── logger.go                   <- JSON Structured Logging (slog) + Lumberjack rotation
├── database/
│   └── postgres.go                 <- Pool koneksi pgxpool
├── helper/
│   ├── context.go                  <- Pembaca identitas user dari c.Locals (CurrentUser)
│   ├── jwt.go                      <- JWT Manager (Generate, Parse, Proteksi algorithm confusion)
│   ├── request.go                  <- Parser query string, pagination, context timeout
│   ├── response.go                 <- Presenter amplop standar WebResponse
│   └── security.go                 <- Bcrypt hashing, dummy hash, secure random token, SHA256
├── middleware/
│   ├── auth.go                     <- RequireAuth (JWT verification) & LoginRateLimiter (429)
│   └── middleware.go               <- RequestID, Helmet, CORS ketat, RequestLogger, RequireJSON
├── migrations/
│   ├── 001_create_students.sql     <- Skema tabel students
│   └── 002_auth.sql                <- Alter users add role & skema refresh_tokens
├── main.go                         <- Composition root, validasi secret, graceful shutdown
├── README.md
├── go.mod
└── go.sum
```

---

## 📋 Kontrak Endpoint API

### 1. Endpoint Publik
| Method | Endpoint | Deskripsi | Status Sukses |
| :--- | :--- | :--- | :--- |
| `GET` | `/api/v1/health` | Health check koneksi server dan database | `200 OK` |
| `POST` | `/api/v1/auth/register` | Pendaftaran akun baru (Role mutlak "user") | `201 Created` |
| `POST` | `/api/v1/auth/login` | Login akun (Dibatasi 5x/menit per IP) | `200 OK` |
| `POST` | `/api/v1/auth/refresh` | Rotasi refresh token dan penerbitan token baru | `200 OK` |
| `POST` | `/api/v1/auth/logout` | Pencabutan refresh token | `200 OK` |

### 2. Endpoint Terlindungi (Wajib Header `Authorization: Bearer <access_token>`)
| Method | Endpoint | Deskripsi | Status Sukses |
| :--- | :--- | :--- | :--- |
| `GET` | `/api/v1/auth/me` | Mengambil profil user yang sedang login | `200 OK` |
| `GET` | `/api/v1/students` | Mengambil daftar mahasiswa (Paginasi, Search, Sort) | `200 OK` |
| `GET` | `/api/v1/students/:id` | Mengambil detail mahasiswa berdasarkan ID | `200 OK` |
| `POST` | `/api/v1/students` | Menambahkan data mahasiswa baru | `201 Created` |
| `PUT` | `/api/v1/students/:id` | Mengganti seluruh data mahasiswa | `200 OK` |
| `PATCH`| `/api/v1/students/:id` | Mengubah sebagian data mahasiswa | `200 OK` |
| `DELETE`| `/api/v1/students/:id`| Menghapus mahasiswa | `204 No Content`|

---

## 🧪 Menjalankan Pengujian

Jalankan seluruh rangkaian unit testing:
```bash
go test ./app/service/... -v
```

Jalankan server aplikasi:
```bash
go run main.go
```
