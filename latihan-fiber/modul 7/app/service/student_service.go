package service

import (
	"fmt"
	"strings"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"

	"github.com/gofiber/fiber/v2"
)

// StudentService memegang tanggung jawab Controller untuk entitas mahasiswa.
type StudentService struct {
	repo  repository.StudentRepository
	perms *helper.PermissionSet
}

func NewStudentService(
	repo repository.StudentRepository,
	perms *helper.PermissionSet,
) *StudentService {
	return &StudentService{repo: repo, perms: perms}
}

// List menangani permintaan GET /api/v1/students dengan Keyset (Cursor) Pagination dan Content Negotiation.
func (s *StudentService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	// Content negotiation: memilih JSON atau CSV sebelum query basis data
	format, err := helper.Negotiate(c, helper.FormatJSON, helper.FormatCSV)
	if err != nil {
		return err
	}

	q, err := helper.ParseCursorQuery(c)
	if err != nil {
		return err
	}

	// Role-Based Scope: bila tidak memiliki izin "student:read:any", hanya ambil milik sendiri
	var ownerID *int
	if !s.perms.Can(current.Role, "student:read:any") {
		ownerID = &current.UserID
	}

	rows, err := s.repo.FindAfterCursor(ctx, q, ownerID)
	if err != nil {
		return helper.Internal(err)
	}

	if format == helper.FormatCSV {
		return helper.WriteStudentsCSV(c, rows)
	}

	hasMore := len(rows) > q.Limit
	if hasMore {
		rows = rows[:q.Limit]
	}

	meta := &model.CursorMeta{Limit: q.Limit, HasMore: hasMore}
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		meta.NextCursor = helper.EncodeCursor(last.CreatedAt, last.ID)
	}

	return helper.SuccessCursor(c, "daftar mahasiswa berhasil diambil", rows, meta)
}

// Get menangani permintaan GET /api/v1/students/:id
func (s *StudentService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return TranslateError(err, "mahasiswa")
	}

	// Pemeriksaan hak akses & kepemilikan (ownership atau student:read:any)
	if !CanAccessStudent(current, student.OwnerID, s.perms, "student:read:any") {
		return helper.Forbidden("tidak berhak mengakses data mahasiswa lain")
	}

	return helper.Success(c, fiber.StatusOK, "mahasiswa ditemukan", student)
}

// Create menangani permintaan POST /api/v1/students (dijaga middleware student:create)
func (s *StudentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	// Validasi deklaratif menggunakan struct tags
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	// owner_id diisi OTOMATIS dari identitas pemanggil (current.UserID)
	// dan tidak dapat dimanipulasi dari body request klien
	newStudent := model.Student{
		NIM:      strings.TrimSpace(req.NIM),
		Name:     strings.TrimSpace(req.Name),
		Grade:    req.Grade,
		IsActive: req.IsActive,
		OwnerID:  current.UserID,
	}

	created, err := s.repo.Create(ctx, newStudent)
	if err != nil {
		return TranslateError(err, "mahasiswa")
	}

	location := fmt.Sprintf("/api/v1/students/%d", created.ID)
	return helper.Created(c, "data mahasiswa berhasil ditambahkan", created, location)
}

// Replace menangani permintaan PUT /api/v1/students/:id
func (s *StudentService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return TranslateError(err, "mahasiswa")
	}

	// Pemeriksaan hak akses & kepemilikan (ownership atau student:update:any)
	if !CanAccessStudent(current, existing.OwnerID, s.perms, "student:update:any") {
		return helper.Forbidden("tidak berhak mengubah data mahasiswa lain")
	}

	var req model.ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	existing.NIM = strings.TrimSpace(req.NIM)
	existing.Name = strings.TrimSpace(req.Name)
	existing.Grade = req.Grade
	existing.IsActive = req.IsActive
	// owner_id tetap dipertahankan milik pemilik semula

	updated, err := s.repo.Update(ctx, existing)
	if err != nil {
		return TranslateError(err, "mahasiswa")
	}

	return helper.Success(c, fiber.StatusOK, "data mahasiswa berhasil diperbarui", updated)
}

// Patch menangani permintaan PATCH /api/v1/students/:id
func (s *StudentService) Patch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return TranslateError(err, "mahasiswa")
	}

	// Pemeriksaan hak akses & kepemilikan (ownership atau student:update:any)
	if !CanAccessStudent(current, existing.OwnerID, s.perms, "student:update:any") {
		return helper.Forbidden("tidak berhak mengubah data mahasiswa lain")
	}

	var req model.PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if IsEmptyStudentPatch(req) {
		return helper.BadRequest("setidaknya satu field harus dikirim")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	patched := ApplyStudentPatch(existing, req)

	updated, err := s.repo.Update(ctx, patched)
	if err != nil {
		return TranslateError(err, "mahasiswa")
	}

	return helper.Success(c, fiber.StatusOK, "mahasiswa berhasil diperbarui", updated)
}

// Delete menangani permintaan DELETE /api/v1/students/:id (dijaga middleware student:delete)
func (s *StudentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return TranslateError(err, "mahasiswa")
	}

	return helper.NoContent(c)
}
