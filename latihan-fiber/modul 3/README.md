# API Students — Database & Repository Pattern (Modul 3)

Proyek REST API untuk manajemen data mahasiswa menggunakan framework **Go Fiber v2**, driver PostgreSQL **pgx v5** (`pgxpool`), dan arsitektur **Repository Pattern**. Proyek ini dibuat untuk memenuhi penugasan **Modul Praktikum Pertemuan 3: Database & Repository Pattern** pada mata kuliah Praktikum Pemrograman Backend Lanjut, Program Studi D4 Teknik Informatika, Fakultas Vokasi, Universitas Airlangga.

---

## 🏗️ Arsitektur & Pola Desain (Repository Pattern)

Aplikasi ini memisahkan tanggung jawab kode ke dalam lapisan hierarki searah:
```text
Handler (HTTP & Validation)  --->  Repository (Data Access Interface)  --->  PostgreSQL Database
             |                                    |
             v                                    v
     model.WebResponse                       model.Student
```

1. **Abstraksi Interface (`StudentRepository`)**:
   Handler hanya mengenal interface, bukan struct konkret atau detail implementasi basis data.
2. **Ketergantungan Searah**:
   Paket `repository` murni mengelola operasi basis data dan **sama sekali tidak mengimpor atau mengenal Fiber (`fiber.Ctx`)**.
3. **Pencegahan SQL Injection**:
   Seluruh nilai input dari klien dipassing sebagai parameter query terpisah (`$1, $2, ...`). Bagian query yang tidak dapat diparameterkan (`ORDER BY`) diamankan menggunakan daftar putih (*whitelist*).
4. **Penyaringan & Paginasi di Sisi Basis Data**:
   Pencarian teks menggunakan operator `ILIKE`, penyaringan status menggunakan `WHERE`, pemenggalan menggunakan `LIMIT` & `OFFSET`, serta total dihitung secara efisien lewat `SELECT COUNT(*)`.

---

## 🗄️ Skema Tabel & Migrasi Basis Data

Skema tabel didefinisikan pada file `migrations/001_create_students.sql`:

```sql
CREATE TABLE IF NOT EXISTS students (
    id SERIAL PRIMARY KEY,
    nim VARCHAR(20) NOT NULL,
    name VARCHAR(100) NOT NULL,
    grade NUMERIC(5,2) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indeks unik untuk NIM tanpa membedakan huruf besar dan kecil.
-- Menjamin keunikan data di tingkat basis data (ACID) dan mencegah race condition.
CREATE UNIQUE INDEX IF NOT EXISTS students_nim_lower_key
    ON students (LOWER(nim));

-- Indeks tambahan pada nama mahasiswa untuk mempercepat klausa pencarian ILIKE dan ORDER BY
CREATE INDEX IF NOT EXISTS students_name_lower_idx
    ON students (LOWER(name));
```

---

## 🚀 Panduan Menyiapkan Basis Data dari Nol

Bagi rekan mahasiswa yang baru mengklona repositori ini, ikuti langkah-langkah penyiapan berikut:

### 1. Buat Basis Data PostgreSQL
Buka terminal dan buat basis data baru menggunakan `psql`:
```bash
psql -U postgres -c "CREATE DATABASE praktikum_backend;"
```

### 2. Jalankan Migrasi Skema Tabel
Jalankan file SQL migrasi ke dalam basis data yang baru dibuat:
```bash
psql -U postgres -d praktikum_backend -f migrations/001_create_students.sql
```

Verifikasi tabel yang terbentuk:
```bash
psql -U postgres -d praktikum_backend -c "\d students"
```

### 3. Konfigurasi File Environment (`.env`)
Salin file template `.env.example` menjadi `.env`:
```bash
cp .env.example .env
```

Buka file `.env` dan sesuaikan nilainya dengan konfigurasi PostgreSQL lokal Anda:
```env
APP_PORT=3001
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=kata_sandi_postgres_anda
DB_NAME=praktikum_backend
DB_SSLMODE=disable
DB_MAX_CONNS=10
```

### 4. Jalankan Aplikasi
Unduh dependensi dan jalankan server Go:
```bash
go mod tidy
go run .
```
Server akan aktif pada: `http://localhost:3001`.

---

## 📑 Dokumen Kontrak API

Berikut adalah tabel spesifikasi kontrak antarmuka API mahasiswa (`api-students`):

