package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Data in-memory penyimpanan mahasiswa
var students = []Student{
	{
		ID:        1,
		NIM:       "2308101001",
		Name:      "Putri Apriliana",
		Grade:     92.5,
		IsActive:  true,
		CreatedAt: time.Now().Add(-3 * time.Hour),
	},
	{
		ID:        2,
		NIM:       "2308101002",
		Name:      "Budi Santoso",
		Grade:     85.0,
		IsActive:  true,
		CreatedAt: time.Now().Add(-2 * time.Hour),
	},
	{
		ID:        3,
		NIM:       "2308101003",
		Name:      "Citra Lestari",
		Grade:     74.0,
		IsActive:  false,
		CreatedAt: time.Now().Add(-1 * time.Hour),
	},
}

var nextID = 4

// findStudentIndex mencari indeks mahasiswa dalam slice berdasarkan ID numerik.
func findStudentIndex(id int) int {
	for i := range students {
		if students[i].ID == id {
			return i
		}
	}
	return -1
}

// cocokPencarian memeriksa apakah kata kunci pencarian cocok dengan nama atau NIM mahasiswa (case-insensitive).
func cocokPencarian(s Student, kata string) bool {
	kata = strings.ToLower(kata)
	return strings.Contains(strings.ToLower(s.Name), kata) ||
		strings.Contains(strings.ToLower(s.NIM), kata)
}

// isNIMDuplicate mengecek apakah NIM sudah digunakan oleh mahasiswa lain (mengecualikan excludeID).
func isNIMDuplicate(nim string, excludeID int) bool {
	for _, s := range students {
		if s.ID != excludeID && strings.EqualFold(s.NIM, nim) {
			return true
		}
	}
	return false
}

// paramID mem-parsing parameter :id dari route dan memvalidasi nilainya harus angka bulat positif.
func paramID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

// listStudents menangani GET /api/v1/students
// Mendukung paginasi, pencarian case-insensitive, pengurutan whitelist, serta filter status dan grade.
func listStudents(c *fiber.Ctx) error {
	q := parseListStudentQuery(c)

	// 1) Saring (Filter)
	hasil := []Student{}
	for _, s := range students {
		if q.IsActive != nil && s.IsActive != *q.IsActive {
			continue
		}
		if q.MinGrade != nil && s.Grade < *q.MinGrade {
			continue
		}
		if q.MaxGrade != nil && s.Grade > *q.MaxGrade {
			continue
		}
		if q.Search != "" && !cocokPencarian(s, q.Search) {
			continue
		}
		hasil = append(hasil, s)
	}

	// 2) Urutkan (Sort & Order)
	sort.SliceStable(hasil, func(i, j int) bool {
		var lebihKecil bool
		switch q.Sort {
		case "nim":
			lebihKecil = hasil[i].NIM < hasil[j].NIM
		case "name":
			lebihKecil = strings.ToLower(hasil[i].Name) < strings.ToLower(hasil[j].Name)
		case "grade":
			lebihKecil = hasil[i].Grade < hasil[j].Grade
		case "created_at":
			lebihKecil = hasil[i].CreatedAt.Before(hasil[j].CreatedAt)
		default: // "id"
			lebihKecil = hasil[i].ID < hasil[j].ID
		}

		if q.Order == "desc" {
			return !lebihKecil
		}
		return lebihKecil
	})

	// 3) Potong sesuai halaman (Pagination)
	total := len(hasil)
	totalPages := 0
	if total > 0 {
		totalPages = (total + q.Limit - 1) / q.Limit
	}

	mulai := (q.Page - 1) * q.Limit
	if mulai > total {
		mulai = total
	}
	akhir := mulai + q.Limit
	if akhir > total {
		akhir = total
	}

	return okList(c, "daftar mahasiswa berhasil diambil", hasil[mulai:akhir], &Meta{
		Page:       q.Page,
		Limit:      q.Limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

// getStudent menangani GET /api/v1/students/:id
func getStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}

	return ok(c, "mahasiswa ditemukan", students[i])
}

// createStudent menangani POST /api/v1/students
func createStudent(c *fiber.Ctx) error {
	var req CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	errs := map[string]string{}
	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	if req.NIM == "" {
		errs["nim"] = "NIM wajib diisi"
	}
	if req.Name == "" {
		errs["name"] = "nama mahasiswa wajib diisi"
	}
	if req.Grade < 0.0 || req.Grade > 100.0 {
		errs["grade"] = "nilai grade harus berada pada rentang 0.0 sampai 100.0"
	}

	// Kembalikan 422 jika validasi field masukan gagal
	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	// Kembalikan 409 Conflict bila NIM sudah digunakan
	if isNIMDuplicate(req.NIM, 0) {
		return fail(c, fiber.StatusConflict, fmt.Sprintf("NIM '%s' sudah terdaftar pada sistem", req.NIM))
	}

	baru := Student{
		ID:        nextID,
		NIM:       req.NIM,
		Name:      req.Name,
		Grade:     req.Grade,
		IsActive:  req.IsActive,
		CreatedAt: time.Now(),
	}
	students = append(students, baru)
	nextID++

	location := fmt.Sprintf("/api/v1/students/%d", baru.ID)
	return created(c, "mahasiswa berhasil dibuat", baru, location)
}

// replaceStudent menangani PUT /api/v1/students/:id
// Semantik PUT: Mengganti seluruh isi sumber daya. Seluruh field wajib dikirimkan.
func replaceStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}

	var req ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	errs := map[string]string{}
	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	if req.NIM == "" {
		errs["nim"] = "NIM wajib diisi pada PUT"
	}
	if req.Name == "" {
		errs["name"] = "nama mahasiswa wajib diisi pada PUT"
	}
	if req.Grade < 0.0 || req.Grade > 100.0 {
		errs["grade"] = "nilai grade harus berada pada rentang 0.0 sampai 100.0"
	}

	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	// Cek konflik NIM ganda dengan mahasiswa lain
	if isNIMDuplicate(req.NIM, id) {
		return fail(c, fiber.StatusConflict, fmt.Sprintf("NIM '%s' sudah digunakan oleh mahasiswa lain", req.NIM))
	}

	// Timpa seluruh data field
	students[i].NIM = req.NIM
	students[i].Name = req.Name
	students[i].Grade = req.Grade
	students[i].IsActive = req.IsActive

	return ok(c, "mahasiswa berhasil diganti seluruhnya (PUT)", students[i])
}

