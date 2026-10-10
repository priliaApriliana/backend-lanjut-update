package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

type UserService struct {
	repo  repository.UserRepository
	perms *helper.PermissionSet
}

func NewUserService(
	repo repository.UserRepository,
	perms *helper.PermissionSet,
) *UserService {
	return &UserService{repo: repo, perms: perms}
}

// TranslateError mengubah error milik repository menjadi AppError.
func TranslateError(err error, entity string) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound(entity + " tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		if entity == "user" {
			return helper.Conflict("username sudah dipakai")
		}
		return helper.Conflict(entity + " sudah terdaftar")
	default:
		return helper.Internal(err)
	}
}

// ---------- GET /users ----------
// Dijaga middleware RequirePermission(perms, "user:list")
func (s *UserService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	// Format dipilih SEBELUM query dijalankan. Bila client meminta format
	// yang tidak dapat kita hasilkan, tidak ada gunanya membebani database
	// untuk hasil yang akan dibuang.
	format, err := helper.Negotiate(c, helper.FormatJSON, helper.FormatCSV)
	if err != nil {
		return err
	}

	q, err := helper.ParseCursorQuery(c)
	if err != nil {
		return err
	}

	rows, err := s.repo.FindAfterCursor(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}

	if format == helper.FormatCSV {
		return helper.WriteUsersCSV(c, rows)
	}

	// Baris tambahan hasil limit+1 dipotong di sini. Ia hanya penanda bahwa
	// masih ada halaman berikutnya, bukan bagian dari halaman ini.
	hasMore := len(rows) > q.Limit
	if hasMore {
		rows = rows[:q.Limit]
	}

	meta := &model.CursorMeta{Limit: q.Limit, HasMore: hasMore}
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		meta.NextCursor = helper.EncodeCursor(last.CreatedAt, last.ID)
	}

	return helper.SuccessCursor(c, "daftar user berhasil diambil", rows, meta)
}

// ---------- GET /users/:id ----------
func (s *UserService) Get(c *fiber.Ctx) error {
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

	if !CanAccessUser(current, id, s.perms, "user:read:any") {
		return helper.Forbidden("tidak berhak mengakses data user lain")
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return TranslateError(err, "user")
	}

	return helper.Success(c, fiber.StatusOK, "user ditemukan", user)
}

// ---------- POST /users ----------
// Dijaga middleware RequirePermission(perms, "user:update:any")
func (s *UserService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	hashed, err := helper.HashPassword(req.Password)
	if err != nil {
		return helper.Internal(err)
	}

	user, err := s.repo.Create(ctx, model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashed,
		Role:     "user",
		IsActive: true,
	})
	if err != nil {
		return TranslateError(err, "user")
	}

	return helper.Created(c, "user berhasil dibuat", user, fmt.Sprintf("/api/v1/users/%d", user.ID))
}

// ---------- PUT /users/:id ----------
func (s *UserService) Replace(c *fiber.Ctx) error {
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

	if !CanAccessUser(current, id, s.perms, "user:update:any") {
		return helper.Forbidden("tidak berhak mengubah data user lain")
	}

	var req model.ReplaceUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return TranslateError(err, "user")
	}

	existing.Username = strings.TrimSpace(req.Username)
	existing.Email = strings.TrimSpace(req.Email)
	existing.IsActive = req.IsActive

	updated, err := s.repo.Update(ctx, existing)
	if err != nil {
		return TranslateError(err, "user")
	}

	return helper.Success(c, fiber.StatusOK, "user berhasil diperbarui", updated)
}

// ---------- PATCH /users/:id ----------
func (s *UserService) Patch(c *fiber.Ctx) error {
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

	if !CanAccessUser(current, id, s.perms, "user:update:any") {
		return helper.Forbidden("tidak berhak mengubah data user lain")
	}

	var req model.PatchUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if IsEmptyUserPatch(req) {
		return helper.BadRequest("setidaknya satu field harus dikirim")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return TranslateError(err, "user")
	}

	existing = ApplyUserPatch(existing, req)

	updated, err := s.repo.Update(ctx, existing)
	if err != nil {
		return TranslateError(err, "user")
	}

	return helper.Success(c, fiber.StatusOK, "user berhasil diubah", updated)
}

// ---------- PATCH /users/:id/role ----------
// Dijaga middleware RequirePermission(perms, "role:assign")
func (s *UserService) AssignRole(c *fiber.Ctx) error {
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

	var req model.AssignRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := ValidateAssignRole(current, id, req, s.perms); len(errs) > 0 {
		return helper.Validation(errs)
	}

	result, err := s.repo.UpdateRole(ctx, id, strings.TrimSpace(req.Role))
	if err != nil {
		return TranslateError(err, "user")
	}

	return helper.Success(c, fiber.StatusOK, "role user berhasil diubah", result)
}

// ---------- DELETE /users/:id ----------
// Dijaga middleware RequirePermission(perms, "user:delete")
func (s *UserService) Delete(c *fiber.Ctx) error {
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

	// Punya permission menghapus tidak berarti boleh menghapus akun sendiri.
	if current.UserID == id {
		return helper.Forbidden("tidak boleh menghapus akun sendiri")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return TranslateError(err, "user")
	}

	return helper.NoContent(c)
}
