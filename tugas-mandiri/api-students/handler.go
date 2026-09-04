package main

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// students adalah "database" in-memory (slice).
// Data akan hilang saat server di-restart.
var students = []Student{
	{
		ID:        1,
		NIM:       "434241022",
		Name:      "Putri",
		Grade:     3.89,
		IsActive:  true,
		CreatedAt: time.Now(),
	},
	{
		ID:        2,
		NIM:       "434241023",
		Name:      "Khansa",
		Grade:     3.50,
		IsActive:  true,
		CreatedAt: time.Now(),
	},
	{
		ID:        3,
		NIM:       "434241024",
		Name:      "Damara",
		Grade:     3.65,
		IsActive:  false,
		CreatedAt: time.Now(),
	},
}
var nextID = 4

// ==================== GET /api/v1/students ====================
// Menampilkan daftar student dengan paginasi, search, sort, filter.
func ListStudents(c *fiber.Ctx) error {
	// --- 1. Ambil query parameter dengan nilai bawaan ---
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	search := c.Query("search", "")
	sortBy := c.Query("sort", "id")
	order := c.Query("order", "asc")
	filterActive := c.Query("is_active", "")
	filterGrade := c.Query("grade", "")

	// Batas atas limit agar tidak terlalu besar (alasan: mencegah
	// client meminta seluruh data sekaligus yang bisa membebani server)
	if limit > 100 {
		limit = 100
	}
	if page < 1 {
		page = 1
	}

	// --- 2. Filter ---
	filtered := make([]Student, 0)
	for _, s := range students {
		// Filter berdasarkan search (nama, case-insensitive)
		if search != "" && !strings.Contains(
			strings.ToLower(s.Name), strings.ToLower(search)) {
			continue
		}

		// Filter berdasarkan is_active
		if filterActive != "" {
			isActive, err := strconv.ParseBool(filterActive)
			if err == nil && s.IsActive != isActive {
				continue
			}
		}

		// Filter berdasarkan rentang grade
		if filterGrade != "" {
			switch strings.ToUpper(filterGrade) {
			case "A":
				if s.Grade < 3.50 {
					continue
				}
			case "B":
				if s.Grade < 3.00 || s.Grade >= 3.50 {
					continue
				}
			case "C":
				if s.Grade < 2.50 || s.Grade >= 3.00 {
					continue
				}
			case "D":
				if s.Grade >= 2.50 {
					continue
				}
			}
		}

		filtered = append(filtered, s)
	}

	// --- 3. Sort ---
	// Daftar putih field yang boleh diurutkan
	allowedSort := map[string]bool{
		"id": true, "nim": true, "name": true, "grade": true,
	}
	if !allowedSort[sortBy] {
		sortBy = "id"
	}

	sort.Slice(filtered, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "name":
			less = strings.ToLower(filtered[i].Name) < strings.ToLower(filtered[j].Name)
		case "grade":
			less = filtered[i].Grade < filtered[j].Grade
		case "nim":
			less = filtered[i].NIM < filtered[j].NIM
		default: // id
			less = filtered[i].ID < filtered[j].ID
		}
		if order == "desc" {
			return !less
		}
		return less
	})

	// --- 4. Paginasi ---
	total := len(filtered)
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	start := (page - 1) * limit
	end := start + limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paged := filtered[start:end]

	// --- 5. Response dengan meta ---
	return okWithMeta(c, "daftar mahasiswa", paged, &Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

// ==================== GET /api/v1/students/:id ================
// Menampilkan satu student berdasarkan ID, atau 404 jika tidak ada.
func GetStudent(c *fiber.Ctx) error {
	// Ambil parameter :id dari URL
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		// ID bukan angka → 400
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka")
	}

	// Cari student di slice
	for _, s := range students {
		if s.ID == id {
			return ok(c, "data mahasiswa ditemukan", s)
		}
	}

	// Tidak ditemukan → 404
	return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
}

