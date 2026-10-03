# API Students — REST API & HTTP Deep Dive

Proyek REST API untuk manajemen data mahasiswa menggunakan framework **Go Fiber v2**, dibuat untuk memenuhi penugasan **Modul Praktikum Pertemuan 2 — REST API & HTTP Deep Dive** pada mata kuliah Praktikum Pemrograman Backend Lanjut, Program Studi D4 Teknik Informatika, Fakultas Vokasi, Universitas Airlangga.

---

## 📋 Ringkasan Fitur
- **Entitas Mahasiswa (`Student`)**: Membawa field dari Modul 1 (`ID`, `Name`, `Grade`, `IsActive`) ditambah field unik `NIM` dan timestamp `CreatedAt`.
- **Method Receiver Reusability**: Menggunakan pointer receiver `UpdateGrade`, `Activate`, dan `Deactivate` pada proses update parsial (PATCH).
- **Semantik HTTP Yang Ketat**:
  - `GET`: Safe & Idempotent (daftar dengan pagination, search, sort, filter, dan single resource).
  - `POST`: Non-Idempotent (menghasilkan sumber daya baru dengan status `201 Created` dan header `Location`).
  - `PUT`: Idempotent (mengganti seluruh isi sumber daya; seluruh field wajib dikirim).
  - `PATCH`: Non-Idempotent / partial update (hanya field yang dikirim yang diubah menggunakan pointer struct).
  - `DELETE`: Idempotent (menghapus sumber daya dan mengembalikan status `204 No Content` tanpa body).
- **Penanganan Status HTTP Komprehensif**:
  - `200 OK`: Sukses GET, PUT, PATCH.
  - `201 Created`: Sukses POST dengan header `Location`.
  - `204 No Content`: Sukses DELETE tanpa body.
  - `400 Bad Request`: Payload JSON tidak valid, ID non-numerik, atau request body kosong pada PATCH.
  - `404 Not Found`: Mahasiswa tidak ditemukan atau route tidak terdaftar.
  - `409 Conflict`: Bentrokan NIM ganda pada POST/PUT/PATCH.
  - `415 Unsupported Media Type`: Header `Content-Type` bukan `application/json` pada request ber-body.
  - `422 Unprocessable Entity`: Validasi nilai masukan gagal (misal nilai Grade di luar 0–100, nama kosong).
- **Pagination, Sorting, & Filtering Aman**:
  - Batas atas limit aman (`limit <= 50`) guna mencegah serangan DoS dan OOM.
  - Whitelist sorting (`id`, `nim`, `name`, `grade`, `created_at`).
  - Filter `is_active` dan rentang nilai `min_grade` serta `max_grade`.

---

## 🚀 Cara Menjalankan Server

1. **Masuk ke direktori proyek:**
   ```bash
   cd tugas-code/latihan-fiber/api-students
   ```

2. **Jalankan aplikasi:**
   ```bash
   go run .
   ```
   *Secara default server akan berjalan pada port `3001` (`http://localhost:3001`).*

---

## 📑 Dokumen Kontrak API

Berikut adalah tabel spesifikasi kontrak antarmuka API (API Contract):

