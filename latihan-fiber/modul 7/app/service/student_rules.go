package service

import (
	"strings"

	"api-students/app/model"
)

// ApplyStudentPatch menerapkan pembaruan parsial pada entitas mahasiswa (PATCH).
// Field yang bernilai nil dibiarkan apa adanya (tidak dimutasi).
// Pemeriksaan bentuk sudah selesai dikerjakan tag sebelum fungsi ini dipanggil,
// sehingga di sini tugasnya tinggal satu: menggabungkan data.
func ApplyStudentPatch(
	current model.Student, req model.PatchStudentRequest,
) model.Student {
	if req.NIM != nil {
		current.NIM = strings.TrimSpace(*req.NIM)
	}

	if req.Name != nil {
		current.Name = strings.TrimSpace(*req.Name)
	}

	if req.Grade != nil {
		current.UpdateGrade(*req.Grade)
	}

	if req.IsActive != nil {
		if *req.IsActive {
			current.Activate()
		} else {
			current.Deactivate()
		}
	}

	return current
}

// IsEmptyStudentPatch memeriksa apakah payload PATCH tidak mengirimkan satu field pun.
//
// Aturan ini tidak dapat ditulis sebagai tag: tag memeriksa satu field
// pada satu waktu, sedangkan aturan ini berbicara tentang HUBUNGAN antar
// field — setidaknya satu di antara mereka harus ada.
func IsEmptyStudentPatch(req model.PatchStudentRequest) bool {
	return req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil
}

// CountTotalPages menghitung total halaman paginasi pembulatan ke atas tanpa float.
func CountTotalPages(total, limit int) int {
	if limit <= 0 || total <= 0 {
		return 0
	}
	return (total + limit - 1) / limit
}

// IsValidGrade memeriksa rentang nilai grade.
func IsValidGrade(grade float64) bool {
	return grade >= 0.0 && grade <= 100.0
}
