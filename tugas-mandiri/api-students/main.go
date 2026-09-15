package main

import (
	"context"
	"fmt"
	"log"

	"api-students/app/repository"
	"api-students/app/service"
	"api-students/config"
	"api-students/database"
	"api-students/helper"
)

// main hanya berisi urutan perakitan. Tidak ada logika bisnis,
// tidak ada query, dan tidak ada satu pun handler di sini.
func main() {
	// 1. Konfigurasi dan logger
	config.LoadEnv()
	logger := config.NewLogger()

	// 2. Validasi JWT_SECRET — minimal 32 karakter.
	// Jika kurang, server TIDAK BOLEH berjalan karena token yang dihasilkan
	// akan mudah ditebak dengan brute force.
	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if len(jwtSecret) < 32 {
		log.Fatal("JWT_SECRET harus diisi dan minimal 32 karakter")
	}

	// 3. Koneksi basis data
	pool, err := database.NewPool(context.Background())
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	// 4. Perakitan: pool -> repository -> service
	studentRepository := repository.NewStudentRepository(pool)
	studentService := service.NewStudentService(studentRepository)

	// 5. Perakitan autentikasi (Modul 5)
	userRepository := repository.NewUserRepository(pool)
	tokenRepository := repository.NewTokenRepository(pool)

	jwtIssuer := config.GetEnv("JWT_ISSUER", "api-students")
	jwtAccessTTL := config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 15)
	jwtRefreshTTL := config.GetEnvInt("JWT_REFRESH_TTL_DAYS", 7)

	jwtManager := helper.NewJWTManager(jwtSecret, jwtIssuer, jwtAccessTTL)

	authService := service.NewAuthService(
		userRepository,
		tokenRepository,
		jwtManager,
		jwtRefreshTTL,
	)

	// 6. Aplikasi
	app := config.NewApp(logger, pool, studentService, authService, jwtManager)

	fmt.Println("Server berjalan di http://localhost:3000")
	port := config.GetEnv("APP_PORT", "3000")
	log.Fatal(app.Listen(":" + port))
}