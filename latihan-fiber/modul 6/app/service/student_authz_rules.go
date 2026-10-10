package service

import (
	"api-students/app/model"
	"api-students/helper"
)

// CanAccessStudent memutuskan apakah seseorang berhak mengakses/mengubah data mahasiswa.
//
// Aturan Akses:
// 1. Kepemilikan (ownership): jika current.UserID == ownerID, akses selalu diizinkan.
// 2. Hak Istimewa (Permission): jika bukan pemilik, peran (role) pemanggil harus
//    memiliki izin scope global yang sesuai (anyPermission, misal "student:read:any" atau "student:update:any").
//
// Fungsi ini bersifat murni (pure function) dan tidak bergantung pada Fiber maupun database repository.
func CanAccessStudent(
	current model.AuthUser,
	ownerID int,
	perms *helper.PermissionSet,
	anyPermission string,
) bool {
	// Jalur 1: Pemilik data selalu memiliki hak akses penuh atas datanya sendiri
	if current.UserID == ownerID {
		return true
	}

	// Jalur 2: Pemeriksaan hak akses perizinan berbasis role (RBAC)
	if perms == nil {
		return false // Fail closed
	}

	return perms.Can(current.Role, anyPermission)
}
