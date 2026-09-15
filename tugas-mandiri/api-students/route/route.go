package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/service"
	"api-students/helper"
	"api-students/middleware"
)

// Register memetakan URL ke method pada service. Tidak ada logika bisnis,
// tidak ada query, tidak ada validasi di sini.
// Modul 5: tambah authService, authMiddleware; kunci /students.
func Register(
	app *fiber.App,
	pool *pgxpool.Pool,
	studentService *service.StudentService,
	authService *service.AuthService,
	authMiddleware fiber.Handler,
) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello World")
	})

	api := app.Group("/api/v1")
	api.Get("/health", healthCheck(pool))

	// ==================== AUTH (publik) ====================
	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", authService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), authService.Login)
	auth.Post("/refresh", authService.Refresh)
	auth.Post("/logout", authService.Logout)

	// /auth/me memerlukan token.
	auth.Get("/me", authMiddleware, authService.Me)

	// ==================== STUDENTS (dilindungi) ====================
	// Semua endpoint students memerlukan token yang valid.
	s := api.Group("/students", authMiddleware, middleware.RequireJSON)
	s.Get("", studentService.List)
	s.Get("/:id", studentService.Get)
	s.Post("/", studentService.Create)
	s.Put("/:id", studentService.Replace)
	s.Patch("/:id", studentService.Patch)
	s.Delete("/:id", studentService.Delete)
}

// healthCheck melaporkan kondisi layanan beserta databasenya.
func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.Fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dihubungi")
		}
		return helper.Ok(c, "server berjalan", fiber.Map{"status": "ok"})
	}
}
