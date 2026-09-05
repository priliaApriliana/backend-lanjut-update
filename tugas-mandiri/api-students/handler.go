package main

import (
    "errors"
    "fmt"
    "strconv"
    "strings"

    "github.com/gofiber/fiber/v2"

    "api-students/app/model"
    "api-students/app/repository"
)

type StudentHandler struct {
    repo repository.StudentRepository
}

// NewStudentHandler menerima INTERFACE, bukan struct konkret.
// Handler tidak tahu dan tidak perlu tahu datanya disimpan di mana.
func NewStudentHandler(repo repository.StudentRepository) *StudentHandler {
    return &StudentHandler{repo: repo}
}

// terjemahkanError memetakan error milik repository menjadi status HTTP.
// Satu tempat untuk seluruh handler, agar pemetaannya tidak tercecer.
func terjemahkanError(c *fiber.Ctx, err error, pesanUmum string) error {
    switch {
    case errors.Is(err, repository.ErrNotFound):
        return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
    case errors.Is(err, repository.ErrDuplicate):
        return fail(c, fiber.StatusConflict, "nim sudah digunakan oleh mahasiswa lain")
    default:
        return fail(c, fiber.StatusInternalServerError, pesanUmum)
    }
}

// ==================== GET /api/v1/students ====================
func (h *StudentHandler) List(c *fiber.Ctx) error {
    ctx, cancel := reqCtx(c)
    defer cancel()

    q := parseListQuery(c)

    students, total, err := h.repo.FindAll(ctx, q)
    if err != nil {
        return fail(c, fiber.StatusInternalServerError, "gagal mengambil data mahasiswa")
    }

    totalPages := 0
    if q.Limit > 0 {
        totalPages = (total + q.Limit - 1) / q.Limit
    }

    return okWithMeta(c, "daftar mahasiswa", students, &Meta{
        Page: q.Page, Limit: q.Limit, Total: total, TotalPages: totalPages,
    })
}

// ==================== GET /api/v1/students/:id ================
func (h *StudentHandler) Get(c *fiber.Ctx) error {
    ctx, cancel := reqCtx(c)
    defer cancel()

    id, err := strconv.Atoi(c.Params("id"))
    if err != nil {
        return fail(c, fiber.StatusBadRequest, "id harus berupa angka")
    }

    student, err := h.repo.FindByID(ctx, id)
    if err != nil {
        return terjemahkanError(c, err, "gagal mengambil data mahasiswa")
    }
    return ok(c, "data mahasiswa ditemukan", student)
}

// ==================== POST /api/v1/students ===================
func (h *StudentHandler) Create(c *fiber.Ctx) error {
    ctx, cancel := reqCtx(c)
    defer cancel()

    var req model.CreateStudentRequest
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

    newStudent, err := h.repo.Create(ctx, model.Student{
        NIM: req.NIM, Name: req.Name, Grade: req.Grade, IsActive: req.IsActive,
    })
    if err != nil {
        return terjemahkanError(c, err, "gagal menyimpan mahasiswa")
    }

    location := fmt.Sprintf("/api/v1/students/%d", newStudent.ID)
    return created(c, "mahasiswa berhasil ditambahkan", newStudent, location)
}

// ==================== PUT /api/v1/students/:id ================
func (h *StudentHandler) Replace(c *fiber.Ctx) error {
    ctx, cancel := reqCtx(c)
    defer cancel()

    id, err := strconv.Atoi(c.Params("id"))
    if err != nil {
        return fail(c, fiber.StatusBadRequest, "id harus berupa angka")
    }

    var req model.UpdateStudentRequest
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

    updated, err := h.repo.Update(ctx, model.Student{
        ID: id, NIM: req.NIM, Name: req.Name, Grade: req.Grade, IsActive: req.IsActive,
    })
    if err != nil {
        return terjemahkanError(c, err, "gagal memperbarui mahasiswa")
    }

    return ok(c, "data mahasiswa berhasil diperbarui (PUT)", updated)
}

// ==================== PATCH /api/v1/students/:id ==============
func (h *StudentHandler) Patch(c *fiber.Ctx) error {
    ctx, cancel := reqCtx(c)
    defer cancel()

    id, err := strconv.Atoi(c.Params("id"))
    if err != nil {
        return fail(c, fiber.StatusBadRequest, "id harus berupa angka")
    }

    var req model.PatchStudentRequest
    if err := c.BodyParser(&req); err != nil {
        return fail(c, fiber.StatusBadRequest, "body JSON tidak valid")
    }
    if req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil {
        return fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
    }

    saatIni, err := h.repo.FindByID(ctx, id)
    if err != nil {
        return terjemahkanError(c, err, "gagal mengambil data mahasiswa")
    }

    errs := map[string]string{}
    if req.NIM != nil {
        if strings.TrimSpace(*req.NIM) == "" {
            errs["nim"] = "nim tidak boleh kosong"
        } else {
            saatIni.NIM = *req.NIM
        }
    }
    if req.Name != nil {
        if strings.TrimSpace(*req.Name) == "" {
            errs["name"] = "name tidak boleh kosong"
        } else {
            saatIni.Name = *req.Name
        }
    }
    if req.Grade != nil {
        if *req.Grade < 0 || *req.Grade > 4.0 {
            errs["grade"] = "grade harus antara 0 dan 4.0"
        } else {
            saatIni.Grade = *req.Grade
        }
    }
    if req.IsActive != nil {
        saatIni.IsActive = *req.IsActive
    }
    if len(errs) > 0 {
        return failValidation(c, errs)
    }

    updated, err := h.repo.Update(ctx, saatIni)
    if err != nil {
        return terjemahkanError(c, err, "gagal memperbarui mahasiswa")
    }

    return ok(c, "data mahasiswa berhasil diperbarui (PATCH)", updated)
}

// ==================== DELETE /api/v1/students/:id =============
func (h *StudentHandler) Delete(c *fiber.Ctx) error {
    ctx, cancel := reqCtx(c)
    defer cancel()

    id, err := strconv.Atoi(c.Params("id"))
    if err != nil {
        return fail(c, fiber.StatusBadRequest, "id harus berupa angka")
    }

    if err := h.repo.Delete(ctx, id); err != nil {
        return terjemahkanError(c, err, "gagal menghapus mahasiswa")
    }
    return c.SendStatus(fiber.StatusNoContent)
}
