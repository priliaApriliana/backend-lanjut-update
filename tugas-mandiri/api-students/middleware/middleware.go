package middleware

import (
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"api-students/helper"
)

// Register memasang middleware global, urutan sama persis dengan main.go
// Modul 3: requestid -> logger -> cors.
func Register(app *fiber.App, logger *slog.Logger) {
	app.Use(requestid.New())
	app.Use(RequestLogger(logger))
	app.Use(cors.New())
}

// RequestLogger menggantikan logger.New() bawaan Fiber. Sekarang mencatat
// ke logger terstruktur (JSON) sesuai poin 3 Tugas Mandiri.
func RequestLogger(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

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

var metodeBerbody = map[string]bool{
	fiber.MethodPost:  true,
	fiber.MethodPut:   true,
	fiber.MethodPatch: true,
}

// RequireJSON menolak request berisi body yang Content-Type-nya bukan
// JSON. Sama persis dengan requireJSON di main.go Modul 3.
func RequireJSON(c *fiber.Ctx) error {
	if metodeBerbody[c.Method()] {
		ct := c.Get("Content-Type")
		if !strings.HasPrefix(ct, fiber.MIMEApplicationJSON) {
			return helper.Fail(c, fiber.StatusUnsupportedMediaType, "Content-Type harus application/json")
		}
	}
	return c.Next()
}