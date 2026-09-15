# API Students

REST API untuk mengelola data mahasiswa, dibangun menggunakan Go dan Fiber v2.

- **Nama:** Putri Apriliana
- **NIM:** 434241022
- **Repositori:** https://github.com/priliaApriliana/backend-lanjut
- **Lokasi proyek:** `tugas-mandiri/api-students`

---

### Modul 2 — REST API Dasar

Membangun fondasi API dari nol menggunakan Go dan Fiber v2:

- Setup proyek Go dengan Fiber v2 sebagai HTTP framework.
- Membuat CRUD lengkap (GET list, GET by ID, POST, PUT, PATCH, DELETE) untuk entitas Student.
- Data disimpan **in-memory** menggunakan slice — belum ada database.
- Menerapkan response envelope yang konsisten (`success`, `message`, `data`, `meta`, `errors`).
- Fitur tambahan: paginasi, pencarian nama, sorting, filter `is_active` dan rentang `grade`.
- Middleware `requireJSON` untuk menolak request tanpa `Content-Type: application/json`.
- Validasi input (NIM dan nama wajib diisi, grade 0–4.0) dengan error per field (status 422).

### Modul 3 — Database & Repository Pattern

Memindahkan penyimpanan data dari in-memory ke PostgreSQL:

- Menambahkan PostgreSQL sebagai penyimpanan permanen, menggantikan slice in-memory.
- Menggunakan driver `pgx/v5` dengan connection pool (`pgxpool`) untuk efisiensi koneksi.
- Menerapkan **repository pattern**: interface `StudentRepository` memisahkan kontrak akses data dari implementasi PostgreSQL.
- Semua query menggunakan parameterized query (`$1`, `$2`, ...) untuk mencegah SQL injection.
- Sorting menggunakan daftar putih nama kolom karena `ORDER BY` tidak bisa diparameterkan.
- Duplikasi NIM dijaga oleh `UNIQUE INDEX` di database, bukan pengecekan manual di kode Go.
- Menambahkan endpoint `/api/v1/health` untuk cek koneksi database (status 503 jika gagal).
- Konfigurasi environment melalui file `.env` dengan package `godotenv`.

### Modul 4 — Clean Architecture

Merestrukturisasi proyek ke dalam folder-folder terpisah sesuai tanggung jawab, tanpa mengubah perilaku HTTP:

- **Restrukturisasi folder**: kode yang sebelumnya terkumpul di `main.go`, `handler.go`, dan `helper.go` dipecah ke dalam package `config`, `database`, `helper`, `middleware`, `route`, dan sub-package `app/model`, `app/repository`, `app/service`.
- **Pemisahan business rules**: fungsi validasi (`ValidateCreate`, `ValidateReplace`, `ApplyPatch`) dipindahkan ke file `student_rules.go` yang **murni** — tidak mengimpor Fiber maupun SQL.
- **Unit test**: 5 test untuk fungsi business rules (`TestValidateCreate`, `TestValidateReplace`, `TestApplyPatch`, `TestIsEmptyPatch`, `TestCountTotalPages`), semua PASS.
- **Logger terstruktur**: menggunakan `log/slog` dengan format JSON, menulis ke layar (stdout) dan file `logs/app.log` sekaligus. Rotasi file otomatis menggunakan `lumberjack` (maks 10 MB, 5 backup, 14 hari).
- Setiap log request mencatat: `request_id`, `method`, `path`, `status`, `duration`, dan `ip`.
- `main.go` disederhanakan menjadi hanya **urutan perakitan** — tidak ada handler, query, atau logika bisnis.
- File `.env` dan folder `logs/` tidak ikut ter-commit (sudah masuk `.gitignore`).

### Modul 5 — Authentication & Security

Menambahkan sistem autentikasi lengkap. Seluruh endpoint `/students` kini tertutup bagi pemakai yang belum login; `/health` tetap publik:

