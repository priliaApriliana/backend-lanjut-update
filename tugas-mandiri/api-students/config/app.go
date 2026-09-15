package config

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/service"
	"api-students/helper"
	"api-students/middleware"
	"api-students/route"
)

// NewApp merakit aplikasi: membuat instance Fiber, memasang middleware,
// lalu mendaftarkan route.
// Modul 5: tambah authService, jwtManager; BodyLimit 1 MB.
func NewApp(
	logger *slog.Logger,
	pool *pgxpool.Pool,
	studentService *service.StudentService,
	authService *service.AuthService,
	jwtManager *helper.JWTManager,
) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "API Students - Modul 5",
		ErrorHandler: newErrorHandler(),
		BodyLimit:    1 * 1024 * 1024, // 1 MB
	})

	// CORS: origin dibatasi via environment variable.
	allowedOrigins := GetEnv("ALLOWED_ORIGINS", "http://localhost:3000")
	middleware.Register(app, logger, allowedOrigins)

	// Middleware auth: dipakai oleh route yang memerlukan token.
	authMiddleware := middleware.RequireAuth(jwtManager)

	route.Register(app, pool, studentService, authService, authMiddleware)

	// Penampung terakhir untuk URL yang tidak dikenal.
	app.Use(func(c *fiber.Ctx) error {
		return helper.Fail(c, fiber.StatusNotFound, "endpoint tidak ditemukan")
	})

	return app
}

// newErrorHandler sama persis dengan ErrorHandler di main.go Modul 3.
func newErrorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		status := fiber.StatusInternalServerError
		pesan := "terjadi kesalahan pada server"
		if e, ok := err.(*fiber.Error); ok {
			status = e.Code
			pesan = e.Message
		}
		return helper.Fail(c, status, pesan)
	}
}