package service

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

type StudentService struct {
	repo  repository.StudentRepository
	perms *helper.PermissionSet
}

// NewStudentService menerima INTERFACE repository dan PermissionSet.
// Modul 6: perms ditambahkan untuk pemeriksaan kepemilikan.
func NewStudentService(
	repo repository.StudentRepository,
	perms *helper.PermissionSet,
) *StudentService {
	return &StudentService{repo: repo, perms: perms}
}

// terjemahkanError memetakan error milik repository menjadi status HTTP.
func terjemahkanError(c *fiber.Ctx, err error, pesanUmum string) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.Fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Fail(c, fiber.StatusConflict, "nim sudah digunakan oleh mahasiswa lain")
	default:
		return helper.Fail(c, fiber.StatusInternalServerError, pesanUmum)
	}
}

// ==================== GET /api/v1/students ====================
func (s *StudentService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	q := helper.ParseListQuery(c)

	students, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data mahasiswa")
	}

	return helper.OkWithMeta(c, "daftar mahasiswa", students, &model.Meta{
		Page: q.Page, Limit: q.Limit, Total: total, TotalPages: CountTotalPages(total, q.Limit),
	})
}

// ==================== GET /api/v1/students/:id ================
func (s *StudentService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current := helper.CurrentUser(c)
	if current == nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka")
	}

	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return terjemahkanError(c, err, "gagal mengambil data mahasiswa")
	}

	// Cek kepemilikan: milik sendiri ATAU punya permission student:read:any.
	if !CanAccessStudent(*current, student.OwnerID, s.perms, "student:read:any") {
		return helper.Fail(c, fiber.StatusForbidden,
			"tidak berhak mengakses data mahasiswa milik user lain")
	}

	return helper.Ok(c, "data mahasiswa ditemukan", student)
}

// ==================== POST /api/v1/students ===================
func (s *StudentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current := helper.CurrentUser(c)
	if current == nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}

	if errs := ValidateCreate(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	// owner_id diisi otomatis dari identitas pengguna yang sedang login.
	// Klien tidak bisa menyetel owner_id lewat request body.
	newStudent, err := s.repo.Create(ctx, model.Student{
		NIM: req.NIM, Name: req.Name, Grade: req.Grade, IsActive: req.IsActive,
		OwnerID: current.ID,
	})
	if err != nil {
		return terjemahkanError(c, err, "gagal menyimpan mahasiswa")
	}

	location := fmt.Sprintf("/api/v1/students/%d", newStudent.ID)
	return helper.Created(c, "mahasiswa berhasil ditambahkan", newStudent, location)
}

// ==================== PUT /api/v1/students/:id ================
func (s *StudentService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current := helper.CurrentUser(c)
	if current == nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka")
	}

	// Ambil data saat ini untuk cek kepemilikan.
	saatIni, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return terjemahkanError(c, err, "gagal mengambil data mahasiswa")
	}

	if !CanAccessStudent(*current, saatIni.OwnerID, s.perms, "student:update:any") {
		return helper.Fail(c, fiber.StatusForbidden,
			"tidak berhak mengubah data mahasiswa milik user lain")
	}

	var req model.UpdateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}

	if errs := ValidateReplace(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	updated, err := s.repo.Update(ctx, model.Student{
		ID: id, NIM: req.NIM, Name: req.Name, Grade: req.Grade, IsActive: req.IsActive,
	})
	if err != nil {
		return terjemahkanError(c, err, "gagal memperbarui mahasiswa")
	}

	return helper.Ok(c, "data mahasiswa berhasil diperbarui (PUT)", updated)
}

// ==================== PATCH /api/v1/students/:id ==============
func (s *StudentService) Patch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current := helper.CurrentUser(c)
	if current == nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka")
	}

	saatIni, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return terjemahkanError(c, err, "gagal mengambil data mahasiswa")
	}

	// Hak diperiksa lebih dulu, sebelum isi body divalidasi.
	if !CanAccessStudent(*current, saatIni.OwnerID, s.perms, "student:update:any") {
		return helper.Fail(c, fiber.StatusForbidden,
			"tidak berhak mengubah data mahasiswa milik user lain")
	}

	var req model.PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}
	if IsEmptyPatch(req) {
		return helper.Fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
	}

	updated, errs := ApplyPatch(saatIni, req)
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	result, err := s.repo.Update(ctx, updated)
	if err != nil {
		return terjemahkanError(c, err, "gagal memperbarui mahasiswa")
	}

	return helper.Ok(c, "data mahasiswa berhasil diperbarui (PATCH)", result)
}

// ==================== DELETE /api/v1/students/:id =============
// Dijaga middleware dengan permission student:delete, jadi keputusan
// hak sudah diambil di route. Service hanya menghapus.
func (s *StudentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return terjemahkanError(c, err, "gagal menghapus mahasiswa")
	}
	return helper.NoContent(c)
}