- **Tabel `users` dan `refresh_tokens` baru**, terpisah dari tabel `students` — `users` khusus untuk akun login, `students` tetap resource data mahasiswa yang sekarang dilindungi.
- **Password di-hash dengan bcrypt** (cost factor 12), tidak pernah disimpan sebagai plaintext. `VerifyDummyPassword` dijalankan saat username tidak ditemukan, agar waktu respons tetap konsisten dengan kasus password salah (mencegah *timing attack*).
- **JWT (access token)** ditandatangani dengan HS256, berumur 15 menit. Signing method diperiksa eksplisit saat parsing untuk mencegah *algorithm confusion* (serangan `alg:none`).
- **Refresh token** berupa string acak (`crypto/rand`, 32 byte), disimpan di database dalam bentuk hash SHA-256, berumur 7 hari, dan **dirotasi wajib**: setiap kali dipakai, token lama langsung dicabut dan diganti yang baru.
- **Mass assignment dicegah**: struct `RegisterRequest` tidak memiliki field `role` — role akun baru selalu di-set `"user"` oleh server, tidak pernah diambil dari input klien.
- **User enumeration dicegah**: pesan error login untuk "username tidak ditemukan" dan "password salah" dibuat identik.
- **Rate limiting** pada endpoint login: maksimal 5 percobaan per menit per alamat IP, percobaan berikutnya mendapat status 429 dengan header `Retry-After`.
- **Middleware `RequireAuth`** memeriksa header `Authorization: Bearer <token>`, membedakan pesan token kedaluwarsa dan token tidak valid, serta menyertakan header `WWW-Authenticate` pada setiap penolakan.
- **CORS dibatasi** lewat environment variable `ALLOWED_ORIGINS` (sebelumnya mengizinkan semua origin), dan `BodyLimit` diset 1 MB untuk mencegah payload raksasa.
- **Business rules murni** (`auth_rules.go`) memvalidasi kekuatan password (minimal 8 karakter, harus ada huruf dan angka, menolak daftar password umum), diuji lewat unit test terpisah dari HTTP handler.
- Seluruh sepuluh skenario keamanan (register, login, akses tanpa token, token dimodifikasi, rotasi refresh token, brute force, dll.) diuji langsung terhadap server yang berjalan dan lolos.

---

## Cara Menyiapkan dari Nol

1. Pastikan PostgreSQL dan Go (1.26+) sudah terinstall.

2. Clone repositori dan masuk ke folder proyek ini:
```bash
   git clone https://github.com/priliaApriliana/backend-lanjut.git
   cd backend-lanjut/tugas-mandiri/api-students
```

3. Buat database:
```bash
   psql -U postgres -c "CREATE DATABASE tugas_mandiri_students;"
```

4. Jalankan migrasi (urut, keduanya wajib):
```bash
   psql -U postgres -d tugas_mandiri_students -f migrations/001_create_students.sql
   psql -U postgres -d tugas_mandiri_students -f migrations/002_auth.sql
```

5. Salin `.env.example` menjadi `.env`, lalu isi nilainya sesuai konfigurasi PostgreSQL lokal. Khusus `JWT_SECRET`, buat nilai acak sendiri (jangan dikarang manual):
```bash
   cp .env.example .env
   openssl rand -hex 32
```
   Tempel hasilnya ke `JWT_SECRET` di `.env`. Server akan menolak menyala bila `JWT_SECRET` kosong atau kurang dari 32 karakter.

6. Install dependency dan jalankan:
```bash
   go mod tidy
   go run .
```

7. Server berjalan di `http://localhost:3000`. Cek dengan:
```bash
   curl http://localhost:3000/api/v1/health
```

## Variabel Environment

| Variabel | Keterangan | Contoh |
|---|---|---|
| `APP_PORT` | Port server HTTP | `3000` |
| `DB_HOST` | Host PostgreSQL | `localhost` |
| `DB_PORT` | Port PostgreSQL | `5432` |
| `DB_USER` | Username PostgreSQL | `postgres` |
| `DB_PASSWORD` | Password PostgreSQL | *(isi sendiri)* |
| `DB_NAME` | Nama database | `tugas_mandiri_students` |
| `DB_SSLMODE` | Mode SSL koneksi | `disable` (untuk lokal) |
| `DB_MAX_CONNS` | Jumlah koneksi maksimum di connection pool | `10` |
| `LOG_LEVEL` | Level log minimum (`debug`, `info`, `warn`, `error`) | `info` |
| `JWT_SECRET` | Kunci penandatangan JWT, minimal 32 karakter acak | *(hasil `openssl rand -hex 32`)* |
| `JWT_ISSUER` | Nilai klaim `iss` pada token | `api-students` |
| `JWT_ACCESS_TTL_MINUTES` | Masa berlaku access token (menit) | `15` |
| `JWT_REFRESH_TTL_DAYS` | Masa berlaku refresh token (hari) | `7` |
| `ALLOWED_ORIGINS` | Origin yang diizinkan mengakses API (CORS) | `http://localhost:5173` |

