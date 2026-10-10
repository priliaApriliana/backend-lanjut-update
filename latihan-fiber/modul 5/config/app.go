package config

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"api-students/helper"
	"api-students/middleware"
	"api-students/route"
)

// NewApp membuat instance *fiber.App yang sudah dikonfigurasi lengkap dengan middleware dan route.
func NewApp(logger *slog.Logger, deps route.Dependencies) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      GetEnv("APP_NAME", "Praktikum Backend Lanjut"),
		ErrorHandler: newErrorHandler(logger),
		// Membatasi ukuran body maksimal 1 MB untuk mitigasi Denial-of-Service
		BodyLimit:    1 * 1024 * 1024, // 1 MB
	})

	// Pasang middleware global
	middleware.Register(app, logger, GetEnv("ALLOWED_ORIGINS", "http://localhost:5173"))

	// Daftarkan seluruh rute API
	route.Register(app, deps)

	return app
}

func newErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		logger.Error("terjadi kesalahan http",
			slog.Int("status", code),
			slog.String("error", err.Error()),
			slog.String("path", c.Path()),
			slog.String("method", c.Method()),
		)

		return helper.Fail(c, code, err.Error())
	}
}
