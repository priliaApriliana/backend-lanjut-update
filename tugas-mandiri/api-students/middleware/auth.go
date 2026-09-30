package middleware

import (
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-students/helper"
)

// wwwAuth adalah header standar yang dikirim bersama setiap 401.
const wwwAuth = `Bearer realm="api"`

// RequireAuth membuat middleware yang memverifikasi access token JWT.
// Token dibaca dari header Authorization: Bearer <token>.
func RequireAuth(jwtManager *helper.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			c.Set("WWW-Authenticate", wwwAuth)
			return helper.Unauthorized("token diperlukan untuk mengakses resource ini")
		}

		// Format harus "Bearer <token>".
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.Set("WWW-Authenticate", wwwAuth)
			return helper.Unauthorized("format Authorization header tidak valid")
		}

		tokenString := parts[1]
		if tokenString == "" {
			c.Set("WWW-Authenticate", wwwAuth)
			return helper.Unauthorized("token tidak boleh kosong")
		}

		// Parse dan validasi token.
		authUser, err := jwtManager.Parse(tokenString)
		if err != nil {
			c.Set("WWW-Authenticate", wwwAuth)

			// Bedakan pesan error: expired vs invalid.
			pesan := "token tidak valid"
			if err == helper.ErrExpiredToken {
				pesan = "token sudah kedaluwarsa"
			}
			return helper.Unauthorized(pesan)
		}

		// Simpan identitas ke Locals menggunakan key konstanta.
		c.Locals(helper.ContextKeyAuth, authUser)
		return c.Next()
	}
}

// ==================== Login Rate Limiter ====================

// rateLimitEntry menyimpan data rate limit per IP.
type rateLimitEntry struct {
	count     int
	expiresAt time.Time
}

// rateLimitStore adalah in-memory store untuk rate limiting.
// Untuk produksi, sebaiknya diganti dengan Redis.
type rateLimitStore struct {
	mu      sync.Mutex
	entries map[string]*rateLimitEntry
}

var loginLimiter = &rateLimitStore{
	entries: make(map[string]*rateLimitEntry),
}

// LoginRateLimiter membatasi percobaan login: maksimal 5 per menit per IP.
// Request ke-6 dan seterusnya mendapat respons 429 Too Many Requests
// dengan header Retry-After: 60.
func LoginRateLimiter() fiber.Handler {
	// Goroutine pembersih: hapus entry yang sudah expired setiap 5 menit.
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			loginLimiter.mu.Lock()
			now := time.Now()
			for ip, entry := range loginLimiter.entries {
				if now.After(entry.expiresAt) {
					delete(loginLimiter.entries, ip)
				}
			}
			loginLimiter.mu.Unlock()
		}
	}()

	return func(c *fiber.Ctx) error {
		ip := c.IP()
		now := time.Now()

		loginLimiter.mu.Lock()
		entry, exists := loginLimiter.entries[ip]

		if !exists || now.After(entry.expiresAt) {
			// Entry baru atau sudah expired: reset.
			loginLimiter.entries[ip] = &rateLimitEntry{
				count:     1,
				expiresAt: now.Add(1 * time.Minute),
			}
			loginLimiter.mu.Unlock()
			return c.Next()
		}

		entry.count++
		count := entry.count
		loginLimiter.mu.Unlock()

		if count > 5 {
			c.Set("Retry-After", "60")
			return helper.TooManyRequests("terlalu banyak percobaan login, coba lagi dalam 60 detik")
		}

		return c.Next()
	}
}