// ==================== POST /api/v1/students ===================
// Membuat student baru, mengembalikan 201 + header Location.
func CreateStudent(c *fiber.Ctx) error {

	// Parse body JSON
	var req CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}

	errs := map[string]string{}
	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "nim tidak boleh kosong"
	}
	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "name tidak boleh kosong"
	}
	if req.Grade < 0 || req.Grade > 4.0 {
		errs["grade"] = "grade harus antara 0 dan 4.0"
	}

	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	// Cek NIM duplikat → 409
	for _, s := range students {
		if s.NIM == req.NIM {
			return fail(c, fiber.StatusConflict, "nim sudah digunakan oleh mahasiswa lain")
		}
	}

	// Buat student baru
	newStudent := Student{
		ID:        nextID,
		NIM:       req.NIM,
		Name:      req.Name,
		Grade:     req.Grade,
		IsActive:  req.IsActive,
		CreatedAt: time.Now(),
	}
	nextID++
	students = append(students, newStudent)

	// 201 + header Location
	location := fmt.Sprintf("/api/v1/students/%d", newStudent.ID)
	return created(c, "mahasiswa berhasil ditambahkan", newStudent, location)
}

// ==================== PUT /api/v1/students/:id ================
// Mengganti SELURUH data student (semua field wajib dikirim).
func ReplaceStudent(c *fiber.Ctx) error {
	// Parse ID
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka")
	}

	// Parse body
	var req UpdateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}

	// Validasi per field → 422
	errs := make(map[string]string)
	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "nim tidak boleh kosong"
	}
	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "name tidak boleh kosong"
	}
	if req.Grade < 0 || req.Grade > 4.0 {
		errs["grade"] = "grade harus antara 0 dan 4.0"
	}
	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	// Cari student & replace seluruh field
	for i, s := range students {
		if s.ID == id {
			// Cek NIM duplikat (kecuali milik diri sendiri)
			for _, other := range students {
				if other.NIM == req.NIM && other.ID != id {
					return fail(c, fiber.StatusConflict,
						"nim sudah digunakan oleh mahasiswa lain")
				}
			}

			// Replace seluruh isi — inilah bedanya PUT dengan PATCH
			students[i].NIM = req.NIM
			students[i].Name = req.Name
			students[i].Grade = req.Grade
			students[i].IsActive = req.IsActive

			return ok(c, "data mahasiswa berhasil diperbarui (PUT)", students[i])
		}
	}

	return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
}

// ==================== PATCH /api/v1/students/:id ==============
// Mengubah SEBAGIAN data student (hanya field yang dikirim).
func PatchStudent(c *fiber.Ctx) error {
	// Parse ID
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka")
	}

	// Parse body
	var req PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}

	if req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil {
		return fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
	}
	// Cari student & update hanya field yang dikirim (tidak nil)
	for i, s := range students {
		if s.ID == id {
			// Validasi per field yang dikirim → 422
			errs := make(map[string]string)

			if req.NIM != nil {
				if strings.TrimSpace(*req.NIM) == "" {
					errs["nim"] = "nim tidak boleh kosong"
				} else {
					// Cek duplikat NIM
					for _, other := range students {
						if other.NIM == *req.NIM && other.ID != id {
							return fail(c, fiber.StatusConflict,
								"nim sudah digunakan oleh mahasiswa lain")
						}
					}
				}
			}
			if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
				errs["name"] = "name tidak boleh kosong"
			}
			if req.Grade != nil && (*req.Grade < 0 || *req.Grade > 4.0) {
				errs["grade"] = "grade harus antara 0 dan 4.0"
			}

			if len(errs) > 0 {
				return failValidation(c, errs)
			}

			// Update hanya field yang dikirim — inilah bedanya PATCH dengan PUT
			if req.NIM != nil {
				students[i].NIM = *req.NIM
			}
			if req.Name != nil {
				students[i].Name = *req.Name
			}
			if req.Grade != nil {
				students[i].Grade = *req.Grade
			}
			if req.IsActive != nil {
				students[i].IsActive = *req.IsActive
			}

			return ok(c, "data mahasiswa berhasil diperbarui (PATCH)", students[i])
		}
	}

	return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
}

// ==================== DELETE /api/v1/students/:id =============
// Menghapus student, mengembalikan 204 tanpa body.
func DeleteStudent(c *fiber.Ctx) error {
	// Parse ID
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka")
	}

	// Cari dan hapus dari slice
	for i, s := range students {
		if s.ID == id {
			students = append(students[:i], students[i+1:]...)
			// 204 No Content — tanpa body
			return c.SendStatus(fiber.StatusNoContent)
		}
	}

	return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
}
