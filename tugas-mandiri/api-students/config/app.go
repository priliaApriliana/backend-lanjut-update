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
// lalu mendaftarkan route. Sama persis dengan yang dulu tertulis langsung
// di main.go Modul 3.
func NewApp(
	logger *slog.Logger, pool *pgxpool.Pool, studentService *service.StudentService,
) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "API Students - Modul 4",
		ErrorHandler: newErrorHandler(),
	})

	middleware.Register(app, logger)
	route.Register(app, pool, studentService)

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