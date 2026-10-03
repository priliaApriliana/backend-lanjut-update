package main

import (
	"fmt"
	"time"
)

// Student merepresentasikan entitas mahasiswa pada sistem akademik REST API.
// Membawa field dari Modul 1 (ID, Name, Grade, IsActive) ditambah field unik NIM dan CreatedAt.
type Student struct {
	ID        int       `json:"id"`
	NIM       string    `json:"nim"`
	Name      string    `json:"name"`
	Grade     float64   `json:"grade"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// GetInfo mengembalikan informasi lengkap mahasiswa dalam satu baris teks ringkas.
// Value receiver: hanya membaca data (read-only) tanpa memodifikasi struct asli.
func (s Student) GetInfo() string {
	statusStr := "Aktif"
	if !s.IsActive {
		statusStr = "Tidak Aktif"
	}
	return fmt.Sprintf("ID: %d | NIM: %s | Nama: %-15s | Grade: %.2f | Status: %s",
		s.ID, s.NIM, s.Name, s.Grade, statusStr)
}

// UpdateGrade memperbarui nilai mahasiswa secara in-place.
// Pointer receiver: memutasi field Grade pada memori objek pemanggil.
func (s *Student) UpdateGrade(grade float64) {
	s.Grade = grade
}

// Activate mengubah status IsActive menjadi true.
// Pointer receiver: memutasi field IsActive pada memori objek asli.
func (s *Student) Activate() {
	s.IsActive = true
}

// Deactivate mengubah status IsActive menjadi false.
// Pointer receiver: memutasi field IsActive pada memori objek asli.
func (s *Student) Deactivate() {
	s.IsActive = false
}

// CreateStudentRequest adalah DTO request body untuk penambahan data (POST).
// Seluruh field wajib diisi dan divalidasi.
type CreateStudentRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// ReplaceStudentRequest adalah DTO request body untuk penggantian seluruh data (PUT).
// Karakteristik PUT adalah menimpa seluruh entitas; jika ada field yang tidak dikirim,
// maka dianggap tidak valid atau dikosongkan. Oleh karenanya semua field wajib dikirim.
type ReplaceStudentRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// PatchStudentRequest adalah DTO request body untuk pembaruan parsial (PATCH).
// Field menggunakan pointer (*string, *float64, *bool) agar server dapat membedakan
// antara "field tidak dikirim / tidak ingin diubah" (nil) dan "field dikirim bernilai zero/kosong".
type PatchStudentRequest struct {
	NIM      *string  `json:"nim,omitempty"`
	Name     *string  `json:"name,omitempty"`
	Grade    *float64 `json:"grade,omitempty"`
	IsActive *bool    `json:"is_active,omitempty"`
}

// WebResponse adalah format baku amplop respons JSON untuk semua endpoint API.
type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

// Meta menyediakan metadata paginasi pada endpoint list.
type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ListStudentQuery menampung parameter query string untuk penyaringan, pengurutan, dan paginasi.
type ListStudentQuery struct {
	Page     int
	Limit    int
	Search   string
	Sort     string
	Order    string
	IsActive *bool
	MinGrade *float64
	MaxGrade *float64
}
