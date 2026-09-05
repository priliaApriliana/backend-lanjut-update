# API Students — Modul 3 (Database & Repository Pattern)

REST API untuk mengelola data mahasiswa, dibangun menggunakan Go dan Fiber v2, data disimpan permanen di PostgreSQL melalui pola repository.

- **Nama:** Putri Apriliana
- **NIM:** 434241022
- **Repositori:** https://github.com/priliaApriliana/backend-lanjut
- **Lokasi proyek:** `tugas-mandiri/api-students`

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

4. Jalankan migrasi:
```bash
   psql -U postgres -d tugas_mandiri_students -f migrations/001_create_students.sql
```

5. Salin `.env.example` menjadi `.env`, lalu isi nilainya sesuai konfigurasi PostgreSQL lokal:
```bash
   cp .env.example .env
```

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
```

**Kenapa NIM unik dijaga di database, bukan di kode Go:**
Pemeriksaan manual (loop mencari NIM yang sama di kode Go, seperti versi Modul 2) punya celah race condition — dua request POST dengan NIM yang sama, yang datang nyaris bersamaan, bisa lolos pemeriksaan sebelum salah satunya sempat tersimpan. `UNIQUE INDEX` di database tidak punya celah itu, karena database sendiri yang memutuskan mana yang lebih dulu masuk.

**Kenapa ada index tambahan pada `LOWER(name)`:**
Endpoint List melakukan pencarian nama secara case-insensitive lewat `ILIKE`. Tanpa index ini, setiap pencarian nama melakukan full table scan pada seluruh baris tabel.

## Struktur Proyek
```
api-students/
├── app/
│   ├── model/
│   │   └── student.go         struct entitas, request, dan ListQuery
│   └── repository/
│       └── student_repository.go   kontrak dan implementasi akses data (pgx)
├── config/
│   └── env.go                 memuat variabel environment
├── database/
│   └── postgres.go            koneksi dan connection pool
├── migrations/
│   └── 001_create_students.sql
├── .env                        (tidak di-commit)
├── .env.example
├── main.go
├── handler.go
└── helper.go
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
|-----------|---------|----------------------------------------------|
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

## Endpoint

### 1. GET /api/v1/students

| Item             | Detail                              |
|------------------|--------------------------------------|
| **Metode**       | GET                                  |
| **Endpoint**     | `/api/v1/students`                   |
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

`sort` memakai daftar putih nama kolom karena `ORDER BY` tidak bisa diparameterkan secara aman di SQL. Kalau nilai `sort` di luar daftar putih (termasuk percobaan SQL injection), otomatis balik ke `id` dan query tetap berjalan aman.

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

| Status Sukses | Status Gagal |
|---|---|
| 200 | 400, 404 |

**Respons 404:**
```json
{ "success": false, "message": "mahasiswa tidak ditemukan" }
```

### 3. POST /api/v1/students

| Status Sukses | Status Gagal |
|---|---|
| 201 + header `Location` | 400, 409, 415, 422 |

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

| Status Sukses | Status Gagal |
|---|---|
| 200 | 400, 404, 409, 415, 422 |

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

| Status Sukses | Status Gagal |
|---|---|
| 200 | 400, 404, 409, 415, 422 |

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

| Status Sukses | Status Gagal |
|---|---|
| 204 (tanpa body) | 400, 404 |

## Status Code

| Status | Kapan muncul                                              |
|--------|--------------------------------------------------------------|
| 200    | Data berhasil diambil atau diubah                             |
| 201    | Data baru berhasil dibuat + header `Location`                 |
| 204    | Data berhasil dihapus, tidak ada body                         |
| 400    | JSON rusak atau ID bukan angka                                |
| 404    | Data yang dicari tidak ada                                    |
| 409    | NIM bentrok dengan data yang sudah ada (`UNIQUE INDEX`)        |
| 415    | Content-Type bukan `application/json`                          |
| 422    | Ada field yang tidak lolos validasi                            |
| 500    | Error tak terduga (termasuk database mati saat endpoint students dipanggil) |
| 503    | `/health` gagal terhubung ke database                          |

## Tech Stack

- Go 1.26
- Fiber v2.52
- PostgreSQL 18 (driver `pgx/v5`, connection pool `pgxpool`)
- Repository pattern: interface `StudentRepository` di `app/repository`, implementasi PostgreSQL terpisah dari handler

## Sumber Bantuan

Modul Praktikum Pertemuan 2 dan 3 (Universitas Airlangga).
Claude (Anthropic) digunakan sebagai asisten untuk menyusun struktur kode repository pattern, penulisan query berparameter, debugging error compile/environment, dokumentasi mengikuti struktur dan gaya kode yang sudah dibuat sendiri pada tugas Modul 2.