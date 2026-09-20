package service

import (
	"api-students/app/model"
	"api-students/helper"
)

// CanAccessStudent memutuskan apakah seseorang boleh menyentuh data student.
//
// Dua jalur yang diizinkan:
//  1. Kepemilikan — owner_id data itu sama dengan id pemanggil.
//  2. Permission — role-nya berhak atas data siapa pun (anyPermission).
//
// Fungsi ini murni: tidak mengimpor fiber maupun repository, sehingga
// dapat diuji hanya dengan memanggilnya. ownerID bertipe int (bukan
// pointer) karena kolom owner_id bersifat NOT NULL.
func CanAccessStudent(
	current model.AuthUser,
	ownerID int,
	perms *helper.PermissionSet,
	anyPermission string,
) bool {
	if current.ID == ownerID {
		return true
	}
	return perms.Can(current.Role, anyPermission)
}