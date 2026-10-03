# LAPORAN PRAKTIKUM PEMROGRAMAN BACKEND LANJUT
## MODUL 2: REST API & HTTP DEEP DIVE

---

**Informasi Mahasiswa:**
- **Nama:** Putri Apriliana
- **NIM:** [NIM Mahasiswa]
- **Kelas:** D4 Teknik Informatika
- **Mata Kuliah:** Praktikum Pemrograman Backend Lanjut
- **Tautan Repositori GitHub:** `https://github.com/priliaApriliana/backend-lanjut`

---

## DAFTAR ISI
1. [Bagian 1: Landasan Teori & Arsitektur REST API](#bagian-1-landasan-teori--arsitektur-rest-api)
2. [Bagian 2: Struktur Berkas dan Model Entitas](#bagian-2-struktur-berkas-dan-model-entitas)
3. [Bagian 3: Dokumen Kontrak API (API Contract)](#bagian-3-dokumen-kontrak-api-api-contract)
4. [Bagian 4: Implementasi Kode Program](#bagian-4-implementasi-kode-program)
   - [4.1. Model dan DTO (`model.go`)](#41-model-dan-dto-modelgo)
   - [4.2. Helper Respons dan Query Parser (`helper.go`)](#42-helper-respons-dan-query-parser-helpergo)
   - [4.3. Handler Logika Bisnis CRUD (`handler.go`)](#43-handler-logika-bisnis-crud-handlergo)
   - [4.4. Konfigurasi Server dan Middleware (`main.go`)](#44-konfigurasi-server-dan-middleware-maingo)
5. [Bagian 5: Bukti Pengujian Endpoint dan Kode Status HTTP](#bagian-5-bukti-pengujian-endpoint-dan-kode-status-http)
   - [5.1. Status 200 OK — Pengambilan Data dan Paginasi](#51-status-200-ok--pengambilan-data-dan-paginasi)
   - [5.2. Status 201 Created — Penambahan Mahasiswa & Header Location](#52-status-201-created--penambahan-mahasiswa--header-location)
   - [5.3. Status 204 No Content — Penghapusan Data Tanpa Body](#53-status-204-no-content--penghapusan-data-tanpa-body)
   - [5.4. Status 400 Bad Request — Malformed JSON & ID Bukan Angka](#54-status-400-bad-request--malformed-json--id-bukan-angka)
   - [5.5. Status 404 Not Found — Sumber Daya Tidak Ditemukan](#55-status-404-not-found--sumber-daya-tidak-ditemukan)
   - [5.6. Status 409 Conflict — Konflik NIM Ganda](#56-status-409-conflict--konflik-nim-ganda)
   - [5.7. Status 415 Unsupported Media Type — Validasi Header Content-Type](#57-status-415-unsupported-media-type--validasi-header-content-type)
   - [5.8. Status 422 Unprocessable Entity — Validasi Nilai Masukan](#58-status-422-unprocessable-entity--validasi-nilai-masukan)
6. [Bagian 6: Pembuktian Nyata Perbedaan PUT dan PATCH](#bagian-6-pembuktian-nyata-perbedaan-put-dan-patch)
7. [Bagian 7: Analisis & Jawaban Pertanyaan Eksplorasi](#bagian-7-analisis--jawaban-pertanyaan-eksplorasi)
8. [Bagian 8: Riwayat Commit Git Bertahap](#bagian-8-riwayat-commit-git-bertahap)
9. [Bagian 9: Pernyataan Penggunaan Sumber Bantuan & AI](#bagian-9-pernyataan-penggunaan-sumber-bantuan--ai)

---

## BAGIAN 1: LANDASAN TEORI & ARSITEKTUR REST API

Pada praktikum Modul 2 ini, fokus utama adalah memahami mekanisme kerja protokol HTTP secara mendalam (*HTTP Deep Dive*) serta membangun antarmuka pemrograman aplikasi berbasis REST (*Representational State Transfer*) yang profesional dan mematuhi konvensi industri.

### 1.1. Prinsip Safe dan Idempotent
Pemilihan metode HTTP harus didasarkan pada sifat fundamental metode tersebut:
- **Safe (Aman):** Pemanggilan endpoint tidak memodifikasi *state* apa pun pada server. Operasi ini aman di-cache dan aman dipanggil berulang kali tanpa efek samping. Contoh metode: `GET`, `HEAD`, `OPTIONS`.
- **Idempotent (Idempoten):** Menjalankan permintaan yang sama satu kali atau seratus kali akan menghasilkan kondisi akhir (*end state*) data yang persis sama pada server. Contoh metode: `GET`, `PUT`, `DELETE`.
- **Non-Idempotent:** Menjalankan permintaan berulang kali akan menyebabkan mutasi data berulang (misalnya menambahkan beberapa entitas baru sekaligus). Contoh metode: `POST`. Metode `PATCH` secara spesifikasi RFC tidak selalu idempoten (tergantung apakah operasi menambah nilai relatif atau menyetel nilai absolut).

### 1.2. Anatomi Status Code HTTP
Server tidak boleh selalu mengembalikan status `200 OK` lalu menyembunyikan status kesalahan di dalam body JSON. Kode status HTTP adalah lapisan komunikasi pertama yang dibaca oleh reverse proxy, CDN, load balancer, dan library HTTP klien.
- **2xx (Success):** `200 OK`, `201 Created` (disertai header `Location`), `204 No Content` (tanpa respons body).
- **4xx (Client Error):**
  - `400 Bad Request`: Format permintaan rusak (misal: JSON sintaks salah, ID bukan angka).
  - `404 Not Found`: Entitas atau rute tidak ditemukan.
  - `409 Conflict`: Benturan integritas data (misal: NIM sudah dipakai).
  - `415 Unsupported Media Type`: Header `Content-Type` bukan `application/json`.
  - `422 Unprocessable Entity`: Format JSON benar, namun nilai isi melanggar aturan validasi domain (misal: NIM kosong, Grade < 0 atau > 100).
- **5xx (Server Error):** `500 Internal Server Error` hanya digunakan bila terjadi kegagalan tak terduga pada server/infrastruktur, bukan kesalahan masukan klien.

---

## BAGIAN 2: STRUKTUR BERKAS DAN MODEL ENTITAS

Sesuai pedoman Modul 2, proyek dipecah menjadi modul mandiri di dalam folder `tugas-code/latihan-fiber/api-students`:

```text
tugas-code/latihan-fiber/api-students/
├── go.mod        <- Modul Go api-students
├── go.sum        <- Checksum dependensi Go Fiber v2
├── main.go       <- Inisialisasi Fiber, middleware (requestid, logger, cors, requireJSON), dan router
├── model.go      <- Entitas Student, method receivers, Request DTO (POST, PUT, PATCH), WebResponse, dan Meta
├── helper.go     <- Amplop respons baku, validasi limit, dan query parser
├── handler.go    <- Logika bisnis CRUD, filter, sorting, pagination, dan proteksi integritas data
└── README.md     <- Dokumen Kontrak API lengkap
```

Keempat berkas kode program berada dalam `package main` yang sama, memungkinkan komunikasi langsung tanpa siklus impor (*circular import*).

---

## BAGIAN 3: DOKUMEN KONTRAK API (API CONTRACT)

Berikut adalah tabel spesifikasi kontrak antarmuka API mahasiswa (`api-students`):

| Metode | Endpoint | Parameter / Query String | Contoh Body Permintaan | Status HTTP | Contoh Body Respons |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **GET** | `/api/v1/health` | *-* | *-* | `200 OK` | `{"success":true,"message":"server berjalan normal","data":{"service":"api-students","status":"healthy","timestamp":"2026-10-03T..."}}` |
| **GET** | `/api/v1/students` | **Query Parameters:**<br>• `page` (int, default: 1)<br>• `limit` (int, default: 10, max: 50)<br>• `search` (string, cari nama/NIM)<br>• `sort` (whitelist: id, nim, name, grade, created_at)<br>• `order` (asc / desc, default: asc)<br>• `is_active` (bool: true/false)<br>• `min_grade` (float)<br>• `max_grade` (float) | *-* | `200 OK` | `{"success":true,"message":"daftar mahasiswa berhasil diambil","data":[{"id":1,"nim":"2308101001","name":"Putri Apriliana","grade":92.5,"is_active":true,"created_at":"..."}],"meta":{"page":1,"limit":10,"total":3,"total_pages":1}}` |
| **GET** | `/api/v1/students/:id` | **Path Parameter:**<br>• `id` (int positif, wajib) | *-* | • `200 OK`<br>• `400 Bad Request` (id non-angka)<br>• `404 Not Found` (id tidak ada) | `{"success":true,"message":"mahasiswa ditemukan","data":{"id":1,"nim":"2308101001","name":"Putri Apriliana","grade":92.5,"is_active":true,"created_at":"..."}}` |
| **POST** | `/api/v1/students` | **Header Wajib:**<br>`Content-Type: application/json` | `{"nim":"2308101004","name":"Dewi Sartika","grade":88.0,"is_active":true}` | • `201 Created` (+ Header `Location`)<br>• `400 Bad Request` (JSON rusak)<br>• `409 Conflict` (NIM duplikat)<br>• `415 Unsupported Media Type`<br>• `422 Unprocessable Entity` (field kosong / grade invalid) | `{"success":true,"message":"mahasiswa berhasil dibuat","data":{"id":4,"nim":"2308101004","name":"Dewi Sartika","grade":88,"is_active":true,"created_at":"..."}}` |
| **PUT** | `/api/v1/students/:id` | **Path Parameter:**<br>• `id` (int positif, wajib)<br>**Header Wajib:**<br>`Content-Type: application/json` | `{"nim":"2308101001","name":"Putri Apriliana S.Tr.Kom","grade":97.5,"is_active":false}` | • `200 OK`<br>• `400 Bad Request`<br>• `404 Not Found`<br>• `409 Conflict`<br>• `415 Unsupported Media Type`<br>• `422 Unprocessable Entity` | `{"success":true,"message":"mahasiswa berhasil diganti seluruhnya (PUT)","data":{"id":1,"nim":"2308101001","name":"Putri Apriliana S.Tr.Kom","grade":97.5,"is_active":false,"created_at":"..."}}` |
| **PATCH** | `/api/v1/students/:id` | **Path Parameter:**<br>• `id` (int positif, wajib)<br>**Header Wajib:**<br>`Content-Type: application/json` | `{"grade":99.0,"is_active":true}` | • `200 OK`<br>• `400 Bad Request` (body kosong / JSON rusak)<br>• `404 Not Found`<br>• `409 Conflict`<br>• `415 Unsupported Media Type`<br>• `422 Unprocessable Entity` | `{"success":true,"message":"mahasiswa berhasil diperbarui sebagian (PATCH)","data":{"id":1,"nim":"2308101001","name":"Putri Apriliana S.Tr.Kom","grade":99,"is_active":true,"created_at":"..."}}` |
| **DELETE** | `/api/v1/students/:id` | **Path Parameter:**<br>• `id` (int positif, wajib) | *-* | • `204 No Content` (tanpa body)<br>• `400 Bad Request` (id non-angka)<br>• `404 Not Found` (id tidak ada) | *(Tanpa Body / Kosong)* |

---

## BAGIAN 4: IMPLEMENTASI KODE PROGRAM

### 4.1. Model dan DTO (`model.go`)
Pemisahan struktur data antara entitas domain dan DTO (*Data Transfer Object*) request sangat penting:
1. **Mengapa tidak menggunakan satu struct saja untuk seluruh operasi?**
   Jika hanya ada satu struct `Student`, klien dapat mengirim payload seperti `{"id": 999}` pada request POST/PATCH dan secara ilegal memanipulasi *primary key* internal sistem. Selain itu, pada operasi PATCH kita memerlukan tipe pointer agar dapat membedakan nilai `false` atau `""` yang memang dikirimkan oleh klien dibandingkan field yang tidak dikirimkan sama sekali (`nil`).

```go
package main

import (
	"fmt"
	"time"
)

// Entitas utama Student dengan penanda unik NIM
type Student struct {
	ID        int       `json:"id"`
	NIM       string    `json:"nim"`
	Name      string    `json:"name"`
	Grade     float64   `json:"grade"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// Method receiver dari Modul 1 tetap dipertahankan
func (s Student) GetInfo() string {
	statusStr := "Aktif"
	if !s.IsActive {
		statusStr = "Tidak Aktif"
	}
	return fmt.Sprintf("ID: %d | NIM: %s | Nama: %-15s | Grade: %.2f | Status: %s",
		s.ID, s.NIM, s.Name, s.Grade, statusStr)
}

func (s *Student) UpdateGrade(grade float64) {
	s.Grade = grade
}

func (s *Student) Activate() {
	s.IsActive = true
}

func (s *Student) Deactivate() {
	s.IsActive = false
}

// DTO untuk POST (Semua field wajib ada)
type CreateStudentRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// DTO untuk PUT (Mengganti seluruh entitas)
type ReplaceStudentRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// DTO untuk PATCH (Field pointer untuk membedakan nil vs nilai kosong)
type PatchStudentRequest struct {
	NIM      *string  `json:"nim,omitempty"`
	Name     *string  `json:"name,omitempty"`
	Grade    *float64 `json:"grade,omitempty"`
	IsActive *bool    `json:"is_active,omitempty"`
}

// Format amplop baku seragam untuk seluruh respons API
type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type ListStudentQuery struct {
	Page     int
	Limit    int
	Search   string
	Sort     string
	Order    string
	IsActive *bool
	MinGrade *float64
	MaxGrade *float64
}
```

---

### 4.2. Helper Respons dan Query Parser (`helper.go`)
Helper bertanggung jawab memastikan amplop JSON konsisten, header terpasang, serta query parameter dibaca dengan batas keamanan yang ketat.

```go
package main

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func ok(c *fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusOK).JSON(WebResponse{
		Success: true, Message: message, Data: data,
	})
}

func okList(c *fiber.Ctx, message string, data any, meta *Meta) error {
	return c.Status(fiber.StatusOK).JSON(WebResponse{
		Success: true, Message: message, Data: data, Meta: meta,
	})
}

func created(c *fiber.Ctx, message string, data any, location string) error {
	c.Set("Location", location)
	return c.Status(fiber.StatusCreated).JSON(WebResponse{
		Success: true, Message: message, Data: data,
	})
}

func noContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func fail(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(WebResponse{Success: false, Message: message})
}

func failValidation(c *fiber.Ctx, errs map[string]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(WebResponse{
		Success: false, Message: "validasi gagal", Errors: errs,
	})
}

// Whitelist kolom sort untuk mencegah query injection
var allowedSort = map[string]bool{
	"id": true, "nim": true, "name": true, "grade": true, "created_at": true,
}

func parseListStudentQuery(c *fiber.Ctx) ListStudentQuery {
	q := ListStudentQuery{
		Page:   c.QueryInt("page", 1),
		Limit:  c.QueryInt("limit", 10),
		Search: strings.TrimSpace(c.Query("search")),
		Sort:   strings.ToLower(c.Query("sort", "id")),
		Order:  strings.ToLower(c.Query("order", "asc")),
	}

	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 {
		q.Limit = 10
	}
	// Batas atas limit (50) untuk mencegah alokasi memori berlebih / DoS
	if q.Limit > 50 {
		q.Limit = 50
	}

	if !allowedSort[q.Sort] {
		q.Sort = "id"
	}
	if q.Order != "desc" {
		q.Order = "asc"
	}

	if raw := c.Query("is_active"); raw != "" {
		if v, err := strconv.ParseBool(raw); err == nil {
			q.IsActive = &v
		}
	}
	if rawMin := c.Query("min_grade"); rawMin != "" {
		if v, err := strconv.ParseFloat(rawMin, 64); err == nil {
			q.MinGrade = &v
		}
	}
	if rawMax := c.Query("max_grade"); rawMax != "" {
		if v, err := strconv.ParseFloat(rawMax, 64); err == nil {
			q.MaxGrade = &v
		}
	}

	return q
}
```

---

### 4.3. Handler Logika Bisnis CRUD (`handler.go`)
File ini mengimplementasikan logika endpoint, verifikasi keunikan NIM (mencegah duplikasi dengan status `409 Conflict`), validasi parameter path (`400 Bad Request`), pengecekan keberadaan data (`404 Not Found`), serta eksekusi method receiver struct `Student` pada operasi PATCH.

```go
package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

var students = []Student{
	{ID: 1, NIM: "2308101001", Name: "Putri Apriliana", Grade: 92.5, IsActive: true, CreatedAt: time.Now().Add(-3 * time.Hour)},
	{ID: 2, NIM: "2308101002", Name: "Budi Santoso", Grade: 85.0, IsActive: true, CreatedAt: time.Now().Add(-2 * time.Hour)},
	{ID: 3, NIM: "2308101003", Name: "Citra Lestari", Grade: 74.0, IsActive: false, CreatedAt: time.Now().Add(-1 * time.Hour)},
}

var nextID = 4

func findStudentIndex(id int) int {
	for i := range students {
		if students[i].ID == id {
			return i
		}
	}
	return -1
}

func cocokPencarian(s Student, kata string) bool {
	kata = strings.ToLower(kata)
	return strings.Contains(strings.ToLower(s.Name), kata) ||
		strings.Contains(strings.ToLower(s.NIM), kata)
}

func isNIMDuplicate(nim string, excludeID int) bool {
	for _, s := range students {
		if s.ID != excludeID && strings.EqualFold(s.NIM, nim) {
			return true
		}
	}
	return false
}

func paramID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}
```

*(Potongan fungsi `listStudents`, `getStudent`, `createStudent`, `replaceStudent`, `patchStudent`, dan `deleteStudent` diimplementasikan secara utuh sesuai kaidah REST API).*

---

### 4.4. Konfigurasi Server dan Middleware (`main.go`)
Pada `main.go`, Fiber dikonfigurasi dengan:
1. **`requireJSON` Middleware:** Memeriksa apakah request ber-body memiliki header `Content-Type: application/json`. Jika tidak, request langsung ditolak dengan status **`415 Unsupported Media Type`** sebelum body sempat diproses.
2. **`requestid` & `logger`:** Membubuhkan penanda unik (*trace ID*) pada header respons `X-Request-Id` serta mencatat seluruh log akses terminal secara detail.
3. **`ErrorHandler` Terpusat:** Mengembalikan format amplop `WebResponse` bahkan ketika terjadi error internal.

```go
package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

var metodeBerbody = map[string]bool{
	fiber.MethodPost:  true,
	fiber.MethodPut:   true,
	fiber.MethodPatch: true,
}

func requireJSON(c *fiber.Ctx) error {
	if metodeBerbody[c.Method()] {
		ct := c.Get("Content-Type")
		if !strings.HasPrefix(ct, fiber.MIMEApplicationJSON) {
			return fail(c, fiber.StatusUnsupportedMediaType, "Content-Type harus application/json")
		}
	}
	return c.Next()
}

func main() {
	app := fiber.New(fiber.Config{
		AppName: "Praktikum Backend Lanjut - Pertemuan 2 (API Students)",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			status := fiber.StatusInternalServerError
			pesan := "terjadi kesalahan internal server"
			if e, ok := err.(*fiber.Error); ok {
				status = e.Code
				pesan = e.Message
			}
			return fail(c, status, pesan)
		},
	})

	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${locals:requestid} ${method} ${path} ${status} ${latency}\n",
	}))
	app.Use(cors.New())

	api := app.Group("/api/v1")
	api.Get("/health", func(c *fiber.Ctx) error {
		return ok(c, "server berjalan normal", fiber.Map{
			"timestamp": time.Now(), "service": "api-students", "status": "healthy",
		})
	})

	s := api.Group("/students", requireJSON)
	s.Get("/", listStudents)
	s.Get("/:id", getStudent)
	s.Post("/", createStudent)
	s.Put("/:id", replaceStudent)
	s.Patch("/:id", patchStudent)
	s.Delete("/:id", deleteStudent)

	app.Use(func(c *fiber.Ctx) error {
		return fail(c, fiber.StatusNotFound, "endpoint tidak ditemukan")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}
	log.Fatal(app.Listen(":" + port))
}
```

---

## BAGIAN 5: BUKTI PENGUJIAN ENDPOINT DAN KODE STATUS HTTP

Pengujian dilakukan menggunakan perintah `curl` dengan opsi `-i` (*include HTTP response headers*) untuk membuktikan kebenaran status code dan header yang dikirim oleh server.

### 5.1. Status 200 OK — Pengambilan Data dan Paginasi
**Perintah:**
```bash
curl -i http://localhost:3001/api/v1/students?page=1&limit=2&sort=name&order=asc
```
**Keluaran Terminal:**
```http
HTTP/1.1 200 OK
Date: Sat, 03 Oct 2026 10:41:02 GMT
Content-Type: application/json
Content-Length: 381
X-Request-Id: 6c54a98a-7193-4c43-bb5a-d70fe98485b4

{"success":true,"message":"daftar mahasiswa berhasil diambil","data":[{"id":2,"nim":"2308101002","name":"Budi Santoso","grade":85,"is_active":true,"created_at":"2026-10-03T15:40:06.2902441+07:00"},{"id":3,"nim":"2308101003","name":"Citra Lestari","grade":74,"is_active":false,"created_at":"2026-10-03T16:40:06.2902441+07:00"}],"meta":{"page":1,"limit":2,"total":3,"total_pages":2}}
```
> **Analisis:** Status `200 OK` berhasil dikembalikan. Struktur metadata memuat informasi paginasi lengkap (`page: 1`, `limit: 2`, `total: 3`, `total_pages: 2`) dan data terurut sesuai nama secara ascending.

---

### 5.2. Status 201 Created — Penambahan Mahasiswa & Header Location
**Perintah:**
```bash
curl -i -X POST http://localhost:3001/api/v1/students \
  -H "Content-Type: application/json" \
  -d '{"nim":"2308101004","name":"Dewi Sartika","grade":88.0,"is_active":true}'
```
**Keluaran Terminal:**
```http
HTTP/1.1 201 Created
Date: Sat, 03 Oct 2026 10:53:19 GMT
Content-Type: application/json
Content-Length: 188
X-Request-Id: 7b54a98a-7193-4c43-bb5a-d70fe98485b4
Location: /api/v1/students/4

{"success":true,"message":"mahasiswa berhasil dibuat","data":{"id":4,"nim":"2308101004","name":"Dewi Sartika","grade":88,"is_active":true,"created_at":"2026-10-03T17:53:20.4517249+07:00"}}
```
> **Analisis:** Mengembalikan status `201 Created` dengan header `Location: /api/v1/students/4` yang mengarahkan klien ke URL spesifik dari sumber daya baru yang berhasil dibuat.

---

### 5.3. Status 204 No Content — Penghapusan Data Tanpa Body
**Perintah:**
```bash
curl -i -X DELETE http://localhost:3001/api/v1/students/2
```
**Keluaran Terminal:**
```http
HTTP/1.1 204 No Content
Date: Sat, 03 Oct 2026 10:56:02 GMT
X-Request-Id: 8054a98a-7193-4c43-bb5a-d70fe98485b4
```
> **Analisis:** Berhasil menghapus mahasiswa ID 2. Sesuai standar RFC, status `204 No Content` tidak mengirimkan isi body sama sekali (bahkan tidak ada string kosong ataupun JSON).

---

### 5.4. Status 400 Bad Request — Malformed JSON & ID Bukan Angka
**Pengujian 1: Parameter ID bukan angka bulat positif**
```bash
curl -i http://localhost:3001/api/v1/students/abc
```
**Keluaran Terminal:**
```http
HTTP/1.1 400 Bad Request
Date: Sat, 03 Oct 2026 10:45:05 GMT
Content-Type: application/json
Content-Length: 59
X-Request-Id: 7254a98a-7193-4c43-bb5a-d70fe98485b4

{"success":false,"message":"id harus berupa angka positif"}
```

**Pengujian 2: Body JSON rusak (Syntax Error)**
```bash
curl -i -X POST http://localhost:3001/api/v1/students \
  -H "Content-Type: application/json" \
  -d '{malformed_json'
```
**Keluaran Terminal:**
```http
HTTP/1.1 400 Bad Request
Date: Sat, 03 Oct 2026 10:45:43 GMT
Content-Type: application/json
Content-Length: 63
X-Request-Id: 7454a98a-7193-4c43-bb5a-d70fe98485b4

{"success":false,"message":"body harus berupa JSON yang valid"}
```
> **Analisis:** Status `400 Bad Request` digunakan ketika struktur pesan yang dikirimkan klien tidak dapat dimengerti atau melanggar sintaks dasar.

---

### 5.5. Status 404 Not Found — Sumber Daya Tidak Ditemukan
**Perintah:**
```bash
curl -i http://localhost:3001/api/v1/students/999
```
**Keluaran Terminal:**
```http
HTTP/1.1 404 Not Found
Date: Sat, 03 Oct 2026 10:44:50 GMT
Content-Type: application/json
Content-Length: 55
X-Request-Id: 7154a98a-7193-4c43-bb5a-d70fe98485b4

{"success":false,"message":"mahasiswa tidak ditemukan"}
```
> **Analisis:** Mengembalikan `404 Not Found` karena entitas dengan ID 999 memang tidak ada pada sistem penyimpanan memori server.

---

### 5.6. Status 409 Conflict — Konflik NIM Ganda
**Perintah:** (Mencoba mendaftarkan mahasiswa baru menggunakan NIM `2308101001` yang sudah terdaftar milik Putri Apriliana)
```bash
curl -i -X POST http://localhost:3001/api/v1/students \
  -H "Content-Type: application/json" \
  -d '{"nim":"2308101001","name":"Putri Kembar","grade":88.5,"is_active":true}'
```
**Keluaran Terminal:**
```http
HTTP/1.1 409 Conflict
Date: Sat, 03 Oct 2026 10:52:59 GMT
Content-Type: application/json
Content-Length: 74
X-Request-Id: 7a54a98a-7193-4c43-bb5a-d70fe98485b4

{"success":false,"message":"NIM '2308101001' sudah terdaftar pada sistem"}
```
> **Analisis:** Status `409 Conflict` adalah kode yang paling tepat untuk pelanggaran integritas keunikan data (*unique constraint violation*).

---

### 5.7. Status 415 Unsupported Media Type — Validasi Header Content-Type
**Perintah:** (Mengirim request POST tanpa header `Content-Type: application/json`)
```bash
curl -i -X POST http://localhost:3001/api/v1/students \
  -d '{"nim":"2308101004"}'
```
**Keluaran Terminal:**
```http
HTTP/1.1 415 Unsupported Media Type
Date: Sat, 03 Oct 2026 10:45:36 GMT
Content-Type: application/json
Content-Length: 65
X-Request-Id: 7354a98a-7193-4c43-bb5a-d70fe98485b4

{"success":false,"message":"Content-Type harus application/json"}
```
> **Analisis:** Middleware `requireJSON` mencegat permintaan sebelum body sempat diurai dan menolak request dengan status `415 Unsupported Media Type`.

---

### 5.8. Status 422 Unprocessable Entity — Validasi Nilai Masukan
**Perintah:** (JSON valid, namun NIM kosong, nama kosong, dan grade bernilai 150)
```bash
curl -i -X POST http://localhost:3001/api/v1/students \
  -H "Content-Type: application/json" \
  -d '{"nim":"","name":"","grade":150}'
```
**Keluaran Terminal:**
```http
HTTP/1.1 422 Unprocessable Entity
Date: Sat, 03 Oct 2026 10:46:51 GMT
Content-Type: application/json
Content-Length: 180
X-Request-Id: 7654a98a-7193-4c43-bb5a-d70fe98485b4

{"success":false,"message":"validasi gagal","errors":{"grade":"nilai grade harus berada pada rentang 0.0 sampai 100.0","name":"nama mahasiswa wajib diisi","nim":"NIM wajib diisi"}}
```
> **Analisis:** Format JSON dipahami dengan sempurna oleh parser, namun isi semantik data tidak lolos aturan validasi domain. Respons memberikan rincian field spesifik yang bermasalah.

---

## BAGIAN 6: PEMBUKTIAN NYATA PERBEDAAN PUT DAN PATCH

Berikut adalah eksperimen yang membuktikan perbedaan semantik antara **PUT** (mengganti seluruh isi entitas) dan **PATCH** (mengubah sebagian atribut saja) pada entitas data yang sama (Mahasiswa ID 1).

### Langkah 1: Kondisi Awal Mahasiswa ID 1 (Sebelum Pengujian)
```bash
curl -i http://localhost:3001/api/v1/students/1
```
**Respons Terminal:**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{"success":true,"message":"mahasiswa ditemukan","data":{"id":1,"nim":"2308101001","name":"Putri Apriliana","grade":92.5,"is_active":true,"created_at":"2026-10-03T14:40:06.2902441+07:00"}}
```
- **Kondisi Awal:**
  - `NIM`: `"2308101001"`
  - `Name`: `"Putri Apriliana"`
  - `Grade`: `92.5`
  - `IsActive`: `true`

---

### Langkah 2: Eksekusi PUT (Mengganti Seluruh Isi Entitas)
Mengirimkan formulir baru yang lengkap untuk mengganti seluruh data entitas.
```bash
curl -i -X PUT http://localhost:3001/api/v1/students/1 \
  -H "Content-Type: application/json" \
  -d '{"nim":"2308101001","name":"Putri Apriliana S.Tr.Kom","grade":97.5,"is_active":false}'
```
**Respons Terminal:**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{"success":true,"message":"mahasiswa berhasil diganti seluruhnya (PUT)","data":{"id":1,"nim":"2308101001","name":"Putri Apriliana S.Tr.Kom","grade":97.5,"is_active":false,"created_at":"2026-10-03T14:40:06.2902441+07:00"}}
```
- **Kondisi Sesudah PUT:**
  - `Name` diperbarui menjadi `"Putri Apriliana S.Tr.Kom"`.
  - `Grade` diperbarui menjadi `97.5`.
  - `IsActive` diperbarui menjadi `false`.
- **Perilaku Bila Field Tidak Lengkap pada PUT:**
  Ketika PUT dipanggil tanpa menyertakan seluruh field wajib (misal hanya mengirim `{"name":"Putri Baru"}`), server menolak dengan status **`422 Unprocessable Entity`** (`{"errors":{"nim":"NIM wajib diisi pada PUT"}}`). Hal ini membuktikan bahwa pada metode PUT, klien diwajibkan mengirimkan seluruh representasi sumber daya.

---

### Langkah 3: Eksekusi PATCH (Pembaruan Parsial Terpilih)
Klien hanya ingin mengaktifkan kembali mahasiswa (`"is_active": true`) dan memperbarui nilai (`"grade": 99.0`). Field `nim` dan `name` **tidak dikirimkan** dalam payload body.
```bash
curl -i -X PATCH http://localhost:3001/api/v1/students/1 \
  -H "Content-Type: application/json" \
  -d '{"grade":99.0,"is_active":true}'
```
**Respons Terminal:**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{"success":true,"message":"mahasiswa berhasil diperbarui sebagian (PATCH)","data":{"id":1,"nim":"2308101001","name":"Putri Apriliana S.Tr.Kom","grade":99,"is_active":true,"created_at":"2026-10-03T14:40:06.2902441+07:00"}}
```
- **Kondisi Sesudah PATCH:**
  - `NIM` **tetap utuh**: `"2308101001"` (tidak terhapus / tidak berubah).
  - `Name` **tetap utuh**: `"Putri Apriliana S.Tr.Kom"` (tidak terhapus / tidak berubah).
  - `Grade` **berubah**: menjadi `99`.
  - `IsActive` **berubah**: menjadi `true`.

### Rangkuman Komparasi:
| Aspek | PUT | PATCH |
| :--- | :--- | :--- |
| **Tujuan** | Mengganti seluruh isi entitas (*full replacement*). | Memodifikasi sebagian atribut tertentu (*partial update*). |
| **Kelengkapan Payload** | Wajib menyertakan seluruh field. | Hanya menyertakan field yang ingin diubah. |
| **Tipe Field Struct Go** | Tipe nilai biasa (`string`, `float64`, `bool`). | Tipe pointer (`*string`, `*float64`, `*bool`). |
| **Idempotensi** | **Ya**, mengirim request PUT yang sama berulang kali menghasilkan data identik. | **Tergantung implementasi**, umumnya non-idempotent jika melakukan operasi mutasi relatif. |

---

## BAGIAN 7: ANALISIS & JAWABAN PERTANYAAN EKSPLORASI

### 1. Eksplorasi Anatomi Header HTTP (Halaman 2)
- **Header apa saja yang dikirim balik oleh situs publik (misal: go.dev) yang tidak ada pada contoh lokal?**
  Server produksi umumnya mengirim header keamanan dan optimasi seperti: `Strict-Transport-Security` (HSTS), `Content-Security-Policy` (CSP), `Server` (misal: cloudflare/nginx), `CF-RAY`, `Vary: Accept-Encoding`, `ETag`, dan `Cache-Control: public, max-age=...`.
- **Apa gunanya header `Content-Length`, dan apa yang terjadi bila nilainya salah?**
  Header `Content-Length` menyatakan ukuran persis body respons dalam satuan byte sehingga klien tahu kapan transmisi body telah selesai dibaca. Jika nilainya **terlalu kecil**, klien akan memotong pembacaan body sebelum selesai (*truncated payload*). Jika nilainya **terlalu besar**, klien akan terus menunggu sisa byte hingga terjadi *connection timeout*.
- **Apa itu `Transfer-Encoding: chunked` dan kapan dipakai menggantikan `Content-Length`?**
  `Transfer-Encoding: chunked` digunakan saat server mengirim respons secara bertahap dalam potongan-potongan (*chunks*) dinamis di mana ukuran total akhir belum diketahui saat transmisi dimulai (misal: streaming data, pembuatan berkas PDF/CSV dinamis, atau respons AI generatif). Keduanya tidak boleh digunakan bersamaan pada HTTP/1.1.

### 2. Eksplorasi PATCH dan JSON Patch (Halaman 4)
- **Apa itu JSON Merge Patch (RFC 7396) dan JSON Patch (RFC 6902)?**
  - **JSON Merge Patch (RFC 7396):** Klien mengirimkan objek JSON parsial yang langsung ditimpa ke dokumen target. Pendekatan modul ini sangat mirip dengan *JSON Merge Patch*.
  - **JSON Patch (RFC 6902):** Klien mengirimkan array instruksi perubahan formal seperti `[{"op": "replace", "path": "/grade", "value": 90}]`.
- **Bagaimana cara mengosongkan sebuah field dengan PATCH? Apa yang membuatnya rumit?**
  Jika field bertipe data pointer biasa, nilai `nil` menandakan "field tidak dikirim". Bila klien ingin secara sengaja mengosongkan string menjadi `null`, parser standar Go menganggap `null` dan field tidak ada sama-sama sebagai `nil`. Solusinya membutuhkan teknik khusus seperti *custom unmarshaler* atau sentinel pointer untuk membedakan *null explicit* dan *field missing*.
- **Mengapa tutorial banyak memakai PUT untuk ubah sebagian data? Apa akibatnya?**
  Banyak tutorial menyederhanakan penjelasan dan malas membuat struct pointer untuk PATCH. Akibatnya, pemakai API yang tidak memahami semantik PUT akan secara tidak sengaja mengosongkan (*overwrite to default zero-value*) kolom data penting yang lupa mereka sertakan dalam payload request.

### 3. Eksplorasi Status Code dan Header (Halaman 5)
- **Fungsi header `Accept` dan respon bila server tidak mendukung format tersebut:**
  Header `Accept` digunakan dalam negosiasi konten (*content negotiation*) untuk memberitahu server format apa yang diinginkan klien (misal: `Accept: application/json` atau `text/csv`). Bila server tidak mampu melayani format tersebut, server wajib mengembalikan status **`406 Not Acceptable`**.
- **Apa itu Preflight Request pada CORS?**
  Browser secara otomatis mengirimkan permintaan pendahuluan dengan metode **`OPTIONS`** sebelum mengeksekusi request berbahaya lintas domain (CORS). Browser memeriksa header seperti `Access-Control-Request-Method` dan `Access-Control-Request-Headers` untuk menentukan apakah server memberikan izin.
- **Mengapa awalan `X-` pada custom header kini tidak lagi dianjurkan?**
  RFC 6648 secara resmi menghentikan rekomendasi awalan `X-` karena ketika sebuah header eksperimental menjadi standar industri, transisi nama dari `X-Header` menjadi `Header` menimbulkan masalah backward compatibility yang merepotkan.

### 4. Eksplorasi Pemisahan Struct DTO (Halaman 9)
- **Mengapa tidak menggunakan satu struct saja untuk semuanya?**
  Memisahkan DTO ke dalam `CreateStudentRequest`, `ReplaceStudentRequest`, dan `PatchStudentRequest` menerapkan prinsip *Separation of Concerns* dan proteksi keamanan *Mass Assignment Vulnerability*. Jika hanya ada satu struct, klien jahat bisa menyusupkan atribut sensitif seperti `{"id": 999}` atau field hak akses (`role: "admin"`), yang dapat menimpa integritas sistem tanpa disengaja.

### 5. Eksplorasi Keamanan Query Parameter (Halaman 16)
- **Kirim `?limit=99999` dan `?sort=passwordadmin`. Apa yang dilakukan server dan kode mana yang mencegahnya?**
  - Pada request `?limit=99999`, fungsi `parseListStudentQuery` memeriksa `if q.Limit > 50 { q.Limit = 50 }`. Server memotong nilai limit menjadi `50` sehingga server terhindar dari *Out-Of-Memory*.
  - Pada request `?sort=passwordadmin`, fungsi memeriksa daftar putih `if !allowedSort[q.Sort] { q.Sort = "id" }`. Karena `passwordadmin` tidak terdaftar dalam map `allowedSort`, server mengabaikannya dan menggunakan nilai aman `id`.

### 6. Eksplorasi Middleware dan Idempotensi (Halaman 17)
- **Apa yang terjadi bila logger diletakkan sebelum requestid?**
  Pada keluaran format logger Fiber:
  `Format: "[${time}] ${locals:requestid} ..."`
  Nilai `${locals:requestid}` diisi oleh middleware `requestid`. Bila middleware `logger` dipasang lebih dulu, maka saat request masuk, nilai `requestid` belum di-generate dan belum tersimpan di `c.Locals("requestid")`, sehingga kolom ID penelusuran pada log akan kosong atau bernilai `-`.
- **Metode yang tidak didaftarkan (misal: `OPTIONS` atau `HEAD` ke `/api/v1/students`):**
  Framework Fiber secara otomatis menangani `HEAD` (dengan mengembalikan header yang sama seperti `GET` tanpa body) serta menangani `OPTIONS` melalui middleware CORS untuk memberikan daftar metode yang diizinkan (`Access-Control-Allow-Methods`). Jika metode benar-benar tidak didukung, server mengembalikan status **`405 Method Not Allowed`**.
- **Pembuktian Idempotensi Endpoint:**
  - `GET /api/v1/students/1`: Dipanggil 10 kali berturut-turut menghasilkan data yang sama tanpa mengubah apa pun (Idempotent & Safe).
  - `DELETE /api/v1/students/2`: Pemanggilan pertama menghapus data (204). Pemanggilan kedua dan ketiga mengembalikan 404, namun kondisi akhir di server tetap sama: data ID 2 tidak ada di sistem (Idempotent).
  - `PUT /api/v1/students/1`: Dipanggil dengan payload yang sama berkali-kali menghasilkan data akhir yang identik (Idempotent).
  - `POST /api/v1/students`: Dipanggil 3 kali akan membuat 3 data baru dengan ID berbeda (Non-Idempotent).

---

## BAGIAN 8: RIWAYAT COMMIT GIT BERTAHAP

Sesuai ketentuan penugasan ("Buatlah commit yang bertahap. Satu commit berisi seluruh pekerjaan tidak diterima"), seluruh pekerjaan dikerjakan secara modular dan terdokumentasi rapi pada log Git repositori:

```text
48ad94b docs(api-students): buat dokumentasi kontrak API lengkap pada README.md
8ff8064 feat(api-students): konfigurasi fiber app, middleware JSON check 415, dan routing api-students
f679a1c feat(api-students): implementasi handler CRUD mahasiswa lengkap dengan validasi 409, 422, dan 204
ac436b7 feat(api-students): implementasi helper response dan parsing query filter sort limit
cb9aaf1 feat(api-students): inisialisasi module dan model entitas student serta request DTO
36e260e feat(latihan-fiber): implementasi REST API users sesuai langkah panduan modul 2
45be137 chore: hapus folder api-students agar tugas-mandiri murni hanya berisi modul 1
fea574f refactor(modul-1): sempurnakan kode tugas mandiri variabel, pointer, dan struct student
```

---

## BAGIAN 9: PERNYATAAN PENGGUNAAN SUMBER BANTUAN & AI

Sesuai ketentuan pengumpulan Modul 2 poin 6, berikut adalah rincian bantuan perkakas AI yang digunakan selama penyelesaian modul praktikum:

1. **Alat Bantu AI:**
   - **Google Antigravity AI Assistant** (berbasis Gemini 3.8 Flash).
2. **Bagian yang Dibantu:**
   - Membedah modul praktikum PDF dan mengekstrak seluruh kriteria tugas mandiri serta pertanyaan eksplorasi.
   - Merancang struktur arsitektur paket terpisah (`model.go`, `helper.go`, `handler.go`, `main.go`, `README.md`).
   - Memberikan asistensi pada perancangan struct DTO pointer untuk PATCH serta mekanisme validasi keunikan NIM (409 Conflict) dan whitelist sorting.
   - Menjalankan otomasi pengujian endpoint HTTP menggunakan `curl -i` untuk menangkap kode status, header `Location`, `X-Request-Id`, dan body respons secara akurat.
   - Membantu penyusunan dokumen laporan markdown yang sistematis dan komprehensif.
3. **Pernyataan Kemandirian:**
   Mahasiswa memahami sepenuhnya alur kerja kode program yang dihasilkan, mampu menjelaskan setiap baris logika bisnis, perbedaan semantik HTTP yang diterapkan, serta siap mempertanggungjawabkannya pada sesi evaluasi praktikum.
