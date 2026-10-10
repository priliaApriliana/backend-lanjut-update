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

type Dependencies struct {
	Pool           *pgxpool.Pool
	JWT            *helper.JWTManager
	AuthService    *service.AuthService
	StudentService *service.StudentService
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	// --- 1. Endpoint Publik ---
	api.Get("/health", healthCheck(deps.Pool))

	// --- 2. Endpoint Autentikasi ---
	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	// --- 3. Endpoint Students (Wajib Membawa Access Token) ---
	students := api.Group("/students", middleware.RequireAuth(deps.JWT))
	students.Get("/", deps.StudentService.List)
	students.Get("/:id", deps.StudentService.Get)
	students.Post("/", middleware.RequireJSON, deps.StudentService.Create)
	students.Put("/:id", middleware.RequireJSON, deps.StudentService.Replace)
	students.Patch("/:id", middleware.RequireJSON, deps.StudentService.Patch)
	students.Delete("/:id", deps.StudentService.Delete)

	// --- 4. Fallback Handler (404) ---
	app.Use(func(c *fiber.Ctx) error {
		return helper.Fail(c, fiber.StatusNotFound, "rute yang diminta tidak ditemukan di server")
	})
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		dbStatus := "connected"
		if err := pool.Ping(ctx); err != nil {
			dbStatus = "disconnected: " + err.Error()
			return helper.Fail(c, fiber.StatusServiceUnavailable, "database tidak tersedia: "+err.Error())
		}

		return helper.Success(c, fiber.StatusOK, "layanan berjalan normal", fiber.Map{
			"status":    "healthy",
			"database":  dbStatus,
			"timestamp": time.Now(),
		})
	}
}