// patchStudent menangani PATCH /api/v1/students/:id
// Semantik PATCH: Mengubah sebagian atribut saja. Hanya field yang dikirim (pointer != nil) yang dimodifikasi.
func patchStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}

	var req PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	// Wajib mengirimkan minimal 1 field untuk diubah
	if req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil {
		return fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
	}

	errs := map[string]string{}

	if req.NIM != nil {
		trimmed := strings.TrimSpace(*req.NIM)
		if trimmed == "" {
			errs["nim"] = "NIM tidak boleh kosong"
		} else if isNIMDuplicate(trimmed, id) {
			return fail(c, fiber.StatusConflict, fmt.Sprintf("NIM '%s' sudah digunakan oleh mahasiswa lain", trimmed))
		} else {
			*req.NIM = trimmed
		}
	}

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			errs["name"] = "nama mahasiswa tidak boleh kosong"
		} else {
			*req.Name = trimmed
		}
	}

	if req.Grade != nil {
		if *req.Grade < 0.0 || *req.Grade > 100.0 {
			errs["grade"] = "nilai grade harus berada pada rentang 0.0 sampai 100.0"
		}
	}

	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	// Modifikasi hanya field yang dikirim
	if req.NIM != nil {
		students[i].NIM = *req.NIM
	}
	if req.Name != nil {
		students[i].Name = *req.Name
	}
	if req.Grade != nil {
		// Menggunakan method pointer receiver UpdateGrade dari tugas modul 1
		students[i].UpdateGrade(*req.Grade)
	}
	if req.IsActive != nil {
		// Menggunakan method pointer receiver Activate / Deactivate dari tugas modul 1
		if *req.IsActive {
			students[i].Activate()
		} else {
			students[i].Deactivate()
		}
	}

	return ok(c, "mahasiswa berhasil diperbarui sebagian (PATCH)", students[i])
}

// deleteStudent menangani DELETE /api/v1/students/:id
func deleteStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}

	// Hapus elemen dari slice
	students = append(students[:i], students[i+1:]...)

	// Kembalikan 204 No Content tanpa response body
	return noContent(c)
}
