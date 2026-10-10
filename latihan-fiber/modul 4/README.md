# API Students — Clean Architecture (Modul 4)

Proyek REST API untuk manajemen data mahasiswa menggunakan framework **Go Fiber v2**, driver PostgreSQL **pgx v5** (`pgxpool`), dan arsitektur **Clean Architecture (Struktur Berlapis Baku)**. Proyek ini dibuat untuk memenuhi penugasan **Modul Praktikum Pertemuan 4: Clean Architecture** pada mata kuliah Praktikum Pemrograman Backend Lanjut, Program Studi D4 Teknik Informatika, Fakultas Vokasi, Universitas Airlangga.

---

## 🏛️ Arsitektur & Struktur Berlapis (Clean Architecture)

Proyek ini menerapkan prinsip **Clean Architecture (Uncle Bob)** dengan pemetaan lingkaran sepusat:
```text
api-students/
├── app/
│   ├── model/                  <- Layer 1: Entities (Core Business Rules)
│   │   └── student.go          <- Struct Student, DTO Request-Response, Tag JSON
│   ├── repository/             <- Layer 3: Interface Adapters (Gateways)
│   │   └── student_repository.go <- Implementasi query PostgreSQL pgxpool, Sentinel Error
│   └── service/                <- Layer 2: Use Cases & Layer 3: Controller
│       ├── student_rules.go    <- Business rules murni (Validasi, Patching, Pagination Math)
│       ├── student_rules_test.go <- Unit testing business rules tanpa DB & Fiber
│       └── student_service.go  <- Controller penerima *fiber.Ctx & error mapper
├── config/                     <- Layer 4: Frameworks & Drivers
│   ├── app.go                  <- Instance Fiber, wiring middleware, routes, centralized error handler
│   ├── env.go                  <- Pemuat variabel environment (.env)
│   └── logger.go               <- Structured JSON logger (slog) & file rotation (lumberjack)
├── database/                   <- Layer 4: Frameworks & Drivers
│   └── postgres.go             <- Inisialisasi pgxpool Connection Pool dengan startup Ping
├── helper/                     <- Layer 3: Interface Adapters (Presenter & Request Parser)
│   ├── request.go              <- Parser query string, ParamID, context timeout (5s)
│   └── response.go             <- Presenter amplop WebResponse (Success, Created, Fail, dll)
├── logs/                       <- Output file log (logs/app.log, dirotasi otomatis, tidak di-commit)
├── middleware/                 <- Layer 4: Frameworks & Drivers
│   └── middleware.go           <- RequestID, Recover, Helmet, CORS, RequestLogger, RequireJSON (415)
├── route/                      <- Layer 4: Frameworks & Drivers
│   └── route.go                <- Pendaftaran rute murni (pemetaan URL ke service)
├── migrations/
│   └── 001_create_students.sql <- Skema tabel students dan indeks unik NIM
├── .env                        <- Konfigurasi lokal (tidak di-commit)
├── .env.example                <- Template konfigurasi untuk rekan kolaborator
├── main.go                     <- Titik masuk aplikasi: perakitan dependensi & graceful shutdown
├── go.mod
└── go.sum
```

---

## 🔒 Aturan Dependency (Dependency Rule)

Dependency kode hanya boleh mengarah **ke dalam**:
```text
route ---> service ---> repository ---> model
  |           |              |
  v           v              v
middleware  helper        (independen)
```
- `app/model`: Tidak mengimpor package apa pun dari proyek sendiri.
- `app/repository`: Tidak mengimpor Fiber (`*fiber.Ctx`) sama sekali.
- `app/service`: Tidak menulis perintah SQL langsung; mendelegasikan akses data ke interface `StudentRepository`.
- `route`: Murni pemetaan URL ke method service tanpa logika `if` dan validasi.
- `main.go`: Murni perakitan dependensi dan penanganan sinyal *graceful shutdown*.

---

## 🚀 Panduan Menyiapkan Basis Data dari Nol

1. **Buat basis data di PostgreSQL:**
   ```bash
   psql -U postgres -c "CREATE DATABASE praktikum_backend;"
   ```

2. **Jalankan migrasi skema tabel:**
   ```bash
   psql -U postgres -d praktikum_backend -f migrations/001_create_students.sql
   ```

3. **Salin dan sesuaikan file environment:**
   ```bash
   cp .env.example .env
   ```
   Isi konfigurasi pada `.env`:
   ```env
   APP_PORT=3001
   APP_NAME="Praktikum Backend Lanjut - Pertemuan 4"
   LOG_LEVEL=info
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=kata_sandi_anda
   DB_NAME=praktikum_backend
   DB_SSLMODE=disable
   DB_MAX_CONNS=10
   ```

4. **Jalankan Unit Test:**
   ```bash
   go test ./app/service/... -v
   ```

