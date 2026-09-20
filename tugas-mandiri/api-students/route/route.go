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

// Dependencies menampung semua yang dibutuhkan route.
// Siapa pun yang ingin tahu "endpoint mana yang dijaga apa" cukup
// membuka satu file ini.
type Dependencies struct {
	Pool           *pgxpool.Pool
	JWT            *helper.JWTManager
	Permissions    *helper.PermissionSet
	StudentService *service.StudentService
	AuthService    *service.AuthService
	UserService    *service.UserService
}

// Register memetakan URL ke method pada service. Tidak ada logika bisnis,
// tidak ada query, tidak ada validasi di sini.
func Register(app *fiber.App, deps Dependencies) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello World")
	})

	api := app.Group("/api/v1")
	api.Get("/health", healthCheck(deps.Pool))

	// ==================== AUTH (publik) ====================
	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)

	// /auth/me memerlukan token.
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	// ==================== USERS (wajib login) ====================
	users := api.Group("/users",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))

	perms := deps.Permissions

	// Hak dapat diputuskan TANPA melihat data → middleware.
	users.Get("/",
		middleware.RequirePermission(perms, "user:list"),
		deps.UserService.List)
	users.Delete("/:id",
		middleware.RequirePermission(perms, "user:delete"),
		deps.UserService.Delete)
	users.Patch("/:id/role",
		middleware.RequirePermission(perms, "role:assign"),
		deps.UserService.AssignRole)

	// Hak BERGANTUNG pada kepemilikan data → diperiksa di service.
	// Tidak ada RequirePermission di sini bukan berarti tidak dijaga;
	// pemeriksaan terjadi SETELAH id target diketahui, yaitu di layer
	// service melalui fungsi CanAccessUser.
	users.Get("/:id", deps.UserService.Get)
	users.Put("/:id", deps.UserService.Replace)
	users.Patch("/:id", deps.UserService.Patch)

	// ==================== STUDENTS (dilindungi) ====================
	// Semua endpoint students memerlukan token yang valid.
	s := api.Group("/students", middleware.RequireAuth(deps.JWT), middleware.RequireJSON)

	// Hak dapat diputuskan TANPA melihat data → middleware.
	s.Get("",
		middleware.RequirePermission(perms, "student:list"),
		deps.StudentService.List)
	s.Post("/",
		middleware.RequirePermission(perms, "student:create"),
		deps.StudentService.Create)
	s.Delete("/:id",
		middleware.RequirePermission(perms, "student:delete"),
		deps.StudentService.Delete)

	// Hak BERGANTUNG pada kepemilikan data → diperiksa di service
	// melalui fungsi CanAccessStudent.
	s.Get("/:id", deps.StudentService.Get)
	s.Put("/:id", deps.StudentService.Replace)
	s.Patch("/:id", deps.StudentService.Patch)
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
