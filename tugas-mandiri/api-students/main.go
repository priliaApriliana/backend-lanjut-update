package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"api-students/app/repository"
	"api-students/app/service"
	"api-students/config"
	"api-students/database"
	"api-students/helper"
	"api-students/route"
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

	// 4. Perakitan: pool -> repository
	studentRepository := repository.NewStudentRepository(pool)

	// 5. Perakitan autentikasi (Modul 5)
	userRepository := repository.NewUserRepository(pool)
	tokenRepository := repository.NewTokenRepository(pool)
	roleRepository := repository.NewRoleRepository(pool)

	// Pemetaan role ke permission dibaca SEKALI saat aplikasi menyala.
	// Konsekuensinya: perubahan hak akses di database baru berlaku setelah
	// aplikasi dijalankan ulang. Itu keputusan sadar, bukan kelalaian.
	rawPermissions, err := roleRepository.LoadPermissions(context.Background())
	if err != nil {
		logger.Error("gagal memuat permission", slog.String("error", err.Error()))
		os.Exit(1)
	}
	permissions := helper.NewPermissionSet(rawPermissions)
	logger.Info("permission dimuat", slog.Any("roles", permissions.KnownRoles()))

	jwtIssuer := config.GetEnv("JWT_ISSUER", "api-students")
	jwtAccessTTL := config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 15)
	jwtRefreshTTL := config.GetEnvInt("JWT_REFRESH_TTL_DAYS", 7)

	jwtManager := helper.NewJWTManager(jwtSecret, jwtIssuer, jwtAccessTTL)

	authService := service.NewAuthService(
		userRepository,
		tokenRepository,
		jwtManager,
		permissions,
		jwtRefreshTTL,
	)

	studentService := service.NewStudentService(studentRepository, permissions)
	userService := service.NewUserService(userRepository, permissions)

	// 6. Aplikasi — rakit Dependencies lalu serahkan ke NewApp.
	deps := route.Dependencies{
		Pool:           pool,
		JWT:            jwtManager,
		Permissions:    permissions,
		StudentService: studentService,
		AuthService:    authService,
		UserService:    userService,
	}
	app := config.NewApp(logger, deps)

	fmt.Println("Server berjalan di http://localhost:3000")
	port := config.GetEnv("APP_PORT", "3000")
	log.Fatal(app.Listen(":" + port))
}
