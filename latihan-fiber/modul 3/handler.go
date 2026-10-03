package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"api-students/app/model"
	"api-students/app/repository"

	"github.com/gofiber/fiber/v2"
)

type StudentHandler struct {
	repo repository.StudentRepository
}

// NewStudentHandler menginisialisasi handler dengan dependensi repository (Dependency Injection).
// Handler hanya berkomunikasi melalui interface StudentRepository, tanpa mengetahui apakah datanya
// disimpan di PostgreSQL, SQLite, atau mock testing.
func NewStudentHandler(repo repository.StudentRepository) *StudentHandler {
	return &StudentHandler{repo: repo}
}

// terjemahkanError memetakan error milik repository menjadi status HTTP yang jujur dan semantik.
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

func paramID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

// List menangani GET /api/v1/students dengan query filter, sort, dan paginasi langsung di PostgreSQL
func (h *StudentHandler) List(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	q := parseListStudentQuery(c)
	students, total, err := h.repo.FindAll(ctx, q)
	if err != nil {
		return fail(c, fiber.StatusInternalServerError, "gagal mengambil data mahasiswa")
	}

	totalPages := 0
	if q.Limit > 0 {
		totalPages = (total + q.Limit - 1) / q.Limit
	}

	return okList(c, "daftar mahasiswa berhasil diambil", students, &model.Meta{
		Page:       q.Page,
		Limit:      q.Limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

// Get menangani GET /api/v1/students/:id
func (h *StudentHandler) Get(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	student, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return terjemahkanError(c, err, "gagal mengambil data mahasiswa")
	}

	return ok(c, "mahasiswa ditemukan", student)
}

// Create menangani POST /api/v1/students
func (h *StudentHandler) Create(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	errs := map[string]string{}
	if req.NIM == "" {
		errs["nim"] = "NIM wajib diisi"
	}
	if req.Name == "" {
		errs["name"] = "nama mahasiswa wajib diisi"
	}
	if req.Grade < 0.0 || req.Grade > 100.0 {
		errs["grade"] = "nilai grade harus berada pada rentang 0.0 sampai 100.0"
	}

	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	// Keunikan NIM dijamin sepenuhnya oleh UNIQUE INDEX di database PostgreSQL.
	// Jika terjadi bentrokan NIM, PostgreSQL menghasilkan kode error 23505 yang diterjemahkan menjadi 409 Conflict.
	baru, err := h.repo.Create(ctx, model.Student{
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,
	})
	if err != nil {
		return terjemahkanError(c, err, "gagal menyimpan mahasiswa")
	}

	location := fmt.Sprintf("/api/v1/students/%d", baru.ID)
	return created(c, "mahasiswa berhasil dibuat", baru, location)
}

// Replace menangani PUT /api/v1/students/:id (seluruh field wajib ada)
func (h *StudentHandler) Replace(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	var req model.ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	errs := map[string]string{}
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

	hasil, err := h.repo.Update(ctx, model.Student{
		ID:       id,
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,
	})
	if err != nil {
		return terjemahkanError(c, err, "gagal memperbarui mahasiswa")
	}

	return ok(c, "mahasiswa berhasil diganti seluruhnya (PUT)", hasil)
}

// Patch menangani PATCH /api/v1/students/:id (pembaruan parsial)
func (h *StudentHandler) Patch(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	var req model.PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	if req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil {
		return fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
	}

	// Alur PATCH: Baca dulu data saat ini dari basis data
	saatIni, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return terjemahkanError(c, err, "gagal mengambil data mahasiswa")
	}

	errs := map[string]string{}
	if req.NIM != nil {
		trimmed := strings.TrimSpace(*req.NIM)
		if trimmed == "" {
			errs["nim"] = "NIM tidak boleh kosong"
		} else {
			saatIni.NIM = trimmed
		}
	}

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			errs["name"] = "nama mahasiswa tidak boleh kosong"
		} else {
			saatIni.Name = trimmed
		}
	}

	if req.Grade != nil {
		if *req.Grade < 0.0 || *req.Grade > 100.0 {
			errs["grade"] = "nilai grade harus berada pada rentang 0.0 sampai 100.0"
		} else {
			// Menggunakan method pointer receiver UpdateGrade
			saatIni.UpdateGrade(*req.Grade)
		}
	}

	if req.IsActive != nil {
		// Menggunakan method pointer receiver Activate/Deactivate
		if *req.IsActive {
			saatIni.Activate()
		} else {
			saatIni.Deactivate()
		}
	}

	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	// Simpan entitas yang telah dimodifikasi sebagian ke basis data
	hasil, err := h.repo.Update(ctx, saatIni)
	if err != nil {
		return terjemahkanError(c, err, "gagal memperbarui mahasiswa")
	}

	return ok(c, "mahasiswa berhasil diperbarui sebagian (PATCH)", hasil)
}

// Delete menangani DELETE /api/v1/students/:id
func (h *StudentHandler) Delete(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	if err := h.repo.Delete(ctx, id); err != nil {
		return terjemahkanError(c, err, "gagal menghapus mahasiswa")
	}

	return noContent(c)
}
