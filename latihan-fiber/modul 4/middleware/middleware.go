package middleware

import (
	"log/slog"
	"strings"
	"time"

	"api-students/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

// Register memasang seluruh middleware global yang berlaku untuk seluruh route.
// URUTAN PENTING: dieksekusi secara berurutan sesuai pemasangan.
func Register(app *fiber.App, logger *slog.Logger) {
	app.Use(requestid.New())       // 1. Generate unique request ID untuk pelacakan
	app.Use(recover.New())         // 2. Tangkap panic agar proses server tidak crash
	app.Use(helmet.New())          // 3. Pasang header proteksi keamanan HTTP dasar
	app.Use(cors.New())            // 4. Konfigurasi Cross-Origin Resource Sharing
	app.Use(RequestLogger(logger)) // 5. Catat setiap request ke structured logger
}

// RequestLogger mencatat setiap request HTTP ke logger terstruktur (stdout & logs/app.log).
func RequestLogger(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next() // Eksekusi handler berikutnya

		requestID, _ := c.Locals("requestid").(string)

		logger.Info("http_request",
			slog.String("request_id", requestID),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", c.Response().StatusCode()),
			slog.Duration("duration", time.Since(start)),
			slog.String("ip", c.IP()),
		)

		return err
	}
}

var methodsWithBody = map[string]bool{
	fiber.MethodPost:  true,
	fiber.MethodPut:   true,
	fiber.MethodPatch: true,
}

// RequireJSON menolak request ber-body yang Content-Type-nya bukan JSON dengan status 415.
func RequireJSON(c *fiber.Ctx) error {
	if methodsWithBody[c.Method()] {
		ct := c.Get("Content-Type")
		if !strings.HasPrefix(ct, fiber.MIMEApplicationJSON) {
			return helper.Fail(c, fiber.StatusUnsupportedMediaType, "Content-Type harus application/json")
		}
	}
	return c.Next()
}
