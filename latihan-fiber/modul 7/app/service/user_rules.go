package service

import (
	"strings"

	"api-students/app/model"
)

// ApplyPatch tidak lagi mengembalikan daftar error. Pemeriksaan bentuk
// sudah selesai dikerjakan tag sebelum fungsi ini dipanggil, sehingga di
// sini tugasnya tinggal satu: menggabungkan.
func ApplyUserPatch(current model.User, req model.PatchUserRequest) model.User {
	if req.Username != nil {
		current.Username = strings.TrimSpace(*req.Username)
	}
	if req.Email != nil {
		current.Email = strings.TrimSpace(*req.Email)
	}
	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}
	return current
}

func IsEmptyUserPatch(req model.PatchUserRequest) bool {
	return req.Username == nil && req.Email == nil && req.IsActive == nil
}
