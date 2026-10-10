package service

import (
	"strings"

	"api-students/app/model"
)

// File ini berisi business rules MURNI (Layer 2: Use Cases):
// - Tidak mengimpor Fiber (*fiber.Ctx)
// - Tidak menyentuh basis data (SQL)
// - Tidak mengetahui detail transportasi HTTP apa pun
// - Dapat diuji secara terisolasi melalui unit test tanpa server atau database

// ValidateCreate memvalidasi payload penambahan mahasiswa baru (POST).
func ValidateCreate(req model.CreateStudentRequest) map[string]string {
	errs := map[string]string{}

	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "NIM wajib diisi"
	}
	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "nama mahasiswa wajib diisi"
	}
	if !IsValidGrade(req.Grade) {
		errs["grade"] = "nilai grade harus berada pada rentang 0.0 sampai 100.0"
	}

	return errs
}

// ValidateReplace memvalidasi payload penggantian seluruh data mahasiswa (PUT).
// Seluruh field wajib hadir dan memenuhi kriteria validasi.
func ValidateReplace(req model.ReplaceStudentRequest) map[string]string {
	errs := map[string]string{}

	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "NIM wajib diisi pada PUT"
	}
	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "nama mahasiswa wajib diisi pada PUT"
	}
	if !IsValidGrade(req.Grade) {
		errs["grade"] = "nilai grade harus berada pada rentang 0.0 sampai 100.0"
	}

	return errs
}

// ApplyPatch menerapkan pembaruan parsial pada entitas mahasiswa (PATCH).
// Field yang bernilai nil dibiarkan apa adanya (tidak dimutasi).
func ApplyPatch(
	current model.Student, req model.PatchStudentRequest,
) (model.Student, map[string]string) {
	errs := map[string]string{}

	if req.NIM != nil {
		trimmed := strings.TrimSpace(*req.NIM)
		if trimmed == "" {
			errs["nim"] = "NIM tidak boleh kosong"
		} else {
			current.NIM = trimmed
		}
	}

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			errs["name"] = "nama mahasiswa tidak boleh kosong"
		} else {
			current.Name = trimmed
		}
	}

	if req.Grade != nil {
		if !IsValidGrade(*req.Grade) {
			errs["grade"] = "nilai grade harus berada pada rentang 0.0 sampai 100.0"
		} else {
			current.UpdateGrade(*req.Grade) // Menggunakan method receiver domain
		}
	}

	if req.IsActive != nil {
		if *req.IsActive {
			current.Activate()
		} else {
			current.Deactivate()
		}
	}

	return current, errs
}

// IsEmptyPatch memeriksa apakah payload PATCH tidak mengirimkan satu field pun.
func IsEmptyPatch(req model.PatchStudentRequest) bool {
	return req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil
}

// CountTotalPages menghitung total halaman paginasi pembulatan ke atas tanpa float.
func CountTotalPages(total, limit int) int {
	if limit <= 0 || total <= 0 {
		return 0
	}
	return (total + limit - 1) / limit
}

// IsValidGrade memeriksa apakah nilai grade berada dalam rentang valid 0.0 s/d 100.0.
func IsValidGrade(grade float64) bool {
	return grade >= 0.0 && grade <= 100.0
}
