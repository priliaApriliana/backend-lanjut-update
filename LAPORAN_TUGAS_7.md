# LAPORAN RESMI PRAKTIKUM
# PEMROGRAMAN BACKEND LANJUT (SEMESTER 5)
## PERTEMUAN 7 — ADVANCED API DESIGN

---

* **Mata Kuliah:** Praktikum Pemrograman Backend Lanjut (2 SKS)
* **Program Studi:** D4 Teknik Informatika — Fakultas Vokasi, Universitas Airlangga
* **Nama Mahasiswa:** Putri Apriliana
* **NIM:** 434241022
* **Dosen Pengampu:** Tim Dosen Pemrograman Backend Lanjut
* **Tautan Repositori GitHub:** [https://github.com/priliaApriliana/backend-lanjut-update](https://github.com/priliaApriliana/backend-lanjut-update)
* **Branch Pengerjaan:** `main` (Direktori: `latihan-fiber/modul 7`)

---

## DAFTAR ISI
1. [D.1 Berita Acara Perbaikan (9 Kesalahan Bagian B)](#d1-berita-acara-perbaikan)
2. [D.2 Validasi Deklaratif pada Entity Students](#d2-validasi-deklaratif-pada-entity-students)
3. [D.3 Keyset (Cursor) Pagination pada Entity Students](#d3-cursor-pagination-pada-entity-students)
4. [D.4 Content Negotiation dan Centralized Error Handling](#d4-content-negotiation-dan-centralized-error-handling)
5. [D.5 Analisis Singkat Desain API Lanjut](#d5-analisis-singkat-desain-api-lanjut)
6. [Daftar Periksa Ringkas (Spesifikasi Penerimaan C.8)](#daftar-periksa-ringkas-spesifikasi-penerimaan-c8)
7. [Riwayat Komit Git Modular](#riwayat-komit-git-modular)

---

## D.1 Berita Acara Perbaikan

Sesuai dengan instruksi modul, kode awal pada Bagian B memuat **SEMBILAN (9) kesalahan** yang sengaja ditanam: **tiga kesalahan ditolak compiler** dan **enam kesalahan lolos compiler tetapi salah secara perilaku**.

Berikut adalah tabel Berita Acara Perbaikan lengkap:

| No | Berkas dan Baris | Jenis Kesalahan | Gejala yang Diamati | Akar Masalah | Perbaikan Kode | Bukti Verifikasi |
|---|---|---|---|---|---|---|
| **1** | `helper/cursor.go` baris 33, 38, 43, 48, 53 | **Ditolak Compiler** | `go build ./...` gagal dengan pesan:<br>`undefined: ErrInvalidCursor` | Identifier `ErrInvalidCursor` digunakan berkali-kali pada fungsi `DecodeCursor` sebagai penanda kegagalan pembacaan cursor, namun variabel error tersebut belum dideklarasikan di package `helper`. | Menambahkan deklarasi sentinel error pada tingkat package:<br>```go\nvar ErrInvalidCursor = errors.New("cursor tidak valid")\n``` | Perintah `go build ./helper/...` sukses dikompilasi tanpa error `undefined: ErrInvalidCursor`. |
| **2** | `middleware/middleware.go` baris 48–56 | **Ditolak Compiler** | `go build ./...` gagal dengan pesan:<br>`declared and not used: status` | Variabel `status` telah dideklarasikan dan dikoreksi berdasarkan `appErr.Status`, namun pada slice `attrs` structured logger kode tetap memanggil `slog.Int("status", c.Response().StatusCode())`. Bahasa Go melarang variabel lokal yang dideklarasikan tetapi tidak pernah dibaca. | Mengganti pemanggilan fungsi menjadi variabel yang telah dievaluasi:<br>```go\nslog.Int("status", status),\n``` | Kompilasi package `middleware` berhasil; log request HTTP mencatat status response yang akurat. |
| **3** | `config/app.go` baris 74 | **Ditolak Compiler** | `go build ./...` gagal dengan pesan:<br>`appErr.cause undefined (cannot refer to unexported field cause)` | Field `cause` pada struct `helper.AppError` berawalan huruf kecil (*unexported*) untuk mencegah kebocoran pesan teknis ke client. Namun fungsi `newErrorHandler` di package `config` berusaha mengakses field private tersebut secara langsung (`appErr.cause.Error()`). | Menambahkan method publik `Cause() error` dan `Unwrap() error` pada `AppError`, lalu membaca error asli secara aman:<br>```go\nerrMsg := appErr.Error()\nif appErr.Cause() != nil {\n    errMsg = appErr.Cause().Error()\n}\n``` | Kompilasi package `config` sukses; error asli database tercatat di log internal tanpa melanggar enkapsulasi struct. |
| **4** | `helper/errors.go` baris 68–73 | **Lolos Compiler (Salah Perilaku)** | Permintaan registrasi/validasi dengan payload tidak sah mengembalikan status `HTTP 400 Bad Request`, bukan `HTTP 422 Unprocessable Entity`. | Fungsi konstruktor `Validation()` menginisialisasi `AppError` dengan `Status: fiber.StatusBadRequest`. Standar RESTful membedakan 400 (sintaks JSON rusak/malformed) dari 422 (semantik aturan validasi form/field dilanggar). | Mengubah status error menjadi 422:<br>```go\nfunc Validation(fields map[string]string) *AppError {\n    return &AppError{\n        Status: fiber.StatusUnprocessableEntity,\n        Code: CodeValidation,\n        Message: "validasi gagal",\n        Fields: fields,\n    }\n}\n``` | `curl -i -X POST .../auth/register` menghasilkan `HTTP/1.1 422 Unprocessable Entity` (Spesifikasi C.8 Butir 2). |
| **5** | `config/app.go` baris 68–83 | **Lolos Compiler (Salah Perilaku)** | Log mencatat kesalahan client (4xx) sebagai `ERROR` dengan pesan `"request_failed"`, sedangkan kegagalan fatal server (5xx) dicatat sebagai `WARN` dengan `"request_rejected"`. | Kondisi percabangan log terbalik (`if appErr.Status < fiber.StatusInternalServerError`). Kesalahan client (4xx < 500) justru masuk ke blok `logger.Error` yang berisik, dan fatal error (>= 500) masuk ke level peringatan rendah `logger.Warn`. | Memperbaiki kondisi percabangan sesuai RFC dan teori A.8:<br>```go\nif appErr.Status >= fiber.StatusInternalServerError {\n    logger.Error("request_failed", ...)\n} else {\n    logger.Warn("request_rejected", ...)\n}\n``` | Trigger endpoint 404/422 menghasilkan log level `WARN` (`request_rejected`); error database menghasilkan log level `ERROR` (`request_failed`) (Spesifikasi C.7 & C.8 Butir 8). |
| **6** | `app/service/user_service.go` baris 33–42 | **Lolos Compiler (Salah Perilaku)** | Ketika database mengalami kegagalan (misal tabel/kolom hilang), API mengembalikan response `HTTP 200 OK` dengan payload kosong, bukan 500. | Pada fungsi `TranslateError`, cabang `default:` mengembalikan `nil`. Akibatnya layer service menganggap tidak ada error dan memproses response seolah berhasil (`helper.Success`). Bertentangan dengan prinsip *Fail-Closed*. | Mengubah cabang default untuk mengembalikan Internal error:<br>```go\ndefault:\n    return helper.Internal(err)\n``` | Skema tabel diubah secara sengaja $\rightarrow$ server mengembalikan `HTTP 500 INTERNAL_ERROR`, tidak pernah 200 (Spesifikasi C.8 Butir 6). |
| **7** | `helper/validator.go` baris 49–51 | **Lolos Compiler (Salah Perilaku)** | Password kuat (misal `RahasiaSuper2026`) ditolak dengan pesan validasi kosong `""`, sedangkan password lemah (misal `12345678`) justru diterima. | Fungsi `passwordStrength(val)` mengembalikan `""` jika password valid/kuat, dan pesan error jika lemah. Pada validator ditulis: `return passwordStrength(...) != ""` sehingga logika boolean terbalik (menganggap string kosong sebagai tidak valid). | Mengubah logika pembanding menjadi sama dengan string kosong:<br>```go\n_ = v.RegisterValidation("strongpassword", func(fl validator.FieldLevel) bool {\n    return passwordStrength(fl.Field().String()) == ""\n})\n``` | Password kuat diterima (201/200); password tanpa angka atau terlalu umum ditolak dengan pesan yang tepat (Spesifikasi C.8 Butir 4). |
| **8** | `app/repository/user_repository.go` baris 269–272 | **Lolos Compiler (Salah Perilaku)** | Data users pada halaman 1 terurut dari yang paling lama ke yang terbaru (`ASC`), dan navigasi keyset cursor `<` melompati baris atau mengembalikan data kosong. | Query SQL menggunakan `ORDER BY created_at ASC, id ASC`. Klausa keyset perbandingan baris `WHERE (created_at, id) < ($1, $2)` dan indeks komposit `(created_at DESC, id DESC)` dirancang untuk urutan menurun (*latest first*). | Mengubah klausa ORDER BY menjadi DESC:<br>```go\nquery := fmt.Sprintf(\n    "SELECT %s FROM users%s ORDER BY created_at DESC, id DESC LIMIT $%d",\n    userColumns, where, len(args))\n``` | Halaman 1 terurut dari user terbaru ke terlama; navigasi cursor mengambil halaman 2 dengan konsisten (Spesifikasi C.8 Butir 10). |
| **9** | `helper/negotiate.go` baris 63–70 | **Lolos Compiler (Salah Perilaku)** | Endpoint `GET /users` dengan header `Accept: text/csv` mengembalikan file CSV yang kosong melompong (0 byte data). | Fungsi `WriteUsersCSV` menggunakan `csv.NewWriter(&buffer)` yang berbasis buffer internal `bufio`, tetapi tidak pernah memanggil `writer.Flush()` sebelum membaca `buffer.String()`. Data masih tertahan di buffer writer. | Memanggil `writer.Flush()` sebelum evaluasi error:<br>```go\nwriter.Flush()\nif err := writer.Error(); err != nil {\n    return Internal(err)\n}\n``` | File `users.csv` dan `students.csv` terunduh dengan baris header dan isi data lengkap (Spesifikasi C.8 Butir 13). |

*(Catatan Tambahan Arsitektur)*:
* Pada struct `PatchUserRequest` di `app/model/user.go`, field `Username` awalnya bertipe `string` (non-pointer). Hal ini menyebabkan komparasi `req.Username != nil` ditolak compiler (`mismatched types string and untyped nil`) serta menggagalkan fungsi tag `omitnil`. Seluruh field parsial PATCH telah diperbaiki menjadi pointer `*string` sehingga `omitnil` berfungsi sebagaimana mestinya.

---

## D.2 Validasi Deklaratif pada Entity Students

### 1. Perbandingan Kode SEBELUM dan SESUDAH (Model Request)

**SEBELUM (Modul 3–6, Validasi Manual Imperatif):**
```go
// Definisi struct tanpa aturan validasi (mudah lupa menambah aturan jika ada field baru)
type CreateStudentRequest struct {
    NIM      string  `json:"nim"`
    Name     string  `json:"name"`
    Grade    float64 `json:"grade"`
    IsActive bool    `json:"is_active"`
}

// Fungsi validasi manual terpisah jauh di layer service
func ValidateCreate(req model.CreateStudentRequest) map[string]string {
    errs := map[string]string{}
    if strings.TrimSpace(req.NIM) == "" {
        errs["nim"] = "NIM wajib diisi"
    }
    if strings.TrimSpace(req.Name) == "" {
        errs["name"] = "nama mahasiswa wajib diisi"
    }
    if !IsValidGrade(req.Grade) {
        errs["grade"] = "nilai grade harus berada pada rentang 0.0 sampai 100.0"
    }
    return errs
}
```

**SESUDAH (Modul 7, Validasi Deklaratif Menggunakan Tag Refleksi):**
```go
// Aturan berada persis pada baris definisi field (self-documenting & compiler-verified)
type CreateStudentRequest struct {
    NIM      string  `json:"nim" validate:"required,nim"`
    Name     string  `json:"name" validate:"required,min=3,max=100"`
    Grade    float64 `json:"grade" validate:"gte=0,lte=100"`
    IsActive bool    `json:"is_active"`
}

// Pada service layer, cukup satu baris generik:
if errs := helper.ValidateStruct(req); errs != nil {
    return helper.Validation(errs)
}
```

### 2. Custom Validation Domain Mahasiswa (`nim`)
Dibuat aturan custom validator `nim` pada `helper/validator.go` untuk memastikan format NIM sesuai standar universitas (terdiri dari 8–20 digit angka murni):
```go
_ = v.RegisterValidation("nim", func(fl validator.FieldLevel) bool {
    val := strings.TrimSpace(fl.Field().String())
    if len(val) < 8 || len(val) > 20 {
        return false
    }
    for _, r := range val {
        if !unicode.IsDigit(r) {
            return false
        }
    }
    return true
})
```
Terjemahan pesan berbahasa Indonesia diintegrasikan pada `messageFor`:
```go
case "nim":
    return "harus berupa digit angka dengan panjang 8-20 karakter"
case "gte":
    return "nilai minimal " + fe.Param()
case "lte":
    return "nilai maksimal " + fe.Param()
```

### 3. Pengujian Endpoint PATCH Students (`omitnil` vs Field Pointer)
Struct `PatchStudentRequest` didefinisikan menggunakan pointer dan tag `omitnil`:
```go
type PatchStudentRequest struct {
    NIM      *string  `json:"nim,omitempty" validate:"omitnil,nim"`
    Name     *string  `json:"name,omitempty" validate:"omitnil,min=3,max=100"`
    Grade    *float64 `json:"grade,omitempty" validate:"omitnil,gte=0,lte=100"`
    IsActive *bool    `json:"is_active,omitempty"`
}
```

**Uji Kasus A: PATCH dengan `name` bernilai string kosong `""`:**
```bash
curl -i -X PATCH http://localhost:3000/api/v1/students/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":""}'
```
**Output Server (DITOLAK dengan status 422):**
```http
HTTP/1.1 422 Unprocessable Entity
Content-Type: application/json

{
  "success": false,
  "code": "VALIDATION_ERROR",
  "message": "validasi gagal",
  "fields": {
    "name": "minimal 3 karakter"
  },
  "request_id": "73ae3ebe-3a1e-4e64-b56b-2db682ef9c41"
}
```

**Uji Kasus B: PATCH tanpa field `name` (hanya memperbarui grade):**
```bash
curl -i -X PATCH http://localhost:3000/api/v1/students/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"grade":92.5}'
```
**Output Server (DITERIMA dengan status 200 OK):**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "success": true,
  "message": "mahasiswa berhasil diperbarui",
  "data": {
    "id": 1,
    "nim": "434241022",
    "name": "Putri_Apriliana_S.Tr.Kom",
    "grade": 92.5,
    "is_active": true,
    "created_at": "2026-10-03T19:23:14.230455+07:00",
    "owner_id": 3
  }
}
```
*Analisis:* Tag `omitnil` melewatkan field `name` yang bernilai pointer `nil` (tidak dikirim). Namun ketika client mengirim `""`, pointer tidak `nil` (menunjuk memory string kosong), sehingga aturan `min=3` dievaluasi dan ditolak.

---

## D.3 Keyset (Cursor) Pagination pada Entity Students

### 1. Pemilihan Pasangan Column Pengurut dan Jaminan Keunikan
Pasangan column pengurut yang dipilih adalah **`(created_at, id)` descending (`ORDER BY created_at DESC, id DESC`)**.
* **Alasan Desain:** Kolom `created_at` merepresentasikan waktu pembuatan data. Namun dalam sistem konkurensi tinggi, dua mahasiswa dapat terdaftar pada mikrodetik yang persis sama. Jika hanya `created_at` yang dibandingkan, kedua baris akan saling menutupi (*tie*) sehingga data dapat terlewat atau terkirim dua kali. Kolom `id` (Primary Key auto-increment) bertindak sebagai *tie-breaker* yang pasti unik. Pasangan `(created_at, id)` menjamin *strict total ordering* di seluruh baris tabel.
* **Perbandingan Baris (Row Constructor Comparison):**
  PostgreSQL mendukung perbandingan tuple:
  $$\text{WHERE } (\text{created\_at}, \text{id}) < (\$1, \$2)$$
  yang secara semantik setara dengan `created_at < $1 OR (created_at = $1 AND id < $2)`, namun dapat mengeksploitasi composite index secara langsung.

### 2. Migration File (`migrations/005_cursor_index.sql`)
```sql
CREATE INDEX IF NOT EXISTS students_created_at_id_desc_idx
ON students (created_at DESC, id DESC);
```

### 3. Pembuktian Cursor Tidak Menduplikasi Data
Pengujian membuktikan keunggulan keyset pagination dibanding pagination offset klasik:
1. Client mengambil Halaman 1 (limit=3):
   * Data diterima: `user`, `staff`, `admin`.
   * Cursor baris terakhir: `MTc5MTYyMDU2NDc0NTE0ODAwMHwz` (merepresentasikan `1791620564745148000|3`).
2. Sebelum Halaman 2 diminta, disisipkan satu baris baru di posisi paling atas: `baru99` (ID 6).
3. Client meminta Halaman 2 menggunakan cursor dari Halaman 1:
   * Data diterima: `hacker`, `sari`.
* **Hasil Pengamatan:** Tidak ada satu pun baris dari Halaman 1 (`admin`) yang muncul kembali di Halaman 2, dan tidak ada baris yang melompati jendela tampilan. Kehadiran data baru tidak menggeser jendela penunjuk keyset.

### 4. Hasil `EXPLAIN ANALYZE` Penggunaan Indeks PostgreSQL
Pengujian dieksekusi langsung pada database `praktikum_backend`:

```sql
EXPLAIN ANALYZE 
SELECT id, nim, name, grade, is_active, created_at, COALESCE(owner_id, 0)
FROM students 
WHERE (created_at, id) < ('2026-10-10 15:34:32', 7) 
ORDER BY created_at DESC, id DESC 
LIMIT 10;
```

**Keluaran Query Plan Resmi:**
```text
Limit  (cost=0.13..8.14 rows=1 width=305) (actual time=0.004..0.004 rows=1.00 loops=1)
  Buffers: shared hit=2
  ->  Index Scan using students_created_at_id_desc_idx on students  (cost=0.13..8.14 rows=1 width=305) (actual time=0.003..0.003 rows=1.00 loops=1)
        Index Cond: (ROW(created_at, id) < ROW('2026-10-10 15:34:32+07'::timestamp with time zone, 7))
        Index Searches: 1
        Buffers: shared hit=2
Planning Time: 0.052 ms
Execution Time: 0.019 ms
```
*Bukti:* Query menggunakan **`Index Scan using students_created_at_id_desc_idx`** dengan execution time sangat cepat (**0.019 ms**). Query langsung melompat (*b-tree seek*) ke baris cursor tanpa melakukan pembacaan sekuensial tabel.

---

## D.4 Content Negotiation dan Centralized Error Handling

### 1. Pembuktian Format Respons
* **Permintaan CSV (`Accept: text/csv`):**
  Mengembalikan header `Content-Type: text/csv; charset=utf-8` dan `Content-Disposition: attachment; filename="students.csv"`.
  ```csv
  id,nim,name,grade,is_active,created_at,owner_id
  7,434241071,AdminStudent,88.00,true,2026-10-10T08:34:32Z,3
  1,434241022,Putri_Apriliana_S.Tr.Kom,92.50,true,2026-10-03T12:23:14Z,3
  ```
* **Permintaan Format Tidak Didukung (`Accept: application/xml`):**
  Mengembalikan status `HTTP 406 Not Acceptable`:
  ```json
  {
    "success": false,
    "code": "NOT_ACCEPTABLE",
    "message": "format yang diminta tidak tersedia, pilih salah satu dari: application/json, text/csv",
    "request_id": "72ae3ebe-3a1e-4e64-b56b-2db682ef9c41"
  }
  ```
* **Permintaan Default (`Accept: */*` atau kosong):**
  Dijawab dengan format standar `HTTP 200 OK` (JSON).

### 2. Bukti Penghapusan `helper.Fail` di Seluruh Endpoint
Dilakukan pencarian teks menggunakan `ripgrep` pada seluruh direktori `modul 7`:
```bash
grep -rn "helper.Fail" ./latihan-fiber/modul 7/
```
**Hasil:** `No results found` (0 kemunculan). Seluruh handler dan service hanya mengembalikan `*helper.AppError`.

### 3. Matriks Kode Error Endpoint Students

| Kode Error (`code`) | Status HTTP | Situasi Pemicu pada Endpoint Students |
|---|:---:|---|
| `VALIDATION_ERROR` | **422** | Aturan tag struct dilanggar (NIM bukan angka, nama < 3 karakter, grade di luar 0–100, PATCH string kosong). |
| `BAD_REQUEST` | **400** | Syntax body JSON malformed, parameter `:id` bukan angka positif, cursor query string rusak/cacat, body PATCH `{}` tanpa field. |
| `UNAUTHORIZED` | **401** | Header `Authorization` tidak ada, format token bukan Bearer, JWT token kedaluwarsa, atau tanda tangan tidak valid. |
| `FORBIDDEN` | **403** | User/Staff mencoba mengakses, mengubah, atau menghapus data mahasiswa yang dimiliki oleh user lain tanpa hak `student:*:any`. |
| `NOT_FOUND` | **404** | Mahasiswa dengan ID yang diminta tidak ada pada basis data. |
| `CONFLICT` | **409** | NIM mahasiswa yang didaftarkan sudah ada di sistem (pelanggaran *unique constraint*). |
| `UNSUPPORTED_MEDIA_TYPE` | **415** | Request `POST`/`PUT`/`PATCH` dikirim dengan `Content-Type` bukan `application/json`. |
| `NOT_ACCEPTABLE` | **406** | Header `Accept` meminta format yang tidak didukung server (misal `application/xml`). |
| `TOO_MANY_REQUESTS` | **429** | Percobaan request melebihi ambang batas rate limiter. |
| `INTERNAL_ERROR` | **500** | Terjadi kegagalan tak terduga pada server/database (koneksi putus, disk penuh, error SQL). Detail teknis disembunyikan dari client. |

---

## D.5 Analisis Singkat Desain API Lanjut

### 8. Kehilangan Kemampuan Lompat Halaman ("Halaman 3 dari 40") pada Keyset Pagination
Kehilangan navigasi acak seperti "lompat ke halaman 15" atau informasi "halaman 3 dari 40" merupakan konsekuensi fundamental dari keyset pagination, karena posisinya ditambatkan pada nilai baris data terakhir (*stateful pointer*) dan sengaja meniadakan operasi `COUNT(*)` demi performa $O(1)$. Untuk kasus domain entitas mahasiswa ini, jika API digunakan oleh sistem antarmuka akademik internal seperti modul evaluasi atau feed aktivitas yang digulir terus (*infinite scroll/stream*), kehilangan tersebut sepenuhnya dapat diterima. Namun, jika sistem ditujukan untuk modul pelaporan administrasi akademik yang memerlukan pencarian berbasis nomor halaman terstruktur untuk verifikasi berkas, keyset pagination bukanlah solusi yang tepat. Solusi yang akan saya ambil jika nomor halaman mutlak diperlukan adalah mengimplementasikan pendekatan **Hybrid Pagination**: mempertahankan Keysets Pagination untuk endpoint publik bervolume tinggi, dan menyediakan opsi Offset Pagination dengan teknik *Deferred Join* (`SELECT ... FROM students JOIN (SELECT id FROM students ... LIMIT 10 OFFSET 200) USING (id)`) yang dibatasi maksimum halamannya (misal maks halaman 100) serta menerapkan caching pada nilai `COUNT(*)`.

### 9. Keamanan Cursor dan Data yang Tidak Boleh Disimpan
Satu hal yang **MUTLAK TIDAK BOLEH** disimpan di dalam cursor adalah **identitas hak akses, status otorisasi, atau data rahasia/sensitif** (seperti `user_id`, `role`, `is_admin: true`, atau token keamanan). Nilai cursor dienkode menggunakan Base64 URL-safe, yang merupakan format representasi teks biasa (*encoding*) dan **BUKAN enkripsi**. Siapa pun dapat mendekode cursor tersebut dengan fungsi `base64 -d`, melihat isinya, dan mengubah nilainya dengan mudah. Bila status otorisasi atau peran disimpan di dalam cursor dan dipercaya oleh server sebagai dasar penyaringan data, penyerang dapat memalsukan nilai cursor (*cursor tampering/privilege escalation*) untuk melewati validasi otorisasi dan membajak data seluruh pengguna tanpa terdeteksi. Cursor hanya boleh memuat penanda posisi pengurutan baris publik (`created_at` dan `id`).

### 10. Aturan Validasi yang Tetap Tidak Dapat Dinyatakan sebagai Tag
Setelah sebagian besar aturan dipindahkan ke tag refleksi, aturan yang **TETAP TIDAK DAPAT** dinyatakan sebagai struct tag adalah **aturan validasi ketergantungan antar-field (*cross-field validation*) dan aturan yang membutuhkan state eksternal (basis data/IO)**. Contohnya pada entitas mahasiswa adalah aturan `IsEmptyPatch` (memastikan minimal ada salah satu dari `nim`, `name`, `grade`, atau `is_active` yang dikirim dalam request PATCH), serta aturan integritas bisnis seperti *"NIM tidak boleh duplikat dengan mahasiswa lain yang aktif"* atau *"Grade hanya boleh diubah oleh dosen pembimbing mahasiswa yang bersangkutan"*. Alasan teknisnya: tag validator membaca dan mengevaluasi satu field secara individual dalam satu waktu (*field-level reflection*); tag tidak memiliki konteks menyeluruh mengenai relasi kondisional antar field dinamis, dan tidak boleh diberi dependensi terhadap koneksi basis data karena akan melanggar prinsip *Single Responsibility* serta memperlambat eksekusi refleksi validasi.

---

## Daftar Periksa Ringkas (Spesifikasi Penerimaan C.8)

| No | Spesifikasi Perilaku yang Harus Benar | Status Verifikasi | Bukti Hasil Pengujian Faktual |
|:---:|---|:---:|---|
| 1 | `go build ./...` berjalan tanpa satu pun pesan kesalahan | **TERPENUHI** | Exit code 0, binary terkompilasi bersih tanpa peringatan. |
| 2 | Aturan tag dilanggar $\rightarrow$ status 422, bukan 400 | **TERPENUHI** | `POST /register` dengan invalid fields $\rightarrow$ `HTTP 422 Unprocessable Entity`. |
| 3 | Pesan validasi tidak pernah berupa string kosong | **TERPENUHI** | Seluruh field error memiliki deskripsi lengkap dalam bahasa Indonesia. |
| 4 | Password kuat DITERIMA; password lemah/umum DITOLAK | **TERPENUHI** | `RahasiaSuper2026` diterima; `password123` $\rightarrow$ ditolak `"password terlalu umum"`. |
| 5 | PATCH kosong $\rightarrow$ 400; PATCH `{"username":""}` $\rightarrow$ 422; PATCH sah $\rightarrow$ 200 | **TERPENUHI** | Teruji pada endpoint `/users` dan `/students` (status 400, 422, dan 200 akurat). |
| 6 | Kegagalan database $\rightarrow$ 500 INTERNAL_ERROR, tidak pernah 200 kosong | **TERPENUHI** | Kolom dirusak $\rightarrow$ `HTTP 500 INTERNAL_ERROR` dengan JSON standar. |
| 7 | Pesan asli database tidak pernah muncul pada response client | **TERPENUHI** | Response hanya memuat `"terjadi kesalahan pada server"`; SQL detail masuk ke log. |
| 8 | 4xx tercatat WARN (`request_rejected`); 5xx tercatat ERROR (`request_failed`) | **TERPENUHI** | Terverifikasi pada `logs/app.log` untuk seluruh variasi request. |
| 9 | Status pada access log sama dengan status yang diterima client | **TERPENUHI** | Access log `http_request` mencatat status 422, 401, 404, 500 (bukan 200). |
| 10 | Halaman 1 terurut dari yang TERBARU ke yang terlama | **TERPENUHI** | Query menggunakan `ORDER BY created_at DESC, id DESC`. |
| 11 | Halaman 2 tidak memuat satu pun baris yang sudah muncul di halaman 1 | **TERPENUHI** | Keyset cursor comparison `(created_at, id) < ($1, $2)` memisahkan halaman bersih. |
| 12 | Halaman terakhir tidak menyertakan `next_cursor` dan `has_more: false` | **TERPENUHI** | Evaluasi `len(rows) > limit` menghasilkan `has_more: false` tanpa cursor. |
| 13 | `Accept: text/csv` menghasilkan berkas BERISI data, bukan berkas kosong | **TERPENUHI** | Pemanggilan `writer.Flush()` memastikan seluruh baris data tertulis ke respons. |
| 14 | `Accept: application/xml` $\rightarrow$ 406; `Accept: */*` $\rightarrow$ 200 JSON | **TERPENUHI** | Negosiasi konten membalas 406 Not Acceptable atau fallback JSON default. |

---

## Riwayat Komit Git Modular

Pengerjaan Modul 7 dicatat secara modular dalam 4 komit terstruktur:
1. `59c3779` — `feat(modul-7): migrasi indeks keyset pagination dan validasi deklaratif tag pada model DTO`
2. `262db33` — `feat(modul-7): implementasi AppError terpusat, validator deklaratif, cursor encode/decode, dan content negotiation`
3. `3468dc6` — `feat(modul-7): implementasi FindAfterCursor pada repositori, rules murni, dan unit testing validasi deklaratif`
4. `deaf402` — `feat(modul-7): integrasi service terpusat error, middleware access log akurat, rute content negotiation, dan panduan`

---
*Laporan ini disusun dengan sebenar-benarnya berdasarkan hasil pengujian faktual dan kode program yang terverifikasi.*
