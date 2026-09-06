package main

import (
	"context"
	"fmt"
	"log"

	"api-students/app/repository"
	"api-students/app/service"
	"api-students/config"
	"api-students/database"
)

// main hanya berisi urutan perakitan. Tidak ada logika bisnis,
// tidak ada query, dan tidak ada satu pun handler di sini.
func main() {
	// 1. Konfigurasi dan logger
	config.LoadEnv()
	logger := config.NewLogger()

	// 2. Koneksi basis data
	pool, err := database.NewPool(context.Background())
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	// 3. Perakitan: pool -> repository -> service
	studentRepository := repository.NewStudentRepository(pool)
	studentService := service.NewStudentService(studentRepository)

	// 4. Aplikasi
	app := config.NewApp(logger, pool, studentService)

	fmt.Println("Server berjalan di http://localhost:3000")
	port := config.GetEnv("APP_PORT", "3000")
	log.Fatal(app.Listen(":" + port))
}