## Skema Tabel

```sql
CREATE TABLE IF NOT EXISTS students (
    id          SERIAL PRIMARY KEY,
    nim         VARCHAR(20) NOT NULL,
    name        VARCHAR(100) NOT NULL,
    grade       NUMERIC(3,2) NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS students_nim_key
    ON students (nim);

CREATE INDEX IF NOT EXISTS students_name_lower_idx
    ON students (LOWER(name));

-- Modul 5: tabel akun login, terpisah dari data mahasiswa
CREATE TABLE IF NOT EXISTS users (
    id          SERIAL PRIMARY KEY,
    username    VARCHAR(50) NOT NULL UNIQUE,
    email       VARCHAR(255) NOT NULL UNIQUE,
    password    TEXT NOT NULL,
    role        VARCHAR(20) NOT NULL DEFAULT 'user',
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Modul 5: refresh token disimpan sebagai hash, bukan nilai aslinya
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id          BIGSERIAL PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS refresh_tokens_user_id_idx
    ON refresh_tokens (user_id);
```

## Struktur Proyek

```
api-students/
├── app/
│   ├── model/
│   │   ├── student.go              entitas Student, struct request, dan ListQuery
│   │   ├── response.go             struct Response dan Meta
│   │   ├── user.go                 entitas User (akun login) — Modul 5
│   │   └── auth.go                 struct request/response autentikasi — Modul 5
│   ├── repository/
│   │   ├── student_repository.go   kontrak (interface) dan implementasi akses data (pgx)
│   │   ├── user_repository.go      akses data tabel users — Modul 5
│   │   └── token_repository.go     akses data tabel refresh_tokens — Modul 5
│   └── service/
│       ├── student_service.go      handler HTTP yang memanggil repository
│       ├── student_rules.go        business rules murni (validasi, patch) — tanpa Fiber/SQL
│       ├── student_rules_test.go   unit test untuk business rules
│       ├── auth_service.go         handler HTTP: register, login, refresh, logout, me — Modul 5
│       ├── auth_rules.go           business rules murni autentikasi — Modul 5
│       └── auth_rules_test.go      unit test untuk business rules autentikasi — Modul 5
├── config/
│   ├── app.go                      perakitan aplikasi Fiber (middleware + route)
│   ├── env.go                      pembaca variabel environment (.env)
│   └── logger.go                   logger JSON terstruktur + rotasi file
├── database/
│   └── postgres.go                 koneksi dan connection pool PostgreSQL
├── helper/
│   ├── request.go                  pembaca query string dan parameter path
│   ├── response.go                 response envelope (Ok, Created, Fail, dll.)
│   ├── security.go                 hashing password (bcrypt) dan token acak — Modul 5
│   ├── jwt.go                      pembuatan dan verifikasi JWT — Modul 5
│   └── context.go                  pembaca identitas pengguna dari Locals — Modul 5
├── middleware/
│   ├── middleware.go               middleware global (requestid, logger, CORS) + RequireJSON
│   └── auth.go                     RequireAuth dan rate limiter login — Modul 5
├── route/
│   └── route.go                    pendaftaran route ke handler
├── migrations/
│   ├── 001_create_students.sql
│   └── 002_auth.sql                tabel users dan refresh_tokens — Modul 5
├── logs/                           output log (tidak di-commit)
├── .env                            konfigurasi environment (tidak di-commit)
├── .env.example
├── main.go                         titik masuk — hanya urutan perakitan
├── go.mod
└── go.sum
```

## Format Respons

Semua endpoint pakai format respons yang sama:

```json
{
  "success": true,
  "message": "...",
  "data": {},
  "meta": {},
  "errors": {}
}
```

| Field     | Tipe    | Keterangan                                   |
|-----------|---------|-----------------------------------------------|
| `success` | boolean | `true` kalau berhasil, `false` kalau gagal   |
| `message` | string  | Pesan singkat                                |
| `data`    | any     | Isi data (muncul kalau ada)                  |
| `meta`    | object  | Info paginasi (cuma di endpoint daftar)       |
| `errors`  | object  | Detail error tiap field (cuma di 422)         |

## Entitas Student

| Field        | Tipe      | Keterangan                   |
|--------------|-----------|-------------------------------|
| `id`         | int       | ID unik, auto-increment (dari `SERIAL` PostgreSQL) |
| `nim`        | string    | NIM, harus unik (dijaga `UNIQUE INDEX`) |
| `name`       | string    | Nama mahasiswa                |
| `grade`      | float64   | IPK, rentang 0–4.0             |
| `is_active`  | bool      | Status aktif/nonaktif          |
| `created_at` | datetime  | Waktu data dibuat (diisi otomatis oleh database) |

