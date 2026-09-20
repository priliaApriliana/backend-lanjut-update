package config

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"api-students/helper"
	"api-students/middleware"
	"api-students/route"
)

// NewApp merakit aplikasi: membuat instance Fiber, memasang middleware,
// lalu mendaftarkan route.
// Modul 6: parameter disederhanakan menjadi Dependencies.
func NewApp(
	logger *slog.Logger,
	deps route.Dependencies,
) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "API Students - Modul 6",
		ErrorHandler: newErrorHandler(),
		BodyLimit:    1 * 1024 * 1024, // 1 MB
	})

	// CORS: origin dibatasi via environment variable.
	allowedOrigins := GetEnv("ALLOWED_ORIGINS", "http://localhost:3000")
	middleware.Register(app, logger, allowedOrigins)

	route.Register(app, deps)

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