| Metode | Endpoint | Parameter / Query String | Contoh Body Permintaan | Status yang Mungkin Dikembalikan | Contoh Body Respons |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **GET** | `/api/v1/health` | *-* | *-* | • `200 OK` (Server & DB sehat)<br>• `503 Service Unavailable` (DB mati) | `{"success":true,"message":"server dan basis data berjalan normal","data":{"service":"api-students","status":"healthy","database":"connected","timestamp":"2026-10-03T..."}}` |
| **GET** | `/api/v1/students` | **Query Parameter:**<br>• `page` (int, default: 1)<br>• `limit` (int, default: 10, max: 50)<br>• `search` (string, cari nama/NIM via ILIKE)<br>• `sort` (whitelist: id, nim, name, grade, created_at)<br>• `order` (asc / desc, default: asc)<br>• `is_active` (bool: true/false)<br>• `min_grade` (float)<br>• `max_grade` (float) | *-* | • `200 OK`<br>• `500 Internal Server Error` | `{"success":true,"message":"daftar mahasiswa berhasil diambil","data":[{"id":1,"nim":"434241022","name":"Putri Apriliana","grade":95,"is_active":true,"created_at":"..."}],"meta":{"page":1,"limit":10,"total":1,"total_pages":1}}` |
| **GET** | `/api/v1/students/:id` | **Path Parameter:**<br>• `id` (int positif, wajib) | *-* | • `200 OK`<br>• `400 Bad Request` (id bukan angka)<br>• `404 Not Found` (id tidak ada di DB)<br>• `500 Internal Server Error` | `{"success":true,"message":"mahasiswa ditemukan","data":{"id":1,"nim":"434241022","name":"Putri Apriliana","grade":95,"is_active":true,"created_at":"..."}}` |
| **POST** | `/api/v1/students` | **Header Wajib:**<br>`Content-Type: application/json` | `{"nim":"434241025","name":"Dina","grade":88.0,"is_active":true}` | • `201 Created` (+ Header `Location`)<br>• `400 Bad Request` (JSON rusak)<br>• `409 Conflict` (NIM bentrok / kode 23505)<br>• `415 Unsupported Media Type`<br>• `422 Unprocessable Entity`<br>• `500 Internal Server Error` | `{"success":true,"message":"mahasiswa berhasil dibuat","data":{"id":2,"nim":"434241025","name":"Dina","grade":88,"is_active":true,"created_at":"..."}}` |
| **PUT** | `/api/v1/students/:id` | **Path Parameter:**<br>• `id` (int positif, wajib)<br>**Header Wajib:**<br>`Content-Type: application/json` | `{"nim":"434241022","name":"Putri Apriliana S.Tr.Kom","grade":98.0,"is_active":false}` | • `200 OK`<br>• `400 Bad Request`<br>• `404 Not Found`<br>• `409 Conflict`<br>• `415 Unsupported Media Type`<br>• `422 Unprocessable Entity`<br>• `500 Internal Server Error` | `{"success":true,"message":"mahasiswa berhasil diganti seluruhnya (PUT)","data":{"id":1,"nim":"434241022","name":"Putri Apriliana S.Tr.Kom","grade":98,"is_active":false,"created_at":"..."}}` |
| **PATCH** | `/api/v1/students/:id` | **Path Parameter:**<br>• `id` (int positif, wajib)<br>**Header Wajib:**<br>`Content-Type: application/json` | `{"grade":99.0,"is_active":true}` | • `200 OK`<br>• `400 Bad Request`<br>• `404 Not Found`<br>• `409 Conflict`<br>• `415 Unsupported Media Type`<br>• `422 Unprocessable Entity`<br>• `500 Internal Server Error` | `{"success":true,"message":"mahasiswa berhasil diperbarui sebagian (PATCH)","data":{"id":1,"nim":"434241022","name":"Putri Apriliana S.Tr.Kom","grade":99,"is_active":true,"created_at":"..."}}` |
| **DELETE** | `/api/v1/students/:id` | **Path Parameter:**<br>• `id` (int positif, wajib) | *-* | • `204 No Content` (tanpa body)<br>• `400 Bad Request`<br>• `404 Not Found`<br>• `500 Internal Server Error` | *(Tanpa Body / Kosong)* |

---

## 🛡️ Pemetaan Kode Status & Error Basis Data

| Yang Terjadi di Basis Data | Error Repository | Status HTTP yang Diterjemahkan |
| :--- | :--- | :--- |
| Query tidak mengembalikan baris (`pgx.ErrNoRows`) | `ErrNotFound` | **404 Not Found** |
| Pelanggaran batasan `UNIQUE INDEX` (Kode PostgreSQL `23505`) | `ErrDuplicate` | **409 Conflict** |
| Perintah `DELETE` tidak mempengaruhi baris mana pun (`RowsAffected == 0`) | `ErrNotFound` | **404 Not Found** |
| Layanan PostgreSQL mati / koneksi putus saat endpoint dipanggil | Error diteruskan | **500 Internal Server Error** (atau **503 Service Unavailable** pada `/health`) |