## Entitas User (Modul 5)

| Field        | Tipe      | Keterangan                   |
|--------------|-----------|-------------------------------|
| `id`         | int       | ID unik, auto-increment        |
| `username`   | string    | Username, harus unik           |
| `email`      | string    | Email, harus unik              |
| `password`   | string    | Hash bcrypt — tidak pernah keluar sebagai JSON (`json:"-"`) |
| `role`       | string    | Selalu `"user"`, ditentukan server saat register |
| `is_active`  | bool      | Status aktif/nonaktif           |
| `created_at` | datetime  | Waktu akun dibuat               |

## Endpoint

### Autentikasi (Modul 5)

| Method | Endpoint | Auth? | Keterangan |
|---|---|---|---|
| POST | `/api/v1/auth/register` | publik | Registrasi akun baru. Role selalu `"user"`, status 201 |
| POST | `/api/v1/auth/login` | publik, dibatasi 5x/menit/IP | Login, mengembalikan `access_token` dan `refresh_token` |
| POST | `/api/v1/auth/refresh` | publik | Menukar refresh token lama dengan pasangan token baru (rotasi) |
| POST | `/api/v1/auth/logout` | publik | Mencabut refresh token yang dikirim |
| GET | `/api/v1/auth/me` | wajib token | Profil pengguna yang sedang login |

**Body `POST /auth/register`:**
```json
{ "username": "sari", "email": "sari@example.com", "password": "rahasia123" }
```

**Respons login `200`:**
```json
{
  "success": true,
  "message": "login berhasil",
  "data": {
    "access_token": "eyJhbGciOi...",
    "refresh_token": "fec29c135a..."
  }
}
```

Seluruh endpoint di bawah ini **wajib menyertakan header** `Authorization: Bearer <access_token>` sejak Modul 5. Tanpa header ini, permintaan ditolak dengan status 401.

### 1. GET /api/v1/students

| Item             | Detail                              |
|------------------|--------------------------------------|
| **Metode**       | GET                                  |
| **Endpoint**     | `/api/v1/students`                   |
| **Auth**         | Wajib token                          |
| **Status Sukses**| 200                                  |

**Query Parameter (seluruhnya dikerjakan di sisi database, bukan di Go):**

| Parameter   | Default | Keterangan                                        |
|-------------|---------|-----------------------------------------------------|
| `page`      | 1       | Halaman ke berapa                                  |
| `limit`     | 10      | Jumlah per halaman (maks 100)                      |
| `search`    | —       | Cari nama, `ILIKE` case-insensitive                |
| `sort`      | id      | Urutkan: `id`, `nim`, `name`, `grade`, `created_at`; pakai daftar putih kolom |
| `order`     | asc     | `asc` atau `desc`                                  |
| `is_active` | —       | Filter: `true` / `false`                           |
| `grade`     | —       | Filter rentang: `A` (>=3.5), `B` (3.0–3.49), `C` (2.5–2.99), `D` (<2.5) |

**Respons 200:**
```json
{
  "success": true,
  "message": "daftar mahasiswa",
  "data": [
    {
      "id": 1,
      "nim": "434241022",
      "name": "Putri Apriliana",
      "grade": 3.99,
      "is_active": true,
      "created_at": "2026-09-05T17:41:06.375046+07:00"
    }
  ],
  "meta": { "page": 1, "limit": 10, "total": 1, "total_pages": 1 }
}
```

### 2. GET /api/v1/students/:id

| Auth | Status Sukses | Status Gagal |
|---|---|---|
| Wajib token | 200 | 400, 401, 404 |

**Respons 404:**
```json
{ "success": false, "message": "mahasiswa tidak ditemukan" }
```

### 3. POST /api/v1/students

| Auth | Status Sukses | Status Gagal |
|---|---|---|
| Wajib token | 201 + header `Location` | 400, 401, 409, 415, 422 |

**Body:**
```json
{ "nim": "434241025", "name": "Dina", "grade": 3.70, "is_active": true }
```

**Respons 409 (NIM duplikat — dilempar oleh `UNIQUE INDEX` di database):**
```json
{ "success": false, "message": "nim sudah digunakan oleh mahasiswa lain" }
```

