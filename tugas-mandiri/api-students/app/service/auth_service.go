package service

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

// AuthService menangani semua handler HTTP untuk autentikasi.
// Pola ini mengikuti StudentService: struct menyimpan dependency,
// method-method menjadi handler Fiber.
type AuthService struct {
	userRepo   repository.UserRepository
	tokenRepo  repository.TokenRepository
	jwt        *helper.JWTManager
	refreshTTL time.Duration
}

// NewAuthService merakit AuthService dengan semua dependency-nya.
func NewAuthService(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	jwt *helper.JWTManager,
	refreshTTLDays int,
) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		jwt:        jwt,
		refreshTTL: time.Duration(refreshTTLDays) * 24 * time.Hour,
	}
}

// ==================== POST /api/v1/auth/register ====================

func (s *AuthService) Register(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}

	if errs := ValidateRegister(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	// Hash password dengan bcrypt (cost 12).
	hashed, err := helper.HashPassword(req.Password)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memproses registrasi")
	}

	// Role SELALU "user" — tidak pernah diambil dari request.
	// Ini mencegah mass assignment: meskipun klien mengirim "role":"admin",
	// field tersebut tidak ada di RegisterRequest sehingga diabaikan.
	user, err := s.userRepo.Create(ctx, model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashed,
		Role:     "user",
		IsActive: true,
	})
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Fail(c, fiber.StatusConflict, "username atau email sudah digunakan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menyimpan user")
	}

	return helper.Created(c, "registrasi berhasil", user, "/api/v1/auth/me")
}

// ==================== POST /api/v1/auth/login ====================

func (s *AuthService) Login(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}

	if errs := ValidateLogin(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	// Pesan error yang SAMA PERSIS untuk kedua kasus:
	// 1. Username tidak ditemukan
	// 2. Password salah
	// Ini mencegah user enumeration.
	const pesanGagalLogin = "username atau password salah"

	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// User tidak ditemukan → tetap jalankan bcrypt dummy agar
			// waktu respons konsisten dengan kasus password salah.
			helper.VerifyDummyPassword(req.Password)
			return helper.Fail(c, fiber.StatusUnauthorized, pesanGagalLogin)
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memproses login")
	}

	// Verifikasi password.
	if !helper.VerifyPassword(user.Password, req.Password) {
		return helper.Fail(c, fiber.StatusUnauthorized, pesanGagalLogin)
	}

	// Generate token pair.
	pair, err := s.generateTokenPair(ctx, user)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal membuat token")
	}

	return helper.Ok(c, "login berhasil", pair)
}

// ==================== POST /api/v1/auth/refresh ====================

func (s *AuthService) Refresh(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}

	if req.RefreshToken == "" {
		return helper.Fail(c, fiber.StatusBadRequest, "refresh_token tidak boleh kosong")
	}

	// Hash token yang dikirim klien untuk mencocokkan dengan database.
	tokenHash := helper.SHA256Hex(req.RefreshToken)

	// Cari token aktif (belum expired, belum revoked).
	storedToken, err := s.tokenRepo.FindActive(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusUnauthorized, "refresh token tidak valid atau sudah digunakan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memverifikasi refresh token")
	}

	// ROTASI WAJIB: revoke token lama sebelum membuat yang baru.
	// Token yang sudah dipakai tidak boleh bisa dipakai lagi.
	if err := s.tokenRepo.Revoke(ctx, storedToken.ID); err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mencabut token lama")
	}

	// Ambil data user untuk claims token baru.
	user, err := s.userRepo.FindByID(ctx, storedToken.UserID)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data user")
	}

	// Generate token pair baru.
	pair, err := s.generateTokenPair(ctx, user)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal membuat token baru")
	}

	return helper.Ok(c, "token berhasil diperbarui", pair)
}

// ==================== POST /api/v1/auth/logout ====================

func (s *AuthService) Logout(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}

	if req.RefreshToken == "" {
		return helper.Fail(c, fiber.StatusBadRequest, "refresh_token tidak boleh kosong")
	}

	tokenHash := helper.SHA256Hex(req.RefreshToken)

	// Cari token untuk mendapatkan user_id.
	storedToken, err := s.tokenRepo.FindActive(ctx, tokenHash)
	if err != nil {
		// Meskipun token tidak ditemukan, tetap kembalikan 200
		// agar tidak bocorkan info apakah token itu pernah ada.
		return helper.Ok(c, "logout berhasil", nil)
	}

	// Cabut HANYA refresh token yang dikirim klien ini,
	// sesuai spesifikasi: satu logout = satu sesi berakhir.
	_ = s.tokenRepo.Revoke(ctx, storedToken.ID)

	return helper.Ok(c, "logout berhasil", nil)
}

// ==================== GET /api/v1/auth/me ====================

func (s *AuthService) Me(c *fiber.Ctx) error {
	authUser := helper.CurrentUser(c)
	if authUser == nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "tidak terautentikasi")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	user, err := s.userRepo.FindByID(ctx, authUser.ID)
	if err != nil {
		return helper.Fail(c, fiber.StatusNotFound, "user tidak ditemukan")
	}

	return helper.Ok(c, "profil pengguna", user)
}

// generateTokenPair membuat pasangan access token + refresh token.
func (s *AuthService) generateTokenPair(
	ctx context.Context,
	user model.User,
) (*model.TokenPair, error) {
	// 1. Access token (JWT).
	accessToken, err := s.jwt.GenerateAccess(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, err
	}

	// 2. Refresh token (random 32 byte, hex-encoded).
	rawRefresh, err := helper.RandomToken()
	if err != nil {
		return nil, err
	}

	// 3. Simpan HASH refresh token ke database, bukan nilai aslinya.
	rt := model.RefreshToken{
		UserID:    user.ID,
		TokenHash: helper.SHA256Hex(rawRefresh),
		ExpiresAt: time.Now().Add(s.refreshTTL),
	}
	if err := s.tokenRepo.Save(ctx, rt); err != nil {
		return nil, err
	}

	return &model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
	}, nil
}