| Metode | Endpoint | Parameter / Query String | Contoh Body Permintaan | Status yang Mungkin Dikembalikan | Contoh Body Respons |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **GET** | `/api/v1/health` | *-* | *-* | `200 OK` | `{"success":true,"message":"server berjalan normal","data":{"service":"api-students","status":"healthy","timestamp":"2026-10-03T..."}}` |
| **GET** | `/api/v1/students` | **Query Parameter (Opsional):**<br>• `page` (int, default: 1)<br>• `limit` (int, default: 10, max: 50)<br>• `search` (string, cari nama/NIM)<br>• `sort` (whitelist: id, nim, name, grade, created_at)<br>• `order` (asc / desc, default: asc)<br>• `is_active` (bool: true/false)<br>• `min_grade` (float)<br>• `max_grade` (float) | *-* | `200 OK` | `{"success":true,"message":"daftar mahasiswa berhasil diambil","data":[{"id":1,"nim":"2308101001","name":"Putri Apriliana","grade":92.5,"is_active":true,"created_at":"..."}],"meta":{"page":1,"limit":10,"total":1,"total_pages":1}}` |
| **GET** | `/api/v1/students/:id` | **Path Parameter:**<br>• `id` (integer positif, wajib) | *-* | • `200 OK`<br>• `400 Bad Request` (id non-angka)<br>• `404 Not Found` (id tidak ada) | `{"success":true,"message":"mahasiswa ditemukan","data":{"id":1,"nim":"2308101001","name":"Putri Apriliana","grade":92.5,"is_active":true,"created_at":"..."}}` |
| **POST** | `/api/v1/students` | **Header Wajib:**<br>`Content-Type: application/json` | `{"nim":"2308101004","name":"Dewi Sartika","grade":88.0,"is_active":true}` | • `201 Created` (+ Header `Location: /api/v1/students/4`)<br>• `400 Bad Request` (JSON rusak)<br>• `409 Conflict` (NIM sudah ada)<br>• `415 Unsupported Media Type` (bukan JSON)<br>• `422 Unprocessable Entity` (validasi gagal) | `{"success":true,"message":"mahasiswa berhasil dibuat","data":{"id":4,"nim":"2308101004","name":"Dewi Sartika","grade":88,"is_active":true,"created_at":"..."}}` |
| **PUT** | `/api/v1/students/:id` | **Path Parameter:**<br>• `id` (integer positif, wajib)<br>**Header Wajib:**<br>`Content-Type: application/json` | `{"nim":"2308101001","name":"Putri Apriliana S.Tr.Kom","grade":98.0,"is_active":true}` *(seluruh field wajib)* | • `200 OK`<br>• `400 Bad Request` (id / JSON salah)<br>• `404 Not Found` (id tidak ada)<br>• `409 Conflict` (NIM konflik)<br>• `415 Unsupported Media Type`<br>• `422 Unprocessable Entity` (field kosong / invalid) | `{"success":true,"message":"mahasiswa berhasil diganti seluruhnya (PUT)","data":{"id":1,"nim":"2308101001","name":"Putri Apriliana S.Tr.Kom","grade":98,"is_active":true,"created_at":"..."}}` |
| **PATCH** | `/api/v1/students/:id` | **Path Parameter:**<br>• `id` (integer positif, wajib)<br>**Header Wajib:**<br>`Content-Type: application/json` | `{"grade":95.5}` *(hanya field yang ingin diubah)* | • `200 OK`<br>• `400 Bad Request` (id/JSON salah atau body kosong)<br>• `404 Not Found` (id tidak ada)<br>• `409 Conflict` (NIM konflik)<br>• `415 Unsupported Media Type`<br>• `422 Unprocessable Entity` (field invalid) | `{"success":true,"message":"mahasiswa berhasil diperbarui sebagian (PATCH)","data":{"id":1,"nim":"2308101001","name":"Putri Apriliana","grade":95.5,"is_active":true,"created_at":"..."}}` |
| **DELETE** | `/api/v1/students/:id` | **Path Parameter:**<br>• `id` (integer positif, wajib) | *-* | • `204 No Content` (tanpa body)<br>• `400 Bad Request` (id non-angka)<br>• `404 Not Found` (id tidak ada) | *(Tidak ada response body / Kosong)* |

---

## 🛡️ Kebijakan Keamanan dan Validasi Query
1. **Batas Atas Limit (Max 50)**:
   Mencegah klien mengirim nilai limit yang luar biasa besar (misal `?limit=1000000`) yang dapat menyebabkan alokasi memori berlebihan (*Out-Of-Memory*) dan *Denial of Service* (DoS).
2. **Whitelist Sort Field**:
   Hanya field yang terdaftar dalam `allowedSort` (`id`, `nim`, `name`, `grade`, `created_at`) yang diperbolehkan untuk pengurutan. Permintaan dengan parameter sort acak akan diabaikan dan kembali ke default (`id`), mencegah *query injection* saat nantinya dihubungkan ke basis data.
3. **Penyaringan Case-Insensitive**:
   Pencarian teks pada nama dan NIM diubah menjadi huruf kecil (*lowercase*) sebelum pencocokan sehingga hasil pencarian bersifat ramah pengguna dan konsisten.
