package service

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

// UserService menangani semua handler HTTP untuk manajemen user.
type UserService struct {
	repo  repository.UserRepository
	perms *helper.PermissionSet
}

// NewUserService merakit UserService dengan repository dan PermissionSet.
func NewUserService(
	repo repository.UserRepository,
	perms *helper.PermissionSet,
) *UserService {
	return &UserService{repo: repo, perms: perms}
}

// translateUserError memetakan error repository ke HTTP status.
// Terpisah dari terjemahkanError di student_service.go karena
// pesan error disesuaikan untuk konteks user.
func translateUserError(c *fiber.Ctx, err error, pesanUmum string) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.Fail(c, fiber.StatusNotFound, "user tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Fail(c, fiber.StatusConflict, "username atau email sudah digunakan")
	default:
		return helper.Fail(c, fiber.StatusInternalServerError, pesanUmum)
	}
}

// ==================== GET /api/v1/users ====================

func (s *UserService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	users, err := s.repo.FindAll(ctx)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError,
			"gagal mengambil daftar user")
	}

	return helper.Ok(c, "daftar user", users)
}

// ==================== GET /api/v1/users/:id ====================

func (s *UserService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current := helper.CurrentUser(c)
	if current == nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	// Pemeriksaan hak akses SEBELUM data diambil.
	// Urutan ini disengaja: kalau dicek sesudah data diambil,
	// penyerang bisa tahu apakah id tertentu ada atau tidak
	// berdasarkan perbedaan respons (404 vs 403).
	if !CanAccessUser(*current, id, s.perms, "user:read:any") {
		return helper.Fail(c, fiber.StatusForbidden,
			"tidak berhak mengakses data user lain")
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateUserError(c, err, "gagal mengambil data user")
	}

	return helper.Ok(c, "data user ditemukan", user)
}

// ==================== PUT /api/v1/users/:id ====================

func (s *UserService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current := helper.CurrentUser(c)
	if current == nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	// Pemeriksaan kepemilikan / permission user:update:any.
	if !CanAccessUser(*current, id, s.perms, "user:update:any") {
		return helper.Fail(c, fiber.StatusForbidden,
			"tidak berhak mengubah data user lain")
	}

	var req model.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}

	// Validasi sederhana.
	errs := map[string]string{}
	if strings.TrimSpace(req.Username) == "" {
		errs["username"] = "username tidak boleh kosong"
	}
	if strings.TrimSpace(req.Email) == "" {
		errs["email"] = "email tidak boleh kosong"
	}
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	updated, err := s.repo.Update(ctx, model.User{
		ID: id, Username: req.Username, Email: req.Email, IsActive: req.IsActive,
	})
	if err != nil {
		return translateUserError(c, err, "gagal memperbarui user")
	}

	return helper.Ok(c, "data user berhasil diperbarui (PUT)", updated)
}

// ==================== PATCH /api/v1/users/:id ====================

func (s *UserService) Patch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current := helper.CurrentUser(c)
	if current == nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	// Pemeriksaan kepemilikan / permission user:update:any.
	if !CanAccessUser(*current, id, s.perms, "user:update:any") {
		return helper.Fail(c, fiber.StatusForbidden,
			"tidak berhak mengubah data user lain")
	}

	var req model.PatchUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}

	if req.Username == nil && req.Email == nil && req.IsActive == nil {
		return helper.Fail(c, fiber.StatusBadRequest,
			"tidak ada field yang diubah")
	}

	// Ambil data saat ini, lalu terapkan perubahan.
	saatIni, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateUserError(c, err, "gagal mengambil data user")
	}

	errs := map[string]string{}
	if req.Username != nil {
		if strings.TrimSpace(*req.Username) == "" {
			errs["username"] = "username tidak boleh kosong"
		} else {
			saatIni.Username = *req.Username
		}
	}
	if req.Email != nil {
		if strings.TrimSpace(*req.Email) == "" {
			errs["email"] = "email tidak boleh kosong"
		} else {
			saatIni.Email = *req.Email
		}
	}
	if req.IsActive != nil {
		saatIni.IsActive = *req.IsActive
	}
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	result, err := s.repo.Update(ctx, saatIni)
	if err != nil {
		return translateUserError(c, err, "gagal memperbarui user")
	}

	return helper.Ok(c, "data user berhasil diperbarui (PATCH)", result)
}

// ==================== DELETE /api/v1/users/:id ====================

func (s *UserService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current := helper.CurrentUser(c)
	if current == nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	// Punya permission menghapus TIDAK berarti boleh menghapus dirinya sendiri.
	// Tanpa aturan ini, admin bisa menghapus dirinya sendiri dan
	// sistem kehilangan satu-satunya admin.
	if current.ID == id {
		return helper.Fail(c, fiber.StatusForbidden,
			"tidak boleh menghapus akun sendiri")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return translateUserError(c, err, "gagal menghapus user")
	}

	return helper.NoContent(c)
}

// ==================== PATCH /api/v1/users/:id/role ====================

func (s *UserService) AssignRole(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current := helper.CurrentUser(c)
	if current == nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	var req model.AssignRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}

	if errs := ValidateAssignRole(*current, id, req, s.perms); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	result, err := s.repo.UpdateRole(ctx, id, strings.TrimSpace(req.Role))
	if err != nil {
		return translateUserError(c, err, "gagal mengubah role user")
	}

	return helper.Ok(c, "role user berhasil diubah", result)
}
