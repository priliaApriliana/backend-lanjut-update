# API Students — Modul 2

REST API sederhana untuk mengelola data mahasiswa, dibangun menggunakan Go dan Fiber v2.

- **Nama:** Putri Apriliana
- **NIM:** 434241022
- **Repositori:** https://github.com/priliaApriliana/backend-lanjut

## Cara Jalankan

```bash
cd api-students
go run .
```

Akses di `http://localhost:3000`

---

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

---

## Entitas Student

| Field        | Tipe      | Keterangan                   |
|--------------|-----------|------------------------------|
| `id`         | int       | ID unik, auto-increment      |
| `nim`        | string    | NIM, harus unik              |
| `name`       | string    | Nama mahasiswa                |
| `grade`      | float64   | IPK, rentang 0–4.0            |
| `is_active`  | bool      | Status aktif/nonaktif         |
| `created_at` | datetime  | Waktu data dibuat             |

---

## Endpoint

### 1. GET /api/v1/students

Ambil daftar mahasiswa. Bisa pakai paginasi, search, sort, dan filter.

| Item             | Detail                              |
|------------------|-------------------------------------|
| **Metode**       | GET                                 |
| **Endpoint**     | `/api/v1/students`                  |
| **Body**         | —                                   |
| **Status Sukses**| 200                                 |

**Query Parameter:**

| Parameter   | Default | Keterangan                                        |
|-------------|---------|---------------------------------------------------|
| `page`      | 1       | Halaman ke berapa                                 |
| `limit`     | 10      | Jumlah per halaman (maks 100)                     |
| `search`    | —       | Cari nama, tidak case-sensitive                   |
| `sort`      | id      | Urutkan: `id`, `nim`, `name`, `grade`             |
| `order`     | asc     | `asc` atau `desc`                                 |
| `is_active` | —       | Filter: `true` / `false`                          |
| `grade`     | —       | Filter rentang: `A`, `B`, `C`, `D`                |

`limit` dibatasi maks 100 supaya client tidak tarik semua data sekaligus. `sort` pakai daftar putih — kalau isinya di luar daftar, otomatis balik ke `id`.

**Contoh:**

