package model

import "time"

type Student struct {
    ID        int       `json:"id"`
    NIM       string    `json:"nim"`
    Name      string    `json:"name"`
    Grade     float64   `json:"grade"`
    IsActive  bool      `json:"is_active"`
    CreatedAt time.Time `json:"created_at"`
}

// POST, semua field wajib
type CreateStudentRequest struct {
    NIM      string  `json:"nim"`
    Name     string  `json:"name"`
    Grade    float64 `json:"grade"`
    IsActive bool    `json:"is_active"`
}

// PUT, ganti semua isi/body, tipe biasa dan semuanya wajib
type UpdateStudentRequest struct {
    NIM      string  `json:"nim"`
    Name     string  `json:"name"`
    Grade    float64 `json:"grade"`
    IsActive bool    `json:"is_active"`
}

// PATCH, ubah sebagian, tipe pointer
type PatchStudentRequest struct {
    NIM      *string  `json:"nim,omitempty"`
    Name     *string  `json:"name,omitempty"`
    Grade    *float64 `json:"grade,omitempty"`
    IsActive *bool    `json:"is_active,omitempty"`
}

// ListQuery menampung parameter query untuk endpoint List.
// Dulu (pertemuan 2) semua ini cuma variabel lokal di handler;
// sekarang jadi struct karena harus diteruskan ke repository.
type ListQuery struct {
    Page     int
    Limit    int
    Search   string
    Sort     string
    Order    string
    IsActive *bool
    Grade    string
}

// Offset menghitung berapa baris yang dilewati untuk halaman ini.
// Pindah ke sini karena kini dipakai langsung oleh SQL (LIMIT/OFFSET).
func (q ListQuery) Offset() int {
    return (q.Page - 1) * q.Limit
}
