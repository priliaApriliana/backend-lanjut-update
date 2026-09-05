package main

import "time"

type Student struct {
	ID 			int 		`json:"id"`
	NIM 		string 		`json:"nim"`	
	Name 		string 		`json:"name"`
	Grade		float64 	`json:"grade"`
	IsActive 	bool 		`json:"is_active"`
	CreatedAt 	time.Time 	`json:"created_at"`
}

// POST, semua field wajib
type CreateStudentRequest struct {
	NIM 		string 		`json:"nim"`	
	Name 		string 		`json:"name"`
	Grade		float64 	`json:"grade"`
	IsActive 	bool 		`json:"is_active"`
}

// PUT, ganti semua isi/body, tipe biasa dan semuanya wajib
type UpdateStudentRequest struct {
	NIM 		string 		`json:"nim"`	
	Name 		string 		`json:"name"`
	Grade		float64 	`json:"grade"`
	IsActive 	bool 		`json:"is_active"`
}

// PATCH, ubah sbgaian, tipe pointer 
type PatchStudentRequest struct {
	NIM 		*string 	`json:"nim,omitempty"`
	Name 		*string 	`json:"name,omitempty"`
	Grade		*float64 	`json:"grade,omitempty"`
	IsActive 	*bool 		`json:"is_active,omitempty"`
}