**Respons 422 (validasi gagal):**
```json
{
  "success": false,
  "message": "validasi gagal",
  "errors": { "nim": "nim tidak boleh kosong" }
}
```

### 4. PUT /api/v1/students/:id

Ganti seluruh data mahasiswa. Semua field wajib dikirim.

| Auth | Status Sukses | Status Gagal |
|---|---|---|
| Wajib token | 200 | 400, 401, 404, 409, 415, 422 |

**Respons 200:**
```json
{
  "success": true,
  "message": "data mahasiswa berhasil diperbarui (PUT)",
  "data": { "id": 1, "nim": "434241022", "name": "Putri Apriliana", "grade": 3.95, "is_active": true, "created_at": "..." }
}
```

### 5. PATCH /api/v1/students/:id

Ubah sebagian data mahasiswa. Cuma field yang dikirim yang berubah; alur baca-dulu-lalu-simpan (`FindByID` → ubah field yang dikirim → `Update`).

| Auth | Status Sukses | Status Gagal |
|---|---|---|
| Wajib token | 200 | 400, 401, 404, 409, 415, 422 |

**Body (misal cuma ubah grade):**
```json
{ "grade": 3.99 }
```

**Respons 200:**
```json
{
  "success": true,
  "message": "data mahasiswa berhasil diperbarui (PATCH)",
  "data": { "id": 1, "nim": "434241022", "name": "Putri Apriliana", "grade": 3.99, "is_active": true, "created_at": "..." }
}
```

**Bedanya PUT sama PATCH:** PUT kirim semua field (data lama ditimpa total), PATCH kirim field tertentu saja (yang lain tidak berubah).

### 6. DELETE /api/v1/students/:id

| Auth | Status Sukses | Status Gagal |
|---|---|---|
| Wajib token | 204 (tanpa body) | 400, 401, 404 |

## Status Code

| Status | Kapan muncul                                              |
|--------|--------------------------------------------------------------|
| 200    | Data berhasil diambil atau diubah                             |
| 201    | Data baru berhasil dibuat + header `Location`                 |
| 204    | Data berhasil dihapus, tidak ada body                         |
| 400    | JSON rusak atau ID bukan angka                                |
| 401    | Belum login (token tidak ada, tidak valid, atau kedaluwarsa) — Modul 5 |
| 404    | Data yang dicari tidak ada                                    |
| 409    | NIM/username/email bentrok dengan data yang sudah ada          |
| 415    | Content-Type bukan `application/json`                          |
| 422    | Ada field yang tidak lolos validasi                            |
| 429    | Terlalu banyak percobaan login dalam satu menit — Modul 5       |
| 500    | Error tak terduga (termasuk database mati saat endpoint dipanggil) |
| 503    | `/health` gagal terhubung ke database                          |

## Keamanan (Modul 5)

| Ancaman | Mitigasi |
|---|---|
| Password bocor | Hash bcrypt cost 12, tidak pernah tersimpan/keluar sebagai plaintext |
| User enumeration | Pesan error login identik untuk username tidak ada maupun password salah, plus dummy hash check |
| Algorithm confusion (`alg:none`) | Signing method diverifikasi eksplisit saat parsing JWT |
| Mass assignment | Struct request register tidak memiliki field `role` |
| Refresh token dicuri | Disimpan sebagai hash SHA-256, wajib dirotasi setiap pemakaian |
| Brute force login | Rate limit 5 percobaan/menit/IP, respons 429 + `Retry-After` |
| Payload raksasa | `BodyLimit` 1 MB pada konfigurasi Fiber |
| CORS longgar | Origin dibatasi lewat `ALLOWED_ORIGINS` |

## Tech Stack

- Go 1.26
- Fiber v2.52
- PostgreSQL 18 (driver `pgx/v5`, connection pool `pgxpool`)
- `log/slog` + `lumberjack` untuk logger JSON terstruktur dengan rotasi file
- Repository pattern: interface `StudentRepository`/`UserRepository`/`TokenRepository` memisahkan kontrak dari implementasi
- `golang-jwt/jwt/v5` untuk pembuatan dan verifikasi JSON Web Token
- `golang.org/x/crypto/bcrypt` untuk hashing password

## Menjalankan Unit Test

```bash
go test ./app/service/ -v
```

## Sumber Bantuan

- Modul Praktikum Pertemuan 2, 3, 4, dan 5 (Universitas Airlangga).
- Claude (Anthropic) .