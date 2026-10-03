package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"api-students/app/repository"
	"api-students/config"
	"api-students/database"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

var metodeBerbody = map[string]bool{
	fiber.MethodPost:  true,
	fiber.MethodPut:   true,
	fiber.MethodPatch: true,
}

func requireJSON(c *fiber.Ctx) error {
	if metodeBerbody[c.Method()] {
		ct := c.Get("Content-Type")
		if !strings.HasPrefix(ct, fiber.MIMEApplicationJSON) {
			return fail(c, fiber.StatusUnsupportedMediaType, "Content-Type harus application/json")
		}
	}
	return c.Next()
}

func main() {
	// 1. Memuat variabel environment dari file .env
	config.LoadEnv()

	// 2. Inisialisasi PostgreSQL Connection Pool
	pool, err := database.NewPool(context.Background())
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	defer pool.Close()

	// 3. Perakitan Lapisan Arsitektur (Dependency Injection):
	// Connection Pool -> Repository -> Handler
	studentRepository := repository.NewStudentRepository(pool)
	studentHandler := NewStudentHandler(studentRepository)

	// 4. Konfigurasi Aplikasi Fiber
	app := fiber.New(fiber.Config{
		AppName: "Praktikum Backend Lanjut - Pertemuan 3 (API Students PostgreSQL)",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			status := fiber.StatusInternalServerError
			pesan := "terjadi kesalahan internal server"
			if e, ok := err.(*fiber.Error); ok {
				status = e.Code
				pesan = e.Message
			}
			return fail(c, status, pesan)
		},
	})

	// 5. Middleware Global
	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${locals:requestid} ${method} ${path} ${status} ${latency}\n",
	}))
	app.Use(cors.New())

	// 6. Rute Root & Health Check
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("API Students (PostgreSQL & Repository Pattern) - Praktikum Backend Lanjut Pertemuan 3")
	})

	api := app.Group("/api/v1")

	// Endpoint health check yang turut memverifikasi kesehatan koneksi basis data
	api.Get("/health", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dihubungi")
		}

		return ok(c, "server dan basis data berjalan normal", fiber.Map{
			"status":    "healthy",
			"database":  "connected",
			"service":   "api-students",
			"timestamp": time.Now(),
		})
	})

	// 7. Rute Sumber Daya Students
	s := api.Group("/students", requireJSON)
	s.Get("/", studentHandler.List)
	s.Get("/:id", studentHandler.Get)
	s.Post("/", studentHandler.Create)
	s.Put("/:id", studentHandler.Replace)
	s.Patch("/:id", studentHandler.Patch)
	s.Delete("/:id", studentHandler.Delete)

	// 8. Handler Fallback 404
	app.Use(func(c *fiber.Ctx) error {
		return fail(c, fiber.StatusNotFound, "endpoint tidak ditemukan")
	})

	// 9. Jalankan Server
	port := config.GetEnv("APP_PORT", "3001")
	fmt.Printf("Server API Students berjalan di http://localhost:%s\n", port)
	log.Fatal(app.Listen(":" + port))
}
