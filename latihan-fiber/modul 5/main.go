package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-students/app/repository"
	"api-students/app/service"
	"api-students/config"
	"api-students/database"
	"api-students/helper"
	"api-students/route"
)

const minSecretLength = 32

func main() {
	// 1. Muat konfigurasi environment & logger terstruktur
	config.LoadEnv()
	logger := config.NewLogger()

	// 2. Validasi rahasia JWT sebelum server menyala
	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if len(jwtSecret) < minSecretLength {
		logger.Error("JWT_SECRET tidak diisi atau terlalu pendek",
			slog.Int("minimal_karakter", minSecretLength))
		os.Exit(1)
	}

	// 3. Inisialisasi pool koneksi basis data PostgreSQL
	pool, err := database.NewPool(context.Background())
	if err != nil {
		logger.Error("gagal terhubung ke database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	// 4. Inisialisasi JWT Manager
	jwtManager := helper.NewJWTManager(
		jwtSecret,
		config.GetEnv("JWT_ISSUER", "praktikum-backend"),
		time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 15))*time.Minute,
	)

	// 5. Inisialisasi Repositories (Layer 3 - Interface Adapters)
	studentRepo := repository.NewStudentRepository(pool)
	userRepo := repository.NewUserRepository(pool)
	tokenRepo := repository.NewTokenRepository(pool)

	// 6. Inisialisasi Services (Layer 2 & 3 - Use Cases & Controllers)
	studentService := service.NewStudentService(studentRepo)
	authService := service.NewAuthService(
		userRepo,
		tokenRepo,
		jwtManager,
		time.Duration(config.GetEnvInt("JWT_REFRESH_TTL_DAYS", 7))*24*time.Hour,
	)

	// 7. Inisialisasi Fiber Application lengkap dengan Middleware & Routes
	app := config.NewApp(logger, route.Dependencies{
		Pool:           pool,
		JWT:            jwtManager,
		AuthService:    authService,
		StudentService: studentService,
	})

	// 8. Jalankan Server HTTP dalam goroutine terpisah
	port := config.GetEnv("APP_PORT", "3000")
	go func() {
		logger.Info("server berjalan", slog.String("port", port))
		if err := app.Listen(":" + port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server gagal berjalan", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// 9. Graceful Shutdown: Tangkap sinyal interupsi OS
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("sinyal shutdown diterima, menghentikan server secara graceful...")

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := app.ShutdownWithContext(ctxShutdown); err != nil {
		logger.Error("gagal mematikan server secara graceful", slog.String("error", err.Error()))
	} else {
		logger.Info("server HTTP berhasil dihentikan")
	}

	logger.Info("menutup koneksi pool basis data...")
	pool.Close()
	logger.Info("server berhasil dimatikan secara bersih")
	fmt.Println("Graceful shutdown selesai.")
}