```
GET /api/v1/students?page=1&limit=2&search=put&sort=grade&order=desc
```

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
      "created_at": "2026-09-04T12:16:50.46+07:00"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 2,
    "total": 1,
    "total_pages": 1
  }
}
```

---

### 2. GET /api/v1/students/:id

Ambil satu mahasiswa dari ID-nya.

| Item             | Detail                              |
|------------------|-------------------------------------|
| **Metode**       | GET                                 |
| **Endpoint**     | `/api/v1/students/:id`              |
| **Body**         | —                                   |
| **Status Sukses**| 200                                 |
| **Status Gagal** | 400, 404                            |

**Respons 200:**

```json
{
  "success": true,
  "message": "data mahasiswa ditemukan",
  "data": {
    "id": 1,
    "nim": "434241022",
    "name": "Putri Apriliana",
    "grade": 3.99,
    "is_active": true,
    "created_at": "2026-09-04T12:16:50.46+07:00"
  }
}
```

**Respons 400 (id bukan angka):**

```json
{
  "success": false,
  "message": "id harus berupa angka"
}
```

**Respons 404:**

```json
{
  "success": false,
  "message": "mahasiswa tidak ditemukan"
}
```

---

### 3. POST /api/v1/students

Tambah mahasiswa baru.

| Item             | Detail                              |
|------------------|-------------------------------------|
| **Metode**       | POST                                |
| **Endpoint**     | `/api/v1/students`                  |
| **Content-Type** | application/json                    |
| **Status Sukses**| 201 + header `Location`             |
| **Status Gagal** | 400, 409, 415, 422                  |

**Body:**

```json
{
  "nim": "434241025",
  "name": "Dina",
  "grade": 3.70,
  "is_active": true
}
```

**Respons 201:**

```
Header → Location: /api/v1/students/4
```

```json
{
  "success": true,
  "message": "mahasiswa berhasil ditambahkan",
  "data": {
    "id": 4,
    "nim": "434241025",
    "name": "Dina",
    "grade": 3.70,
    "is_active": true,
    "created_at": "2026-09-04T11:52:36.82+07:00"
  }
}
```

**Respons 409 (NIM duplikat):**

```json
{
  "success": false,
  "message": "nim sudah digunakan oleh mahasiswa lain"
}
```

**Respons 415 (Content-Type salah):**

```json
{
  "success": false,
  "message": "Content-Type harus application/json"
}
```

**Respons 422 (validasi gagal):**

```json
{
  "success": false,
  "message": "validasi gagal",
  "errors": {
    "nim": "nim tidak boleh kosong",
    "name": "name tidak boleh kosong"
  }
}
```

---

### 4. PUT /api/v1/students/:id

Ganti seluruh data mahasiswa. Semua field wajib dikirim.

| Item             | Detail                              |
|------------------|-------------------------------------|
| **Metode**       | PUT                                 |
| **Endpoint**     | `/api/v1/students/:id`              |
| **Content-Type** | application/json                    |
| **Status Sukses**| 200                                 |
| **Status Gagal** | 400, 404, 409, 415, 422             |

**Body:**

```json
{
  "nim": "434241022",
  "name": "Putri Apriliana",
  "grade": 3.95,
  "is_active": true
}
```

**Respons 200:**

```json
{
  "success": true,
  "message": "data mahasiswa berhasil diganti seluruhnya (PUT)",
  "data": {
    "id": 1,
    "nim": "434241022",
    "name": "Putri Apriliana",
    "grade": 3.95,
    "is_active": true,
    "created_at": "2026-09-04T12:16:50.46+07:00"
  }
}
```

---

### 5. PATCH /api/v1/students/:id

Ubah sebagian data mahasiswa. Cuma field yang dikirim yang berubah, sisanya tetap.

| Item             | Detail                              |
|------------------|-------------------------------------|
| **Metode**       | PATCH                               |
| **Endpoint**     | `/api/v1/students/:id`              |
| **Content-Type** | application/json                    |
| **Status Sukses**| 200                                 |
| **Status Gagal** | 400, 404, 409, 415, 422             |

**Body (misal cuma ubah grade):**

```json
{
  "grade": 3.99
}
```

**Respons 200:**

```json
{
  "success": true,
  "message": "data mahasiswa berhasil diperbarui sebagian (PATCH)",
  "data": {
    "id": 1,
    "nim": "434241022",
    "name": "Putri Apriliana",
    "grade": 3.99,
    "is_active": true,
    "created_at": "2026-09-04T12:16:50.46+07:00"
  }
}
```

**Bedanya PUT sama PATCH:**
- PUT → kirim semua field, data lama ditimpa total
- PATCH → kirim field tertentu aja, yang lain tidak keubah

**Bukti pengujian di data yang sama (`id: 1`):**

| Tahap | `name` | `grade` |
|---|---|---|
| Sebelum | `"Putri"` | `3.89` |
| Setelah PUT (kirim semua field) | `"Putri Apriliana"` | `3.95` |
| Setelah PATCH (kirim `grade` aja) | `"Putri Apriliana"` *(tetap)* | `3.99` |

---

### 6. DELETE /api/v1/students/:id

Hapus data mahasiswa.

| Item             | Detail                              |
|------------------|-------------------------------------|
| **Metode**       | DELETE                              |
| **Endpoint**     | `/api/v1/students/:id`              |
| **Body**         | —                                   |
| **Status Sukses**| 204 (tanpa body)                    |
| **Status Gagal** | 400, 404                            |

**Respons 404:**

```json
{
  "success": false,
  "message": "mahasiswa tidak ditemukan"
}
```

---

## Status Code

| Status | Kapan muncul                                      |
|--------|---------------------------------------------------|
| 200    | Data berhasil diambil atau diubah                 |
| 201    | Data baru berhasil dibuat + header `Location`      |
| 204    | Data berhasil dihapus, tidak ada body              |
| 400    | JSON rusak atau ID bukan angka                     |
| 404    | Data yang dicari tidak ada                         |
| 409    | NIM bentrok sama data yang sudah ada               |
| 415    | Content-Type bukan `application/json`              |
| 422    | Ada field yang tidak lolos validasi                 |

---

## Tech Stack

- Go 1.26
- Fiber v2.52
- Data disimpan di memory (slice), belum pakai database
