package service

import (
	"api-students/app/model"
)

// File ini berisi business rules yang TIDAK dapat dinyatakan sebagai tag
// validasi — aturan yang berbicara tentang HUBUNGAN antar field, atau
// operasi murni tanpa keputusan valid/tidak valid.
//
// Validasi bentuk data (wajib diisi, panjang, format NIM) sudah
// dipindahkan ke tag pada struct request sejak Modul 7 (Langkah 6 & D.2).

// ApplyPatch menggabungkan perubahan ke entity yang sudah ada.
//
// Pemeriksaan bentuk (kosong, format NIM, rentang grade) sudah selesai
// dikerjakan tag sebelum fungsi ini dipanggil, sehingga di sini tugasnya
// tinggal satu: menggabungkan.
func ApplyPatch(current model.Student, req model.PatchStudentRequest) model.Student {
	if req.NIM != nil {
		current.NIM = *req.NIM
	}
	if req.Name != nil {
		current.Name = *req.Name
	}
	if req.Grade != nil {
		current.Grade = *req.Grade
	}
	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}
	return current
}

// IsEmptyPatch memeriksa body PATCH yang tidak berisi field apa pun.
//
// Aturan ini tidak dapat ditulis sebagai tag: tag memeriksa satu field
// pada satu waktu, sedangkan aturan ini berbicara tentang HUBUNGAN antar
// field — setidaknya satu di antaranya harus ada.
func IsEmptyPatch(req model.PatchStudentRequest) bool {
	return req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil
}

// CountTotalPages menghitung jumlah halaman untuk offset pagination.
// Tetap ada karena endpoint List students masih memakai offset — cursor
// pagination untuk students dikerjakan terpisah di D.3.
func CountTotalPages(total, limit int) int {
	if limit <= 0 {
		return 0
	}
	return (total + limit - 1) / limit
}