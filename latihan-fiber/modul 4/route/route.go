package route

import (
	"context"
	"time"

	"api-students/app/service"
	"api-students/helper"
	"api-students/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Register memetakan endpoint URL ke method pada service.
// File ini murni pendaftaran rute: tidak ada business logic, validasi input, atau query database.
func Register(app *fiber.App, pool *pgxpool.Pool, studentService *service.StudentService) {
	api := app.Group("/api/v1")

	// Endpoint health check yang turut memeriksa ketersediaan koneksi basis data
	api.Get("/health", healthCheck(pool))

	// Grup endpoint students dengan proteksi header Content-Type application/json (415)
	students := api.Group("/students", middleware.RequireJSON)
	students.Get("/", studentService.List)
	students.Get("/:id", studentService.Get)
	students.Post("/", studentService.Create)
	students.Put("/:id", studentService.Replace)
	students.Patch("/:id", studentService.Patch)
	students.Delete("/:id", studentService.Delete)
}

// healthCheck melaporkan kondisi kesehatan layanan dan koneksi basis data PostgreSQL.
func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.Fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dihubungi")
		}

		return helper.Success(c, fiber.StatusOK, "server dan basis data berjalan normal", fiber.Map{
			"status":   "healthy",
			"database": "connected",
			"service":  "api-students",
		})
	}
}