5. **Jalankan Server:**
   ```bash
   go run .
   ```
   Server akan aktif di `http://localhost:3001`.

---

## 📑 Dokumen Kontrak API

| Metode | Endpoint | Parameter / Query String | Contoh Body Permintaan | Status HTTP | Contoh Body Respons |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **GET** | `/api/v1/health` | *-* | *-* | • `200 OK`<br>• `503 Service Unavailable` | `{"success":true,"message":"server dan basis data berjalan normal","data":{"database":"connected","service":"api-students","status":"healthy"}}` |
| **GET** | `/api/v1/students` | **Query Parameter:**<br>• `page`, `limit` (max: 50)<br>• `search` (ILIKE)<br>• `sort` (whitelist), `order`<br>• `is_active`, `min_grade`, `max_grade` | *-* | • `200 OK`<br>• `500 Internal Server Error` | `{"success":true,"message":"daftar mahasiswa berhasil diambil","data":[{"id":1,"nim":"434241022","name":"Putri Apriliana","grade":95,"is_active":true,"created_at":"..."}],"meta":{"page":1,"limit":10,"total":1,"total_pages":1}}` |
| **GET** | `/api/v1/students/:id` | **Path Parameter:**<br>• `id` (int positif, wajib) | *-* | • `200 OK`<br>• `400 Bad Request`<br>• `404 Not Found`<br>• `500 Internal Server Error` | `{"success":true,"message":"mahasiswa ditemukan","data":{"id":1,"nim":"434241022","name":"Putri Apriliana","grade":95,"is_active":true,"created_at":"..."}}` |
| **POST** | `/api/v1/students` | **Header Wajib:**<br>`Content-Type: application/json` | `{"nim":"434241023","name":"Budi Santoso","grade":85.0,"is_active":true}` | • `201 Created` (+ Header `Location`)<br>• `400 Bad Request`<br>• `409 Conflict` (NIM ganda)<br>• `415 Unsupported Media Type`<br>• `422 Unprocessable Entity`<br>• `500 Internal Server Error` | `{"success":true,"message":"mahasiswa berhasil dibuat","data":{"id":2,"nim":"434241023","name":"Budi Santoso","grade":85,"is_active":true,"created_at":"..."}}` |
| **PUT** | `/api/v1/students/:id` | **Path Parameter:**<br>• `id` (int positif, wajib)<br>**Header Wajib:**<br>`Content-Type: application/json` | `{"nim":"434241022","name":"Putri Apriliana S.Tr.Kom","grade":98.0,"is_active":false}` | • `200 OK`<br>• `400 Bad Request`<br>• `404 Not Found`<br>• `409 Conflict`<br>• `415 Unsupported Media Type`<br>• `422 Unprocessable Entity`<br>• `500 Internal Server Error` | `{"success":true,"message":"mahasiswa berhasil diganti seluruhnya (PUT)","data":{"id":1,"nim":"434241022","name":"Putri Apriliana S.Tr.Kom","grade":98,"is_active":false,"created_at":"..."}}` |
| **PATCH** | `/api/v1/students/:id` | **Path Parameter:**<br>• `id` (int positif, wajib)<br>**Header Wajib:**<br>`Content-Type: application/json` | `{"grade":99.5,"is_active":true}` | • `200 OK`<br>• `400 Bad Request`<br>• `404 Not Found`<br>• `409 Conflict`<br>• `415 Unsupported Media Type`<br>• `422 Unprocessable Entity`<br>• `500 Internal Server Error` | `{"success":true,"message":"mahasiswa berhasil diperbarui sebagian (PATCH)","data":{"id":1,"nim":"434241022","name":"Putri Apriliana S.Tr.Kom","grade":99.5,"is_active":true,"created_at":"..."}}` |
| **DELETE** | `/api/v1/students/:id` | **Path Parameter:**<br>• `id` (int positif, wajib) | *-* | • `204 No Content` (tanpa body)<br>• `400 Bad Request`<br>• `404 Not Found`<br>• `500 Internal Server Error` | *(Tanpa Body / Kosong)* |

---

## 📝 Logging Terstruktur & Rotasi Log
Aplikasi menulis log HTTP ke dua target sekaligus:
1. **Layar (stdout):** Format JSON terbaca langsung saat pengembangan.
2. **File `logs/app.log`:** Dirotasi otomatis oleh `lumberjack` (maksimal 10 MB per file, retensi 5 file cadangan, maksimal 14 hari).

Contoh baris log:
```json
{"time":"2026-10-10T14:15:00.123+07:00","level":"INFO","msg":"http_request","request_id":"54a2e8c2-3e4b-4f93-b672-1c25189e4719","method":"GET","path":"/api/v1/students","status":200,"duration":1850400,"ip":"127.0.0.1"}
```
