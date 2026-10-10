# API Students — Authorization & RBAC (Modul 6)

Proyek REST API untuk manajemen data mahasiswa, autentikasi berbasis JWT, dan otorisasi berbasis peran (**Role-Based Access Control / RBAC**) menggunakan framework **Go Fiber v2**, driver PostgreSQL **pgx v5** (`pgxpool`), dan arsitektur **Clean Architecture**. Proyek ini dibuat untuk memenuhi penugasan **Modul Praktikum Pertemuan 6: Authorization & RBAC** pada mata kuliah Praktikum Pemrograman Backend Lanjut, Program Studi D4 Teknik Informatika, Fakultas Vokasi, Universitas Airlangga.

---

## 🛡️ Fitur Otorisasi & Keamanan (RBAC Features)

1. **Skema RBAC Relasional:**
   - Tabel `roles`: `admin`, `staff`, `user`.
   - Tabel `permissions`: format terstandar `domain:action:scope` (`user:list`, `user:read:any`, `user:update:any`, `user:delete`, `role:assign`, `student:list`, `student:read:any`, `student:create`, `student:update:any`, `student:delete`).
   - Tabel `role_permissions`: tabel penghubung relasi banyak-ke-banyak (Many-to-Many).
2. **In-Memory PermissionSet (Fail Closed):**
   - Pemetaan izin dimuat **sekali saat aplikasi menyala** (`LoadPermissions`) ke dalam memori struktur data `map[string]map[string]struct{}` untuk pencarian $O(1)$ tanpa membebani basis data pada setiap request.
   - Mengikuti prinsip **Fail Closed**: jika role atau permission tidak dikenal atau `PermissionSet == nil`, akses **selalu ditolak**.
3. **Pemisahan Keputusan Akses (Middleware vs Service):**
   - **Di Middleware (`RequirePermission`):** Keputusan akses yang tidak membutuhkan data, seperti hak mengakses endpoint daftar (`GET /users`, `GET /students`) atau membuat data baru.
   - **Di Service (`CanAccessUser` & `CanAccessStudent`):** Keputusan akses berbasis kepemilikan data (*ownership*). Pemilik data (`owner_id == current.UserID`) selalu diizinkan mengakses dan mengubah datanya; selain pemilik, diperlukan izin global `:any`.
4. **Proteksi Integritas & Anti-Spoofing:**
   - Kolom `owner_id` pada tabel `students` diisi otomatis dari token pemanggil saat `POST /students` dan tidak dapat diubah oleh request body klien.
   - Admin tidak dapat menurunkan role dirinya sendiri (mencegah *lockout* sistem tanpa admin).
   - Pengguna tidak dapat menghapus akunnya sendiri.
5. **Audit Trail & Structured Logging:**
   - Setiap request yang melewati autentikasi mencatat `user_id` dan `role` pada atribut JSON logger (`logs/app.log`), memungkinkan pelacakan penolakan 403 (*Forbidden*).

---

## 📋 Matriks Hak Akses

| Endpoint / Operasi | Admin | Staff | User (Pemilik) | User (Bukan Pemilik) |
| :--- | :---: | :---: | :---: | :---: |
| `GET /api/v1/health` | 200 | 200 | 200 | 200 |
| `POST /api/v1/auth/*` | 200/201 | 200/201 | 200/201 | 200/201 |
| `GET /api/v1/users` | 200 | 200 | 403 | 403 |
| `GET /api/v1/users/:id` | 200 | 200 | 200 | 403 |
| `PUT /api/v1/users/:id` | 200 | 403 | 200 | 403 |
| `DELETE /api/v1/users/:id` (orang lain) | 204 | 403 | 403 | 403 |
| `DELETE /api/v1/users/:id` (diri sendiri) | 403 | 403 | 403 | 403 |
| `PATCH /api/v1/users/:id/role` (orang lain) | 200 | 403 | 403 | 403 |
| `PATCH /api/v1/users/:id/role` (diri sendiri) | 422 | 403 | 403 | 403 |
| `GET /api/v1/students` | 200 | 200 | 403 | 403 |
| `GET /api/v1/students/:id` | 200 | 200 | 200 | 403 |
| `POST /api/v1/students` | 201 | 201 | 403 | 403 |
| `PUT /api/v1/students/:id` | 200 | 403 | 200 | 403 |
| `PATCH /api/v1/students/:id` | 200 | 403 | 200 | 403 |
| `DELETE /api/v1/students/:id` | 204 | 403 | 403 | 403 |

---

## 🧪 Pengujian Unit & Menjalankan Server

Jalankan rangkaian unit testing:
```bash
go test ./app/service/... -v
```

Jalankan server aplikasi:
```bash
go run main.go
```
