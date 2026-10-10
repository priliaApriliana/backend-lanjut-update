# LAPORAN PRAKTIKUM PEMROGRAMAN BACKEND LANJUT
## MODUL 6: AUTHORIZATION & ROLE-BASED ACCESS CONTROL (RBAC)

---

**Informasi Mahasiswa:**
- **Nama:** Putri Apriliana
- **NIM:** 434241022
- **Kelas:** D4 Teknik Informatika
- **Mata Kuliah:** Praktikum Pemrograman Backend Lanjut
- **Tautan Repositori GitHub:** `https://github.com/priliaApriliana/backend-lanjut`

---

## DAFTAR ISI
1. [Bagian 1: Landasan Teori Authorization & RBAC](#bagian-1-landasan-teori-authorization--rbac)
2. [Bagian 2: Skema Basis Data dan Migrasi RBAC](#bagian-2-skema-basis-data-dan-migrasi-rbac)
3. [Bagian 3: Implementasi Arsitektural & Potongan Kode Penting](#bagian-3-implementasi-arsitektural--potongan-kode-penting)
   - [3.1. Struktur Data In-Memory Fail-Closed (`helper/authz.go`)](#31-struktur-data-in-memory-fail-closed-helperauthzgo)
   - [3.2. Pemuatan Permission Sekali saat Startup (`app/repository/role_repository.go`)](#32-pemuatan-permission-sekali-saat-startup-apprepositoryrole_repositorygo)
   - [3.3. Middleware Pemeriksaan Hak Akses (`middleware/authz.go`)](#33-middleware-pemeriksaan-hak-akses-middlewareauthzgo)
   - [3.4. Aturan Kepemilikan Domain Murni (`app/service/student_authz_rules.go`)](#34-aturan-kepemilikan-domain-murni-appservicestudent_authz_rulesgo)
   - [3.5. Penegakan Otorisasi pada Service Mahasiswa (`app/service/student_service.go`)](#35-penegakan-otorisasi-pada-service-mahasiswa-appservicestudent_servicego)
   - [3.6. Audit Trail & Pencatatan Identitas pada Logger (`middleware/middleware.go`)](#36-audit-trail--pencatatan-identitas-pada-logger-middlewaremiddlewarego)
4. [Bagian 4: Matriks Hak Akses & Bukti Pengujian Endpoint](#bagian-4-matriks-hak-akses--bukti-pengujian-endpoint)
   - [4.1. Tabel Matriks Hak Akses Endpoint Users](#41-tabel-matriks-hak-akses-endpoint-users)
   - [4.2. Tabel Matriks Hak Akses Endpoint Students](#42-tabel-matriks-hak-akses-endpoint-students)
5. [Bagian 5: Bukti Pengujian Negatif (Security Negative Cases)](#bagian-5-bukti-pengujian-negatif-security-negative-cases)
6. [Bagian 6: Analisis Singkat & Jawaban Pertanyaan Bagian C.4](#bagian-6-analisis-singkat--jawaban-pertanyaan-bagian-c4)
7. [Bagian 7: Hasil Eksekusi Unit Test Murni (`go test -v`)](#bagian-7-hasil-eksekusi-unit-test-murni-go-test--v)
8. [Bagian 8: Riwayat Commit Git Bertahap](#bagian-8-riwayat-commit-git-bertahap)
9. [Bagian 9: Pernyataan Penggunaan Sumber Bantuan & AI](#bagian-9-pernyataan-penggunaan-sumber-bantuan--ai)

---

## BAGIAN 1: LANDASAN TEORI AUTHORIZATION & RBAC

### 1.1. Mengapa Autentikasi Saja Tidak Cukup?
Pada Modul 5, sistem telah berhasil memverifikasi identitas pengguna melalui tanda tangan digital JSON Web Token (JWT). Namun, tanpa lapisan otorisasi (*authorization*), sistem masih memiliki celah keamanan masif: setiap pengguna yang berhasil login (bahkan pengguna baru bertaraf `user` biasa) dapat melihat seluruh pengguna, memodifikasi data orang lain, dan menghapus entitas penting.

Autentikasi hanya menjawab pertanyaan: **"Siapa Anda?"** (`401 Unauthorized`).
Otorisasi menjawab pertanyaan krusial berikutnya: **"Apakah Anda boleh melakukan tindakan ini?"** (`403 Forbidden`).

### 1.2. Urutan Penjaga: Mengapa `RequireAuth` Selalu Mendahului `RequirePermission`?
Urutan pemasangan middleware tidak boleh dibalik:
```go
// URUTAN BENAR:
users.Get("/", middleware.RequireAuth(jwt), middleware.RequirePermission(perms, "user:list"), handler)
```
Otorisasi hanya bermakna jika identitas pemanggil sudah diketahui dan diverifikasi. Jika `RequirePermission` dipasang sebelum `RequireAuth`, middleware izin akan membaca objek identitas kosong dari context (`c.Locals`), sehingga otomatis menolak request dengan status `401 Unauthorized` karena pemanggil belum terautentikasi.

### 1.3. Perbandingan Model Kontrol Akses: ACL, RBAC, dan ABAC
1. **ACL (Access Control List):** Hak akses dilekatkan langsung pada setiap objek/baris data. Cocok untuk sistem berkas, namun jumlah entri membengkak drastis seiring bertambahnya objek dan pengguna.
2. **RBAC (Role-Based Access Control):** Pengguna dikelompokkan ke dalam *Role*, dan *Role* memiliki sekumpulan *Permission*. Model ini sangat modular dan mudah dikelola untuk aplikasi bisnis enterprise.
3. **ABAC (Attribute-Based Access Control):** Keputusan akses dihitung secara dinamis berdasarkan kombinasi atribut subjek, objek, dan konteks lingkungan (waktu, IP, relasi). Sangat fleksibel namun kompleks dalam pemeliharaan dan pengujian.

Aplikasi kita menerapkan **kombinasi RBAC murni dengan aturan kepemilikan (*ownership*)**, yang merupakan subset pragmatis dari ABAC.

### 1.4. Konvensi Penamaan Permission & Arti Sufiks `:any`
Permission dinamai menggunakan pola standar `domain:action:scope`:
- `student:list` $\rightarrow$ Melihat daftar ringkasan seluruh mahasiswa.
- `student:read:any` $\rightarrow$ Membaca detail mahasiswa mana pun, bukan hanya milik sendiri.
- `student:update:any` $\rightarrow$ Mengubah data mahasiswa milik siapa pun.
- `student:delete` $\rightarrow$ Menghapus data mahasiswa.

Sufiks `:any` menegaskan bahwa izin tersebut berlaku untuk data milik orang lain. Hak atas data milik sendiri mengalir secara alami dari **kepemilikan (*ownership*)**, bukan dari izin khusus.

### 1.5. Prinsip Fail-Closed
Sistem keamanan harus selalu bersikap **Fail-Closed**: saat terjadi keraguan, nilai nil, role tidak dikenal, atau permission salah ketik, sistem **selalu menolak akses** (`return false` / `403 Forbidden`). Sikap *fail-open* (mengizinkan saat ragu) adalah salah satu penyebab utama insiden kebocoran data terburuk di dunia.

---

## BAGIAN 2: SKEMA BASIS DATA DAN MIGRASI RBAC

### 2.1. Berkas Migrasi `migrations/003_rbac.sql`
Menyiapkan tabel relasional `roles`, `permissions`, dan tabel penghubung `role_permissions`:
```sql
CREATE TABLE IF NOT EXISTS roles (
  name VARCHAR(20) PRIMARY KEY,
  description VARCHAR(150) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS permissions (
  name VARCHAR(50) PRIMARY KEY,
  description VARCHAR(150) NOT NULL
);

CREATE TABLE IF NOT EXISTS role_permissions (
  role_name VARCHAR(20) NOT NULL REFERENCES roles(name) ON DELETE CASCADE,
  permission_name VARCHAR(50) NOT NULL REFERENCES permissions(name) ON DELETE CASCADE,
  PRIMARY KEY (role_name, permission_name)
);

-- Penguncian constraint integritas referensial pada users.role
UPDATE users SET role = 'user' WHERE role NOT IN (SELECT name FROM roles);
ALTER TABLE users ADD CONSTRAINT users_role_fkey FOREIGN KEY (role) REFERENCES roles(name) ON UPDATE CASCADE;
```

### 2.2. Berkas Migrasi `migrations/004_student_permissions.sql`
Menambahkan hak akses entitas mahasiswa serta kolom `owner_id`:
```sql
INSERT INTO permissions (name, description) VALUES
  ('student:list', 'Melihat daftar seluruh mahasiswa'),
  ('student:read:any', 'Melihat data mahasiswa mana pun'),
  ('student:create', 'Menambahkan data mahasiswa baru'),
  ('student:update:any', 'Mengubah data mahasiswa mana pun'),
  ('student:delete', 'Menghapus data mahasiswa')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name) VALUES
  ('admin', 'student:list'), ('admin', 'student:read:any'), ('admin', 'student:create'),
  ('admin', 'student:update:any'), ('admin', 'student:delete'),
  ('staff', 'student:list'), ('staff', 'student:read:any'), ('staff', 'student:create')
ON CONFLICT DO NOTHING;

ALTER TABLE students ADD COLUMN IF NOT EXISTS owner_id INTEGER;

-- Mengamankan data lama agar tidak melanggar foreign key constraint
UPDATE students SET owner_id = (SELECT id FROM users ORDER BY id ASC LIMIT 1)
WHERE owner_id IS NULL AND EXISTS (SELECT 1 FROM users);

ALTER TABLE students ADD CONSTRAINT students_owner_id_fkey
  FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS students_owner_id_idx ON students (owner_id);
```

---

## BAGIAN 3: IMPLEMENTASI ARSITEKTURAL & POTONGAN KODE PENTING

### 3.1. Struktur Data In-Memory Fail-Closed (`helper/authz.go`)
Pencarian hak akses menggunakan struktur `map[string]map[string]struct{}` untuk performa $O(1)$ tanpa membebani query database berulang kali:
```go
func (p *PermissionSet) Can(role, permission string) bool {
    if p == nil {
        return false // Fail closed
    }
    permissions, ok := p.byRole[role]
    if !ok {
        return false // Role tidak dikenal -> Tolak
    }
    _, granted := permissions[permission]
    return granted
}
```

### 3.2. Pemuatan Permission Sekali saat Startup (`app/repository/role_repository.go`)
Query SQL menggunakan `LEFT JOIN` dan `COALESCE` agar role yang belum memiliki izin (seperti `user`) tetap terdata dalam himpunan role yang sah:
```go
func (r *rolePostgresRepository) LoadPermissions(ctx context.Context) (map[string][]string, error) {
    rows, err := r.pool.Query(ctx,
        `SELECT r.name, COALESCE(rp.permission_name, '')
        FROM roles r
        LEFT JOIN role_permissions rp ON rp.role_name = r.name
        ORDER BY r.name, rp.permission_name`)
    // ...
}
```

### 3.3. Middleware Pemeriksaan Hak Akses (`middleware/authz.go`)
```go
func RequirePermission(perms *helper.PermissionSet, permission string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        user, ok := helper.CurrentUser(c)
        if !ok {
            return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
        }
        if !perms.Can(user.Role, permission) {
            return helper.Fail(c, fiber.StatusForbidden,
                "role "+user.Role+" tidak memiliki hak "+permission)
        }
        return c.Next()
    }
}
```

### 3.4. Aturan Kepemilikan Domain Murni (`app/service/student_authz_rules.go`)
Fungsi murni tanpa ketergantungan pada framework Fiber maupun SQL:
```go
func CanAccessStudent(
    current model.AuthUser,
    ownerID int,
    perms *helper.PermissionSet,
    anyPermission string,
) bool {
    // 1. Kepemilikan (ownership) — pemilik data selalu diizinkan
    if current.UserID == ownerID {
        return true
    }
    // 2. Hak Istimewa (RBAC) — izin global :any
    if perms == nil {
        return false
    }
    return perms.Can(current.Role, anyPermission)
}
```

### 3.5. Penegakan Otorisasi pada Service Mahasiswa (`app/service/student_service.go`)
- **Pencegahan Spoofing `owner_id`:** Saat `POST /students`, kolom `owner_id` diisi secara otomatis dari `current.UserID`.
- **Pemeriksaan Hak Sebelum Eksekusi Mutasi:**
```go
// Pada GET /students/:id
if !CanAccessStudent(current, student.OwnerID, s.perms, "student:read:any") {
    return helper.Fail(c, fiber.StatusForbidden, "tidak berhak mengakses data mahasiswa lain")
}

// Pada PUT /students/:id dan PATCH /students/:id
if !CanAccessStudent(current, existing.OwnerID, s.perms, "student:update:any") {
    return helper.Fail(c, fiber.StatusForbidden, "tidak berhak mengubah data mahasiswa lain")
}
```

### 3.6. Audit Trail & Pencatatan Identitas pada Logger (`middleware/middleware.go`)
Setiap request HTTP yang diautentikasi mencatat `user_id` dan `role` ke dalam output log JSON (`logs/app.log`), memungkinkan audit penolakan 403:
```go
if user, ok := helper.CurrentUser(c); ok {
    attrs = append(attrs,
        slog.Int("user_id", user.UserID),
        slog.String("role", user.Role),
    )
}
```

---

## BAGIAN 4: MATRIKS HAK AKSES & BUKTI PENGUJIAN ENDPOINT

Seluruh status code di bawah ini merupakan **hasil eksekusi terminal nyata** terhadap server lokal:

### 4.1. Tabel Matriks Hak Akses Endpoint Users
| Request / Endpoint | Admin (ID 3) | Staff (ID 4) | User (ID 5 - Pemilik) | User (Bukan Pemilik) |
| :--- | :---: | :---: | :---: | :---: |
| `GET /api/v1/users` | **200 OK** | **200 OK** | **403 Forbidden** | **403 Forbidden** |
| `GET /api/v1/users/:id` (diri sendiri) | **200 OK** | **200 OK** | **200 OK** | - |
| `GET /api/v1/users/:id` (orang lain) | **200 OK** | **200 OK** | - | **403 Forbidden** |
| `PUT /api/v1/users/:id` (orang lain) | **200 OK** | **403 Forbidden** | - | **403 Forbidden** |
| `DELETE /api/v1/users/:id` (orang lain) | **204 No Content** | **403 Forbidden** | - | **403 Forbidden** |
| `DELETE /api/v1/users/:id` (diri sendiri) | **403 Forbidden** | **403 Forbidden** | **403 Forbidden** | - |
| `PATCH /api/v1/users/:id/role` (orang lain) | **200 OK** | **403 Forbidden** | - | **403 Forbidden** |
| `PATCH /api/v1/users/:id/role` (diri sendiri) | **422 Unprocessable**| **403 Forbidden** | **403 Forbidden** | - |
| Tanpa Header `Authorization` | **401 Unauthorized**| **401 Unauthorized**| **401 Unauthorized** | **401 Unauthorized** |

### 4.2. Tabel Matriks Hak Akses Endpoint Students
| Request / Endpoint | Admin (ID 3) | Staff (ID 4) | User (ID 5 - Pemilik) | User (Bukan Pemilik) |
| :--- | :---: | :---: | :---: | :---: |
| `GET /api/v1/students` | **200 OK** | **200 OK** | **403 Forbidden** | **403 Forbidden** |
| `GET /api/v1/students/:id` (milik sendiri) | **200 OK** | **200 OK** | **200 OK** | - |
| `GET /api/v1/students/:id` (milik orang lain) | **200 OK** | **200 OK** | - | **403 Forbidden** |
| `POST /api/v1/students` | **201 Created** | **201 Created** | **403 Forbidden** | **403 Forbidden** |
| `PUT /api/v1/students/:id` (milik sendiri) | **200 OK** | **200 OK** | **200 OK** | - |
| `PUT /api/v1/students/:id` (milik orang lain) | **200 OK** | **403 Forbidden** | - | **403 Forbidden** |
| `PATCH /api/v1/students/:id` (milik sendiri) | **200 OK** | **200 OK** | **200 OK** | - |
| `PATCH /api/v1/students/:id` (milik orang lain)| **200 OK** | **403 Forbidden** | - | **403 Forbidden** |
| `DELETE /api/v1/students/:id` | **204 No Content** | **403 Forbidden** | **403 Forbidden** | **403 Forbidden** |

---

## BAGIAN 5: BUKTI PENGUJIAN NEGATIF (SECURITY NEGATIVE CASES)

### 5.1. Kasus Negatif 1: Eksploitasi Spoofing `owner_id` pada `POST /students`
Penyerang role `user` mencoba menyisipkan `"owner_id": 1` di dalam body JSON untuk mendaftarkan mahasiswa atas nama pengguna lain:
```bash
curl.exe -i -s -X POST http://127.0.0.1:3000/api/v1/students \
  -H "Authorization: Bearer <USER_TOKEN>" \
  -H "Content-Type: application/json" \
  --data '{\"nim\":\"434241073\",\"name\":\"UserStudent\",\"grade\":85.0,\"is_active\":true,\"owner_id\":1}'
```
**Respons Terminal:**
```http
HTTP/1.1 403 Forbidden
Content-Type: application/json
Content-Length: 73

{"success":false,"message":"role user tidak memiliki hak student:create"}
```
*(Percobaan ditolak langsung oleh middleware karena role `user` tidak memiliki hak `student:create`).*

### 5.2. Kasus Negatif 2: Manipulasi Data Mahasiswa Milik Orang Lain oleh Role User
Penyerang role `user` mencoba melakukan `PUT /students/1` pada data mahasiswa milik akun Admin (ID 3):
```bash
curl.exe -i -s -X PUT http://127.0.0.1:3000/api/v1/students/1 \
  -H "Authorization: Bearer <USER_TOKEN>" \
  -H "Content-Type: application/json" \
  --data '{\"nim\":\"434241022\",\"name\":\"HackedPutri\",\"grade\":100.0,\"is_active\":true}'
```
**Respons Terminal:**
```http
HTTP/1.1 403 Forbidden
Content-Type: application/json
Content-Length: 71

{"success":false,"message":"tidak berhak mengubah data mahasiswa lain"}
```
*(Percobaan ditolak oleh fungsi `CanAccessStudent` di layer service karena pemanggil bukan pemilik dan tidak memiliki hak `student:update:any`).*

### 5.3. Kasus Negatif 3: Pencegahan Admin Menurunkan Role Dirinya Sendiri
Admin mencoba melakukan `PATCH /users/3/role` dengan body `{"role":"user"}` pada akunnya sendiri:
```bash
curl.exe -i -s -X PATCH http://127.0.0.1:3000/api/v1/users/3/role \
  -H "Authorization: Bearer <ADMIN_TOKEN>" \
  -H "Content-Type: application/json" \
  --data '{\"role\":\"user\"}'
```
**Respons Terminal:**
```http
HTTP/1.1 422 Unprocessable Entity
Content-Type: application/json
Content-Length: 95

{"success":false,"message":"validasi gagal","errors":{"role":"tidak boleh mengubah role diri sendiri"}}
```
*(Sistem mencegah lockout permanen di mana aplikasi kehilangan seluruh akun admin).*

### 5.4. Kasus Negatif 4: Pencegahan Pengguna Menghapus Akunnya Sendiri
Admin mencoba mengeksekusi `DELETE /users/3` pada akunnya sendiri:
```bash
curl.exe -i -s -X DELETE http://127.0.0.1:3000/api/v1/users/3 \
  -H "Authorization: Bearer <ADMIN_TOKEN>"
```
**Respons Terminal:**
```http
HTTP/1.1 403 Forbidden
Content-Type: application/json
Content-Length: 68

{"success":false,"message":"tidak boleh menghapus akun sendiri"}
```

---

## BAGIAN 6: ANALISIS SINGKAT & JAWABAN PERTANYAAN BAGIAN C.4

### 6.1. Mengapa Pemeriksaan Kepemilikan Tidak Dapat Dipindahkan ke Middleware?
Pemeriksaan kepemilikan data tidak dapat diletakkan di middleware karena pada fase middleware, server **hanya mengetahui identitas pemanggil** dari token (`c.Locals("authUser")`) dan parameter URL mentah (`:id`), tetapi **belum mengetahui siapa pemilik baris data tersebut** (`owner_id`). Jika pemeriksaan dipaksakan di middleware, maka middleware harus melakukan query `SELECT owner_id FROM students WHERE id = $1` ke database. Hal ini melanggar prinsip *Clean Architecture* (middleware tercemar oleh logika detail persistensi database) dan memicu inefisiensi query ganda (*duplicate query*), di mana database dibaca sekali di middleware dan dibaca kembali di layer service/repository.

### 6.2. Risiko Dua Tempat Pemeriksaan Akses (Route & Service) dan Cara Mitigasinya
Risiko terbesarnya adalah **asimetri pengawasan dan kelalaian pengembang (*inconsistent enforcement gap*)**, di mana seorang pengembang baru menambahkan method baru di layer service tetapi lupa mendaftarkan middleware `RequirePermission` di file `route/route.go`, atau sebaliknya memasang permission di route tetapi lupa memvalidasi kepemilikan di service sehingga membuka celah IDOR (*Insecure Direct Object Reference*). Mitigasi konkretnya adalah dengan menetapkan arsitektur konvensi yang ketat: seluruh endpoint wajib dibungkus dalam route group yang mewajibkan `RequireAuth`, pembuatan *automated integration tests* yang menguji matriks hak akses secara komprehensif, serta audit route berkala saat proses *Code Review*.

### 6.3. Kebutuhan "Dosen Wali Hanya Boleh Melihat Mahasiswa Bimbingannya": Apakah RBAC Masih Memadai?
Model RBAC murni **tidak lagi memadai** untuk kebutuhan tersebut. RBAC hanya mampu memetakan role secara statis (misalnya peran `dosen_wali` memiliki izin `student:read`), namun tidak dapat mengevaluasi relasi kontekstual dinamis antar-entitas: "apakah mahasiswa X terdaftar di bawah bimbingan dosen Y". Untuk menyelesaikannya, sistem harus ditingkatkan ke model **ReBAC (Relationship-Based Access Control)** atau **ABAC (Attribute-Based Access Control)**, di mana tabel relasi perwalian (`academic_advisors`) diperiksa di layer service untuk memverifikasi apakah `advisor_id == current.UserID` sebelum data mahasiswa dikembalikan ke klien.

---

## BAGIAN 7: HASIL EKSEKUSI UNIT TEST MURNI (`go test -v`)

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
--- PASS: TestCheckPasswordStrength (0.00s)
=== RUN   TestValidateRegister
--- PASS: TestValidateRegister (0.00s)
=== RUN   TestValidateLogin
--- PASS: TestValidateLogin (0.00s)
=== RUN   TestIsValidUsername
--- PASS: TestIsValidUsername (0.00s)
=== RUN   TestCanAccessUser
=== RUN   TestCanAccessUser/Pemilik_mengakses_akunnya_sendiri_(Role_user)
=== RUN   TestCanAccessUser/Bukan_pemilik_tetapi_memiliki_permission_(Role_admin)
=== RUN   TestCanAccessUser/Bukan_pemilik_dan_tidak_memiliki_permission_(Role_user)
--- PASS: TestCanAccessUser (0.00s)
=== RUN   TestValidateAssignRole
=== RUN   TestValidateAssignRole/Assign_Role_Valid_ke_User_Lain
=== RUN   TestValidateAssignRole/Role_Kosong
=== RUN   TestValidateAssignRole/Role_Tidak_Dikenal
=== RUN   TestValidateAssignRole/Tidak_Boleh_Mengubah_Role_Diri_Sendiri
--- PASS: TestValidateAssignRole (0.00s)
=== RUN   TestCanAccessStudent
=== RUN   TestCanAccessStudent/Pemilik_data_mahasiswa_selalu_diizinkan_(Role_user)
=== RUN   TestCanAccessStudent/Staff_membaca_data_mahasiswa_milik_user_lain_(student:read:any)
=== RUN   TestCanAccessStudent/Staff_mencoba_mengubah_data_mahasiswa_milik_user_lain_(staff_tidak_punya_student:update:any)
=== RUN   TestCanAccessStudent/Admin_membaca_data_mahasiswa_milik_user_lain_(student:read:any)
=== RUN   TestCanAccessStudent/Admin_mengubah_data_mahasiswa_milik_user_lain_(student:update:any)
=== RUN   TestCanAccessStudent/User_lain_mengakses_data_mahasiswa_tanpa_hak_(Role_user_bukan_pemilik)
=== RUN   TestCanAccessStudent/User_lain_mengubah_data_mahasiswa_tanpa_hak_(Role_user_bukan_pemilik)
=== RUN   TestCanAccessStudent/Fail_Closed_saat_PermissionSet_bernilai_nil
--- PASS: TestCanAccessStudent (0.00s)
PASS
ok      api-students/app/service        1.026s
```

---

## BAGIAN 8: RIWAYAT COMMIT GIT BERTAHAP

Pengerjaan Modul 6 di-commit secara bertahap dan terstruktur untuk setiap lapisan arsitektur:

| Commit Hash | Pesan Commit (*Commit Message*) | Cakupan Lapisan |
| :--- | :--- | :--- |
| `9601ace` | `feat(modul-6): migrasi skema database RBAC (roles, permissions, role_permissions, owner_id)` | Migrasi SQL RBAC & model entitas user/student. |
| `d762d4f` | `feat(modul-6): implementasi RoleRepository dan pembaruan repositori untuk owner_id dan update role` | `RoleRepository`, UpdateRole, query owner_id pgxpool. |
| `d437148` | `feat(modul-6): implementasi aturan otorisasi fail closed, student ownership, dan unit testing authz` | Aturan `CanAccessStudent`, `CanAccessUser`, unit test. |
| `886c0c4` | `feat(modul-6): implementasi UserService, student ownership check, helper authz, dan AuthService profil permission` | Controller `UserService`, `StudentService`, profil permissions. |
| `ca4c356` | `feat(modul-6): implementasi middleware RequirePermission, audit logging, rute RBAC, main.go, dan konfigurasi` | Middleware authz, logging user_id/role, wiring `main.go`. |
| `[docs]` | `docs: tambahkan berkas LAPORAN_TUGAS_6.md lengkap` | Laporan praktikum lengkap untuk pengumpulan. |

---

## BAGIAN 9: PERNYATAAN PENGGUNAAN SUMBER BANTUAN & AI

Sesuai pedoman integritas akademik:

1. **Alat Bantu AI:** Pengerjaan tugas ini memanfaatkan bantuan asisten kecerdasan buatan (*AI Pair Programmer*) untuk perancangan skema relasional RBAC, penulisan fungsi murni verifikasi kepemilikan data, pengujian unit test fail-closed, dan penyusunan draf laporan praktikum.
2. **Pemahaman & Tanggung Jawab:** Mahasiswa memahami secara utuh prinsip kerja *Role-Based Access Control*, penegakan hak akses bertingkat (*Route Middleware vs Service Layer*), prinsip *Fail-Closed*, serta pemisahan tanggung jawab kode sesuai *Clean Architecture*. Seluruh kode sumber telah diuji secara nyata terhadap database PostgreSQL lokal dan seluruh pengujian keamanan terbukti berhasil.

---
*Laporan Praktikum Modul 6 disusun secara lengkap dan siap diekspor menjadi dokumen PDF (`Tugas6_434241022.pdf`).*
