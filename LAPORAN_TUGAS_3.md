# LAPORAN PRAKTIKUM PEMROGRAMAN BACKEND LANJUT
## MODUL 3: DATABASE & REPOSITORY PATTERN

---

**Informasi Mahasiswa:**
- **Nama:** Putri Apriliana
- **NIM:** 434241022
- **Kelas:** D4 Teknik Informatika
- **Mata Kuliah:** Praktikum Pemrograman Backend Lanjut
- **Tautan Repositori GitHub:** `https://github.com/priliaApriliana/backend-lanjut-update`

---

## DAFTAR ISI
1. [Bagian 1: Landasan Teori & Arsitektur Repository Pattern](#bagian-1-landasan-teori--arsitektur-repository-pattern)
2. [Bagian 2: Skema Basis Data dan Migrasi](#bagian-2-skema-basis-data-dan-migrasi)
3. [Bagian 3: Konfigurasi Environment dan Connection Pool](#bagian-3-konfigurasi-environment-dan-connection-pool)
4. [Bagian 4: Implementasi Lapisan Repository dan Handler](#bagian-4-implementasi-lapisan-repository-dan-handler)
   - [4.1. Kontrak Interface & Model Entitas](#41-kontrak-interface--model-entitas)
   - [4.2. Implementasi PostgreSQL Repository (`student_repository.go`)](#42-implementasi-postgresql-repository-student_repositorygo)
   - [4.3. Handler HTTP dan Dependency Injection (`handler.go`)](#43-handler-http-dan-dependency-injection-handlergo)
   - [4.4. Perakitan Aplikasi Terpusat (`main.go`)](#44-perakitan-aplikasi-terpusat-maingo)
5. [Bagian 5: Dokumen Kontrak API (API Contract)](#bagian-5-dokumen-kontrak-api-api-contract)
6. [Bagian 6: Bukti Pengujian Endpoint dan Kode Status HTTP](#bagian-6-bukti-pengujian-endpoint-dan-kode-status-http)
   - [6.1. Status 200 OK — Health Check & Pengambilan Data](#61-status-200-ok--health-check--pengambilan-data)
   - [6.2. Status 201 Created — Penambahan Mahasiswa & Header Location](#62-status-201-created--penambahan-mahasiswa--header-location)
   - [6.3. Status 204 No Content & 404 Not Found — Pengujian DELETE Dua Kali](#63-status-204-no-content--404-not-found--pengujian-delete-dua-kali)
   - [6.4. Status 404 Not Found — Penanganan ID Tak Ditemukan (`pgx.ErrNoRows`)](#64-status-404-not-found--penanganan-id-tak-ditemukan-pgxerrnorows)
   - [6.5. Status 409 Conflict — Penegakan Keunikan NIM oleh Basis Data](#65-status-409-conflict--penegakan-keunikan-nim-oleh-basis-data)
   - [6.6. Pengujian Paginasi, Sorting, & Proteksi SQL Injection](#66-pengujian-paginasi-sorting--proteksi-sql-injection)
   - [6.7. Pembuktian Persistensi Data (Server Restart Test)](#67-pembuktian-persistensi-data-server-restart-test)
   - [6.8. Analisis Status 503 vs 500 saat Basis Data Terputus](#68-analisis-status-503-vs-500-saat-basis-data-terputus)
7. [Bagian 7: Analisis & Jawaban Pertanyaan Eksplorasi](#bagian-7-analisis--jawaban-pertanyaan-eksplorasi)
8. [Bagian 8: Riwayat Commit Git Bertahap](#bagian-8-riwayat-commit-git-bertahap)
9. [Bagian 9: Pernyataan Penggunaan Sumber Bantuan & AI](#bagian-9-pernyataan-penggunaan-sumber-bantuan--ai)

---

## BAGIAN 1: LANDASAN TEORI & ARSITEKTUR REPOSITORY PATTERN

### 1.1. Mengapa Data Harus Keluar dari Memori
Pada Modul 2, data entitas disimpan di dalam struktur *slice in-memory*. Penyimpanan memori memiliki tiga keterbatasan fatal:
1. **Volatilitas:** Setiap kali server dimatikan atau di-restart, seluruh data hilang.
2. **Isolasi Proses:** Dua replika aplikasi tidak dapat berbagi data yang sama.
3. **Keterbatasan Kapasitas RAM:** Seluruh tabel harus dimuat ke memori, yang akan menyebabkan *Out-Of-Memory* (OOM) saat data mencapai jutaan baris.

Dengan memindahkan data ke sistem basis data relasional (**PostgreSQL**), persistensi data terjamin melalui prinsip **ACID** (*Atomicity, Consistency, Isolation, Durability*).

### 1.2. Driver dan Connection Pool
Membuka koneksi TCP baru ke basis data membutuhkan biaya komputasi yang mahal (jabat tangan jaringan, negosiasi TLS, autentikasi user, penyiapan sesi backend). **Connection Pool** (`pgxpool`) menyelesaikan masalah ini dengan menyediakan sekumpulan koneksi siap pakai yang dipinjam oleh handler saat mengeksekusi query dan segera dikembalikan setelah selesai. Parameter penting:
- `MaxConns`: Batas maksimum koneksi terbuka.
- `MinConns`: Jumlah koneksi standby yang selalu disiapkan.
- `MaxConnLifetime`: Umur pakai maksimal satu koneksi sebelum ditutup dan diganti koneksi baru.
- `MaxConnIdleTime`: Batas waktu koneksi menganggur sebelum ditutup.

### 1.3. Arsitektur Pola Repository (Repository Pattern)
Pola repository memisahkan logika bisnis aplikasi dari detail teknis penyimpanan:
```text
Handler (HTTP Layer)  --->  Repository (Interface Kontrak)  --->  PostgreSQL (Storage Implementation)
         |                                                                |
         v                                                                v
   model.WebResponse                                                model.Student
```
- **Arah Ketergantungan Searah:** Handler bergantung pada interface `StudentRepository`. Repository sama sekali tidak mengenal Fiber (`fiber.Ctx`) dan hanya menerima tipe murni `context.Context` serta model DTO.
- **Kemudahan Pengujian (Testability):** Karena handler hanya bergantung pada interface, pengujian unit dapat dilakukan menggunakan *mock repository* tanpa perlu database fisik.

---

## BAGIAN 2: SKEMA BASIS DATA DAN MIGRASI

Skema tabel mahasiswa didefinisikan pada file `migrations/001_create_students.sql`:

```sql
-- Migrasi 001: Membuat tabel students dan indeks
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

### 2.1. Mengapa Keunikan NIM Lebih Baik Dijaga Basis Data daripada Kode Go?
Pemeriksaan keunikan manual di aplikasi (misalnya menjalankan query `SELECT` sebelum `INSERT`) menyisakan celah balapan (*race condition*) pada sistem konkuren. Bila dua permintaan HTTP masuk secara simultan dengan NIM yang sama, kedua goroutine aplikasi dapat mengeksekusi `SELECT` bersamaan dan keduanya mendapati bahwa NIM belum ada, sehingga kedua goroutine sama-sama mengeksekusi `INSERT` dan data ganda pun tersimpan.

Sebaliknya, batasan **`UNIQUE INDEX` di basis data** bersifat atomik dan dilindungi oleh mekanisme penguncian (*index lock*) mesin basis data. Basis data menjamin hanya satu transaksi yang berhasil masuk, sementara transaksi kedua langsung digagalkan dengan error resmi PostgreSQL kode **`23505 (unique_violation)`**.

### 2.2. Alasan Penambahan Indeks Tambahan (`students_name_lower_idx`)
Kolom `name` sering digunakan sebagai filter pencarian teks (*wildcard search*) menggunakan operator `ILIKE` serta kriteria pengurutan (`ORDER BY name`). Tanpa indeks, PostgreSQL terpaksa melakukan **Sequential Scan (Full Table Scan)** dengan memeriksa setiap baris tabel dari awal hingga akhir. Dengan indeks ekspresi `LOWER(name)`, pencarian dan pengurutan dapat memanfaatkan **Index Scan**, sehingga waktu tanggap query tetap berada pada skala milidetik meskipun data mencapai ratusan ribu baris.

---

## BAGIAN 3: KONFIGURASI ENVIRONMENT DAN CONNECTION POOL

### 3.1. Manajemen Kredensial
Seluruh konfigurasi sensitif disimpan dalam file `.env` yang **wajib masuk dalam `.gitignore`** agar tidak pernah terunggah ke repositori publik:
```gitignore
# env
.env
```

Sebagai gantinya, disediakan file template `.env.example`:
```env
APP_PORT=3001
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=
DB_NAME=praktikum_backend
DB_SSLMODE=disable
DB_MAX_CONNS=10
```

### 3.2. Implementasi Inisialisasi Pool (`database/postgres.go`)
```go
package database

import (
	"context"
	"fmt"
	"time"

	"api-students/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		config.GetEnv("DB_USER", "postgres"),
		config.GetEnv("DB_PASSWORD", ""),
		config.GetEnv("DB_HOST", "localhost"),
		config.GetEnv("DB_PORT", "5432"),
		config.GetEnv("DB_NAME", "praktikum_backend"),
		config.GetEnv("DB_SSLMODE", "disable"),
	)

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("konfigurasi database tidak valid: %w", err)
	}

	cfg.MaxConns = int32(config.GetEnvInt("DB_MAX_CONNS", 10))
	cfg.MinConns = 2
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat pool: %w", err)
	}

	// Ping memastikan kredensial benar dan server memang dapat dihubungi saat startup
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("gagal terhubung ke database: %w", err)
	}

	return pool, nil
}
```

---

## BAGIAN 4: IMPLEMENTASI LAPISAN REPOSITORY DAN HANDLER

### 4.1. Kontrak Interface & Model Entitas
Didefinisikan pada `app/repository/student_repository.go`:
```go
type StudentRepository interface {
	FindAll(ctx context.Context, q model.ListStudentQuery) ([]model.Student, int, error)
	FindByID(ctx context.Context, id int) (model.Student, error)
	Create(ctx context.Context, s model.Student) (model.Student, error)
	Update(ctx context.Context, s model.Student) (model.Student, error)
	Delete(ctx context.Context, id int) error
}

var (
	ErrNotFound  = errors.New("data tidak ditemukan")
	ErrDuplicate = errors.New("data sudah ada")
)
```

### 4.2. Implementasi PostgreSQL Repository (`student_repository.go`)
Penyaringan, pengurutan, dan paginasi seluruhnya dikerjakan oleh PostgreSQL:

```go
func buildFilter(q model.ListStudentQuery) (string, []any) {
	where := " WHERE 1 = 1"
	args := []any{}

	if q.Search != "" {
		where += fmt.Sprintf(" AND (name ILIKE $%d OR nim ILIKE $%d)",
			len(args)+1, len(args)+1)
		args = append(args, "%"+q.Search+"%")
	}
	if q.IsActive != nil {
		where += fmt.Sprintf(" AND is_active = $%d", len(args)+1)
		args = append(args, *q.IsActive)
	}
	if q.MinGrade != nil {
		where += fmt.Sprintf(" AND grade >= $%d", len(args)+1)
		args = append(args, *q.MinGrade)
	}
	if q.MaxGrade != nil {
		where += fmt.Sprintf(" AND grade <= $%d", len(args)+1)
		args = append(args, *q.MaxGrade)
	}
	return where, args
}

func (r *studentPostgresRepository) FindAll(ctx context.Context, q model.ListStudentQuery) ([]model.Student, int, error) {
	where, args := buildFilter(q)

	// 1. Hitung total baris yang cocok sebelum dipaginasi
	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM students"+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("menghitung total mahasiswa: %w", err)
	}

	// 2. Proteksi SQL Injection pada klausa ORDER BY menggunakan whitelist kolom
	arah := "ASC"
	if q.Order == "desc" {
		arah = "DESC"
	}
	colName, ok := kolomUrut[q.Sort]
	if !ok {
		colName = "id"
	}

	// 3. Eksekusi query dengan LIMIT dan OFFSET
	sqlText := fmt.Sprintf(
		`SELECT id, nim, name, grade, is_active, created_at
		FROM students%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d`,
		where, colName, arah, len(args)+1, len(args)+2,
	)
	args = append(args, q.Limit, q.Offset())

	rows, err := r.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("mengambil daftar mahasiswa: %w", err)
	}
	defer rows.Close() // Wajib agar koneksi segera kembali ke connection pool

	hasil := []model.Student{}
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(&s.ID, &s.NIM, &s.Name, &s.Grade, &s.IsActive, &s.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("membaca baris mahasiswa: %w", err)
		}
		hasil = append(hasil, s)
	}
	return hasil, total, nil
}
```

### 4.3. Handler HTTP dan Dependency Injection (`handler.go`)
Handler menerima dependensi `StudentRepository` melalui konstruktor:
```go
type StudentHandler struct {
	repo repository.StudentRepository
}

func NewStudentHandler(repo repository.StudentRepository) *StudentHandler {
	return &StudentHandler{repo: repo}
}

// Pemetaan error repository ke HTTP status code
func terjemahkanError(c *fiber.Ctx, err error, pesanUmum string) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return fail(c, fiber.StatusConflict, "NIM sudah terdaftar pada sistem")
	default:
		return fail(c, fiber.StatusInternalServerError, pesanUmum)
	}
}
```

### 4.4. Perakitan Aplikasi Terpusat (`main.go`)
```go
func main() {
	config.LoadEnv()

	pool, err := database.NewPool(context.Background())
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	defer pool.Close()

	studentRepo := repository.NewStudentRepository(pool)
	studentHandler := NewStudentHandler(studentRepo)

	app := fiber.New(fiber.Config{ /* ErrorHandler */ })
	app.Use(requestid.New())
	app.Use(logger.New( /* logger config */ ))
	app.Use(cors.New())

	api := app.Group("/api/v1")
	api.Get("/health", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			return fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dihubungi")
		}
		return ok(c, "server dan basis data berjalan normal", fiber.Map{
			"status": "healthy", "database": "connected", "service": "api-students",
		})
	})

	s := api.Group("/students", requireJSON)
	s.Get("/", studentHandler.List)
	s.Get("/:id", studentHandler.Get)
	s.Post("/", studentHandler.Create)
	s.Put("/:id", studentHandler.Replace)
	s.Patch("/:id", studentHandler.Patch)
	s.Delete("/:id", studentHandler.Delete)

	log.Fatal(app.Listen(":" + config.GetEnv("APP_PORT", "3000")))
}
```

---

## BAGIAN 5: DOKUMEN KONTRAK API (API CONTRACT)

| Metode | Endpoint | Parameter / Query String | Contoh Body Permintaan | Status yang Mungkin Dikembalikan | Contoh Body Respons |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **GET** | `/api/v1/health` | *-* | *-* | • `200 OK`<br>• `503 Service Unavailable` | `{"success":true,"message":"server dan basis data berjalan normal","data":{"database":"connected","service":"api-students","status":"healthy"}}` |
| **GET** | `/api/v1/students` | **Query Parameter:**<br>• `page`, `limit` (max: 50)<br>• `search` (ILIKE)<br>• `sort` (whitelist), `order`<br>• `is_active`, `min_grade`, `max_grade` | *-* | • `200 OK`<br>• `500 Internal Server Error` | `{"success":true,"message":"daftar mahasiswa berhasil diambil","data":[{"id":1,"nim":"434241022","name":"Putri Apriliana","grade":95,"is_active":true,"created_at":"..."}],"meta":{"page":1,"limit":10,"total":1,"total_pages":1}}` |
| **GET** | `/api/v1/students/:id` | **Path Parameter:**<br>• `id` (int positif, wajib) | *-* | • `200 OK`<br>• `400 Bad Request`<br>• `404 Not Found`<br>• `500 Internal Server Error` | `{"success":true,"message":"mahasiswa ditemukan","data":{"id":1,"nim":"434241022","name":"Putri Apriliana","grade":95,"is_active":true,"created_at":"..."}}` |
| **POST** | `/api/v1/students` | **Header Wajib:**<br>`Content-Type: application/json` | `{"nim":"434241023","name":"Budi Santoso","grade":85.0,"is_active":true}` | • `201 Created` (+ Header `Location`)<br>• `400 Bad Request`<br>• `409 Conflict` (NIM ganda)<br>• `415 Unsupported Media Type`<br>• `422 Unprocessable Entity`<br>• `500 Internal Server Error` | `{"success":true,"message":"mahasiswa berhasil dibuat","data":{"id":2,"nim":"434241023","name":"Budi Santoso","grade":85,"is_active":true,"created_at":"..."}}` |
| **PUT** | `/api/v1/students/:id` | **Path Parameter:**<br>• `id` (int positif, wajib)<br>**Header Wajib:**<br>`Content-Type: application/json` | `{"nim":"434241022","name":"Putri Apriliana S.Tr.Kom","grade":98.0,"is_active":false}` | • `200 OK`<br>• `400 Bad Request`<br>• `404 Not Found`<br>• `409 Conflict`<br>• `415 Unsupported Media Type`<br>• `422 Unprocessable Entity`<br>• `500 Internal Server Error` | `{"success":true,"message":"mahasiswa berhasil diganti seluruhnya (PUT)","data":{"id":1,"nim":"434241022","name":"Putri Apriliana S.Tr.Kom","grade":98,"is_active":false,"created_at":"..."}}` |
| **PATCH** | `/api/v1/students/:id` | **Path Parameter:**<br>• `id` (int positif, wajib)<br>**Header Wajib:**<br>`Content-Type: application/json` | `{"grade":99.5,"is_active":true}` | • `200 OK`<br>• `400 Bad Request`<br>• `404 Not Found`<br>• `409 Conflict`<br>• `415 Unsupported Media Type`<br>• `422 Unprocessable Entity`<br>• `500 Internal Server Error` | `{"success":true,"message":"mahasiswa berhasil diperbarui sebagian (PATCH)","data":{"id":1,"nim":"434241022","name":"Putri Apriliana S.Tr.Kom","grade":99.5,"is_active":true,"created_at":"..."}}` |
| **DELETE** | `/api/v1/students/:id` | **Path Parameter:**<br>• `id` (int positif, wajib) | *-* | • `204 No Content` (tanpa body)<br>• `400 Bad Request`<br>• `404 Not Found`<br>• `500 Internal Server Error` | *(Tanpa Body / Kosong)* |

---

## BAGIAN 6: BUKTI PENGUJIAN ENDPOINT DAN KODE STATUS HTTP

Pengujian dilakukan menggunakan `curl -i` langsung ke server aktif yang terhubung ke PostgreSQL:

### 6.1. Status 200 OK — Health Check & Pengambilan Data
```bash
curl -i http://localhost:3000/api/v1/health
```
**Respons Terminal:**
```http
HTTP/1.1 200 OK
Date: Sat, 03 Oct 2026 12:21:55 GMT
Content-Type: application/json
Content-Length: 190
X-Request-Id: ea38ffbd-63a7-480a-988c-c90ceb725c6d

{"success":true,"message":"server dan basis data berjalan normal","data":{"database":"connected","service":"api-students","status":"healthy","timestamp":"2026-10-03T19:21:55.9519535+07:00"}}
```
> **Analisis:** Mengonfirmasi bahwa proses aplikasi web dan koneksi PostgreSQL pool berhasil melakukan ping secara sehat.

---

### 6.2. Status 201 Created — Penambahan Mahasiswa & Header Location
```bash
curl -i -X POST http://localhost:3000/api/v1/students \
  -H "Content-Type: application/json" \
  -d '{"nim":"434241022","name":"Putri Apriliana","grade":95.0,"is_active":true}'
```
**Respons Terminal:**
```http
HTTP/1.1 201 Created
Date: Sat, 03 Oct 2026 12:23:13 GMT
Content-Type: application/json
Content-Length: 189
X-Request-Id: ec38ffbd-63a7-480a-988c-c90ceb725c6d
Location: /api/v1/students/1

{"success":true,"message":"mahasiswa berhasil dibuat","data":{"id":1,"nim":"434241022","name":"Putri Apriliana","grade":95,"is_active":true,"created_at":"2026-10-03T19:23:14.230455+07:00"}}
```
> **Analisis:** Data berhasil disimpan ke tabel `students`. Field `id` (1) dan timestamp `created_at` digenerate langsung oleh mesin PostgreSQL melalui klausa `RETURNING`.

---

### 6.3. Status 204 No Content & 404 Not Found — Pengujian DELETE Dua Kali
**Percobaan Pertama (Menghapus Data yang Ada):**
```bash
curl -i -X DELETE http://localhost:3000/api/v1/students/2
```
**Respons Terminal:**
```http
HTTP/1.1 204 No Content
Date: Sat, 03 Oct 2026 12:33:50 GMT
X-Request-Id: f738ffbd-63a7-480a-988c-c90ceb725c6d
```

**Percobaan Kedua (Mengulang DELETE pada ID yang Sama):**
```bash
curl -i -X DELETE http://localhost:3000/api/v1/students/2
```
**Respons Terminal:**
```http
HTTP/1.1 404 Not Found
Date: Sat, 03 Oct 2026 12:34:56 GMT
Content-Type: application/json
Content-Length: 55
X-Request-Id: f838ffbd-63a7-480a-988c-c90ceb725c6d

{"success":false,"message":"mahasiswa tidak ditemukan"}
```
> **Analisis:** Pada percobaan kedua, perintah `DELETE` dieksekusi di database namun `tag.RowsAffected()` bernilai 0. Lapisan repository menerjemahkannya menjadi `ErrNotFound`, dan handler mengembalikan status `404 Not Found`.

---

### 6.4. Status 404 Not Found — Penanganan ID Tak Ditemukan (`pgx.ErrNoRows`)
```bash
curl -i http://localhost:3000/api/v1/students/999
```
**Respons Terminal:**
```http
HTTP/1.1 404 Not Found
Date: Sat, 03 Oct 2026 12:29:23 GMT
Content-Type: application/json
Content-Length: 55
X-Request-Id: f338ffbd-63a7-480a-988c-c90ceb725c6d

{"success":false,"message":"mahasiswa tidak ditemukan"}
```
> **Analisis:** Query `SELECT` tidak menghasilkan baris. Driver melempar `pgx.ErrNoRows`, yang dipetakan secara rapi menjadi `ErrNotFound` dan menghasilkan HTTP `404`.

---

### 6.5. Status 409 Conflict — Penegakan Keunikan NIM oleh Basis Data
Mencoba menambahkan data baru dengan NIM `434241022` yang sudah ada di tabel:
```bash
curl -i -X POST http://localhost:3000/api/v1/students \
  -H "Content-Type: application/json" \
  -d '{"nim":"434241022","name":"Putri Kembar","grade":88.0,"is_active":true}'
```
**Respons Terminal:**
```http
HTTP/1.1 409 Conflict
Date: Sat, 03 Oct 2026 12:24:08 GMT
Content-Type: application/json
Content-Length: 61
X-Request-Id: ef38ffbd-63a7-480a-988c-c90ceb725c6d

{"success":false,"message":"NIM sudah terdaftar pada sistem"}
```
> **Analisis:** `UNIQUE INDEX` pada `LOWER(nim)` di PostgreSQL menolak query `INSERT` dengan error code `23505`. Fungsi `isUniqueViolation` mendeteksi kode ini dan menghasilkan respons `409 Conflict`.

---

### 6.6. Pengujian Paginasi, Sorting, & Proteksi SQL Injection
**Pengujian 1: Paginasi dan Pengurutan di Sisi Basis Data**
```bash
curl -i "http://localhost:3000/api/v1/students?page=1&limit=2&sort=name&order=asc"
```
**Respons Terminal:**
```http
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 377

{"success":true,"message":"daftar mahasiswa berhasil diambil","data":[{"id":2,"nim":"434241023","name":"Budi Santoso","grade":85,"is_active":true,"created_at":"2026-10-03T19:23:41.443817+07:00"},{"id":3,"nim":"434241024","name":"Citra Lestari","grade":74,"is_active":false,"created_at":"2026-10-03T19:23:59.367971+07:00"}],"meta":{"page":1,"limit":2,"total":3,"total_pages":2}}
```

**Pengujian 2: Upaya SQL Injection pada Parameter Sort**
```bash
curl -i "http://localhost:3000/api/v1/students?sort=password;DROP%20TABLE%20students"
```
**Respons Terminal:**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{"success":true,"message":"daftar mahasiswa berhasil diambil","data":[ ... ],"meta":{"page":1,"limit":10,"total":3,"total_pages":1}}
```
> **Analisis:** String injection `password;DROP TABLE students` tidak terdapat pada map `kolomUrut` (whitelist). Repository secara otomatis menggantikannya dengan kolom default `id`, sehingga serangan SQL injection berhasil digagalkan total.

---

### 6.7. Pembuktian Persistensi Data (Server Restart Test)
1. **Server dimatikan:** Proses server dihentikan total (Ctrl+C).
2. **Server dijalankan kembali:** `go run .`
3. **Pengambilan data:**
```bash
curl -i http://localhost:3000/api/v1/students
```
**Respons Terminal:**
```http
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 392
X-Request-Id: 5ec3884c-932b-427a-900f-89ffd974704b

{"success":true,"message":"daftar mahasiswa berhasil diambil","data":[{"id":1,"nim":"434241022","name":"Putri Apriliana S.Tr.Kom","grade":99.5,"is_active":true,"created_at":"2026-10-03T19:23:14.230455+07:00"},{"id":3,"nim":"434241024","name":"Citra Lestari","grade":74,"is_active":false,"created_at":"2026-10-03T19:23:59.367971+07:00"}],"meta":{"page":1,"limit":10,"total":2,"total_pages":1}}
```
> **Analisis:** Seluruh data mahasiswa beserta perubahan hasil operasi PUT dan PATCH tetap tersimpan utuh di PostgreSQL dan tidak hilang saat proses aplikasi dimatikan.

---

### 6.8. Analisis Status 503 vs 500 saat Basis Data Terputus
Ketika layanan PostgreSQL dimatikan sementara:
1. **Pemanggilan Endpoint `/api/v1/health`:**
   Mengembalikan **`503 Service Unavailable`** (`{"success":false,"message":"database tidak dapat dihubungi"}`).
2. **Pemanggilan Endpoint Bisnis `/api/v1/students`:**
   Mengembalikan **`500 Internal Server Error`** (`{"success":false,"message":"gagal mengambil data mahasiswa"}`).

**Penjelasan Status Mana yang Paling Tepat:**
- **Untuk endpoint `/health`:** Status **`503 Service Unavailable`** adalah yang paling tepat. Kode 503 secara semantik memberitahukan load balancer, API gateway, atau Kubernetes probe bahwa pod/layanan sedang tidak sehat (*unhealthy*) akibat dependensi hulu yang mati sementara, sehingga lalu lintas trafik dapat dialihkan ke instans lain.
- **Untuk endpoint operasional (`/students`):** Status **`500 Internal Server Error`** tepat digunakan karena server aplikasi gagal memproses kontrak permintaan akibat kegagalan internal yang tidak terduga pada layer dependensi infrastruktur.

---

## BAGIAN 7: ANALISIS & JAWABAN PERTANYAAN EKSPLORASI

### 1. Eksplorasi Connection Pool & Max Connections (Halaman 2)
- **Berapa nilai `max_connections` pada PostgreSQL Anda? Apa yang terjadi bila lima aplikasi masing-masing membuka pool berisi 50 koneksi?**
  Nilai default `max_connections` pada PostgreSQL lokal umumnya adalah `100`. Bila lima aplikasi masing-masing membuka pool dengan kapasitas 50 koneksi, total kebutuhan koneksi adalah 250 koneksi ($5 \times 50$). Ketika batas 100 terlampaui, PostgreSQL akan menolak koneksi baru dengan pesan error:
  `FATAL: sorry, too many clients already`.
  Aplikasi ke-3 dan seterusnya akan mengalami *connection refused* dan gagal startup.

### 2. Eksplorasi Driver: pgx vs database/sql (Halaman 2)
- **Perbedaan `pgx` native dengan `pgx` melalui `database/sql`:**
  - **`database/sql` standar Go:** Menyediakan antarmuka generik yang kompatibel antar berbagai basis data (MySQL, Postgres, SQLite). Namun, fitur-fitur spesifik PostgreSQL hilang (seperti tipe data native array/JSONB, `Listen/Notify`, operasi *bulk copy* binary `CopyFrom`, dan performa encoding yang lebih cepat).
  - **`pgx` native (v5):** Menawarkan performa jauh lebih tinggi, alokasi memori minimal, dukungan penuh terhadap tipe PostgreSQL khusus, serta pooling koneksi yang sangat efisien (`pgxpool`).

### 3. Eksplorasi Indeks & Tipe Waktu (Halaman 4)
- **Mengapa indeks unik dibuat pada `LOWER(username)` / `LOWER(nim)`?**
  Jika indeks hanya dibuat pada kolom aslinya tanpa `LOWER()`, PostgreSQL membedakan huruf besar dan kecil (*case-sensitive*). Nilai `"Andi"` dan `"andi"` dianggap dua nilai yang berbeda dan keduanya akan tersimpan. Dengan ekspresi `LOWER()`, kedua masukan akan dievaluasi ke nilai huruf kecil yang identik, sehingga duplikasi beda kapitalisasi berhasil dicegah.
- **Perbedaan `TIMESTAMPTZ` vs `TIMESTAMP`:**
  `TIMESTAMP` menyimpan tanggal dan waktu polos tanpa informasi zona waktu (*timezone unaware*). `TIMESTAMPTZ` (*timestamp with time zone*) mengonversi waktu masukan ke waktu UTC secara absolut sebelum disimpan ke disk, dan mengonversinya kembali ke zona waktu sesi klien saat dibaca. `TIMESTAMPTZ` wajib digunakan untuk kolom audit (`created_at`, `updated_at`) guna menghindari ambiguitas waktu saat server dan klien berada di zona waktu yang berbeda.
- **Apa fungsi perkakas migrasi seperti `golang-migrate` atau `goose`?**
  Perkakas migrasi mencatat nomor versi skema yang telah dijalankan ke dalam tabel khusus (`schema_migrations`), mendukung skrip otomatis *up* (penerapan skema) dan *down* (rollback skema), serta menjamin konsistensi skema basis data di seluruh mesin pengembang dan server *continuous delivery* (CI/CD).

### 4. Eksplorasi Kode Error Basis Data (Halaman 5)
- **Kode Error `23503` dan `23502`:**
  - **`23503 (foreign_key_violation)`:** Terjadi saat operasi mengacu pada ID relasi yang tidak ada di tabel induk. Status HTTP yang pantas: **`409 Conflict`** atau **`422 Unprocessable Entity`**.
  - **`23502 (not_null_violation)`:** Terjadi saat kolom yang memiliki batasan `NOT NULL` diisi nilai `NULL`. Status HTTP yang pantas: **`422 Unprocessable Entity`**.
- **Bahaya membocorkan pesan error mentah database ke klien:**
  Pesan seperti `pq: relation "users" does not exist` atau detail nama kolom membocorkan arsitektur skema internal kepada penyerang (*reconnaissance phase*), memudahkan penyerang merancang muatan SQL injection yang tepat sasaran.

### 5. Eksplorasi Bahaya Melupakan `defer rows.Close()` (Halaman 18)
- Bila `defer rows.Close()` dihilangkan, koneksi yang sedang dipakai membaca kumpulan baris data tidak dikembalikan ke pool. Setelah beberapa kali pemanggilan endpoint daftar, seluruh slot koneksi di `pgxpool` (misal 10 koneksi) habis terpakai dan terkunci (*connection leak*). Akibatnya, request berikutnya akan mengalami *hang* tanpa henti (*starvation*) hingga mencapai batas waktu context.

---

## BAGIAN 8: RIWAYAT COMMIT GIT BERTAHAP

Penyelesaian penugasan Modul 3 tercatat rapi melalui commit-commit bertahap pada repositori Git:

```text
54d7c3e feat(modul-3): konfigurasi dependency injection pada main.go dan dokumentasi kontrak API di README.md
6c6b53d feat(modul-3): perbarui handler dan helper untuk menggunakan StudentRepository dan context timeout
0a3b920 feat(modul-3): implementasi StudentRepository interface dan postgres repository dengan query berparameter
2419d06 feat(modul-3): implementasi konfigurasi environment dan connection pool pgxpool
6ac6f1f feat(modul-3): buat skema database dan migrasi tabel students dengan indeks unik NIM
7cce444 feat(latihan-fiber): implementasi postgresql connection pool dan repository pattern sesuai langkah panduan modul 3
```

---

## BAGIAN 9: PERNYATAAN PENGGUNAAN SUMBER BANTUAN & AI

Sesuai ketentuan pengumpulan Modul 3 poin 7:
1. **Alat Bantu AI:**
   - **Google Antigravity AI Assistant** (berbasis Gemini 3.8 Flash).
2. **Rincian Bantuan yang Diberikan:**
   - Membedah modul praktikum PDF dan merancang arsitektur implementasi *Repository Pattern*.
   - Menyusun skrip migrasi SQL (`001_create_students.sql`) lengkap dengan indeks unik ekspresi `LOWER(nim)`.
   - Mengonfigurasi connection pool `pgxpool` dengan parameter `MaxConns`, `MinConns`, dan validasi startup via `Ping`.
   - Menulis logika repository berparameter untuk pencarian `ILIKE`, penyaringan status/nilai, dan whitelist pengurutan.
   - Menjalankan pengujian terminal dan memverifikasi kode status HTTP (`200`, `201`, `204`, `404`, `409`, `500`, `503`).
3. **Pernyataan Kemandirian:**
   Mahasiswa memahami sepenuhnya setiap baris kode yang dihasilkan, mampu menjelaskan perbedaan arsitektural antara penyimpanan memori dan basis data relasional, mekanisme penanganan error `pgconn.PgError`, serta siap mempertanggungjawabkannya pada sesi praktikum dan responsi.
