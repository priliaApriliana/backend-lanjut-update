package model

import (
	"fmt"
	"time"
)

// Student merepresentasikan entitas inti domain mahasiswa (Layer 1: Entities).
// Entitas ini independen dari framework, database, dan HTTP.
type Student struct {
	ID        int       `json:"id"`
	NIM       string    `json:"nim"`
	Name      string    `json:"name"`
	Grade     float64   `json:"grade"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// GetInfo mengembalikan ringkasan data mahasiswa (Value Receiver, read-only).
func (s Student) GetInfo() string {
	statusStr := "Aktif"
	if !s.IsActive {
		statusStr = "Tidak Aktif"
	}
	return fmt.Sprintf("ID: %d | NIM: %s | Nama: %-15s | Grade: %.2f | Status: %s",
		s.ID, s.NIM, s.Name, s.Grade, statusStr)
}

// UpdateGrade memutasi nilai mahasiswa (Pointer Receiver).
func (s *Student) UpdateGrade(grade float64) {
	s.Grade = grade
}

// Activate mengaktifkan status mahasiswa (Pointer Receiver).
func (s *Student) Activate() {
	s.IsActive = true
}

// Deactivate menonaktifkan status mahasiswa (Pointer Receiver).
func (s *Student) Deactivate() {
	s.IsActive = false
}

// CreateStudentRequest adalah DTO request body penambahan data (POST).
type CreateStudentRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// ReplaceStudentRequest adalah DTO request body penggantian seluruh data (PUT).
type ReplaceStudentRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// PatchStudentRequest adalah DTO request body pembaruan sebagian (PATCH).
// Menggunakan pointer untuk membedakan field tidak dikirim (nil) vs zero-value.
type PatchStudentRequest struct {
	NIM      *string  `json:"nim,omitempty"`
	Name     *string  `json:"name,omitempty"`
	Grade    *float64 `json:"grade,omitempty"`
	IsActive *bool    `json:"is_active,omitempty"`
}

// WebResponse adalah amplop baku seluruh respons JSON API.
type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

// Meta menyediakan metadata paginasi.
type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ListStudentQuery menampung parameter penyaringan, pengurutan, dan paginasi.
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

// Offset menghitung baris yang dilewati untuk klausa OFFSET pada query basis data.
func (q ListStudentQuery) Offset() int {
	return (q.Page - 1) * q.Limit
}
