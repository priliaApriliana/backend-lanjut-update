package service_test

import (
	"testing"

	"api-students/app/model"
	"api-students/app/service"
	"api-students/helper"
)

func newTestPermissionSet() *helper.PermissionSet {
	return helper.NewPermissionSet(map[string][]string{
		"admin": {
			"user:list", "user:read:any", "user:update:any", "user:delete", "role:assign",
			"student:list", "student:read:any", "student:create", "student:update:any", "student:delete",
		},
		"staff": {
			"user:list", "user:read:any",
			"student:list", "student:read:any", "student:create",
		},
		"user": {},
	})
}

func TestCanAccessUser(t *testing.T) {
	perms := newTestPermissionSet()

	tests := []struct {
		name          string
		user          model.AuthUser
		targetID      int
		perm          string
		expectedGrant bool
	}{
		{
			name:          "Pemilik mengakses akunnya sendiri (Role user)",
			user:          model.AuthUser{UserID: 10, Username: "andi", Role: "user"},
			targetID:      10,
			perm:          "user:read:any",
			expectedGrant: true,
		},
		{
			name:          "Bukan pemilik tetapi memiliki permission (Role admin)",
			user:          model.AuthUser{UserID: 1, Username: "admin", Role: "admin"},
			targetID:      10,
			perm:          "user:read:any",
			expectedGrant: true,
		},
		{
			name:          "Bukan pemilik dan tidak memiliki permission (Role user)",
			user:          model.AuthUser{UserID: 5, Username: "budi", Role: "user"},
			targetID:      10,
			perm:          "user:read:any",
			expectedGrant: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.CanAccessUser(tt.user, tt.targetID, perms, tt.perm)
			if got != tt.expectedGrant {
				t.Errorf("CanAccessUser() = %v, want %v", got, tt.expectedGrant)
			}
		})
	}
}

func TestValidateAssignRole(t *testing.T) {
	perms := newTestPermissionSet()

	admin := model.AuthUser{UserID: 1, Username: "admin", Role: "admin"}

	t.Run("Assign Role Valid ke User Lain", func(t *testing.T) {
		errs := service.ValidateAssignRole(admin, 2, model.AssignRoleRequest{Role: "staff"}, perms)
		if len(errs) > 0 {
			t.Errorf("ValidateAssignRole menghasilkan error tak terduga: %v", errs)
		}
	})

	t.Run("Role Kosong", func(t *testing.T) {
		errs := service.ValidateAssignRole(admin, 2, model.AssignRoleRequest{Role: "  "}, perms)
		if errs["role"] != "wajib diisi" {
			t.Errorf("ValidateAssignRole error = %v, want 'wajib diisi'", errs["role"])
		}
	})

	t.Run("Role Tidak Dikenal", func(t *testing.T) {
		errs := service.ValidateAssignRole(admin, 2, model.AssignRoleRequest{Role: "superhero"}, perms)
		if errs["role"] == "" {
			t.Errorf("ValidateAssignRole diharapkan menolak role tidak dikenal")
		}
	})

	t.Run("Tidak Boleh Mengubah Role Diri Sendiri", func(t *testing.T) {
		errs := service.ValidateAssignRole(admin, 1, model.AssignRoleRequest{Role: "user"}, perms)
		if errs["role"] != "tidak boleh mengubah role diri sendiri" {
			t.Errorf("ValidateAssignRole error = %v, want 'tidak boleh mengubah role diri sendiri'", errs["role"])
		}
	})
}

func TestCanAccessStudent(t *testing.T) {
	perms := newTestPermissionSet()

	userOwner := model.AuthUser{UserID: 10, Username: "andi", Role: "user"}
	userOther := model.AuthUser{UserID: 11, Username: "budi", Role: "user"}
	staffUser := model.AuthUser{UserID: 2, Username: "sari", Role: "staff"}
	adminUser := model.AuthUser{UserID: 1, Username: "admin", Role: "admin"}

	studentOwnerID := 10

	tests := []struct {
		name          string
		user          model.AuthUser
		ownerID       int
		perm          string
		expectedGrant bool
	}{
		{
			name:          "Pemilik data mahasiswa selalu diizinkan (Role user)",
			user:          userOwner,
			ownerID:       studentOwnerID,
			perm:          "student:read:any",
			expectedGrant: true,
		},
		{
			name:          "Staff membaca data mahasiswa milik user lain (student:read:any)",
			user:          staffUser,
			ownerID:       studentOwnerID,
			perm:          "student:read:any",
			expectedGrant: true,
		},
		{
			name:          "Staff mencoba mengubah data mahasiswa milik user lain (staff tidak punya student:update:any)",
			user:          staffUser,
			ownerID:       studentOwnerID,
			perm:          "student:update:any",
			expectedGrant: false,
		},
		{
			name:          "Admin membaca data mahasiswa milik user lain (student:read:any)",
			user:          adminUser,
			ownerID:       studentOwnerID,
			perm:          "student:read:any",
			expectedGrant: true,
		},
		{
			name:          "Admin mengubah data mahasiswa milik user lain (student:update:any)",
			user:          adminUser,
			ownerID:       studentOwnerID,
			perm:          "student:update:any",
			expectedGrant: true,
		},
		{
			name:          "User lain mengakses data mahasiswa tanpa hak (Role user bukan pemilik)",
			user:          userOther,
			ownerID:       studentOwnerID,
			perm:          "student:read:any",
			expectedGrant: false,
		},
		{
			name:          "User lain mengubah data mahasiswa tanpa hak (Role user bukan pemilik)",
			user:          userOther,
			ownerID:       studentOwnerID,
			perm:          "student:update:any",
			expectedGrant: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.CanAccessStudent(tt.user, tt.ownerID, perms, tt.perm)
			if got != tt.expectedGrant {
				t.Errorf("CanAccessStudent() = %v, want %v", got, tt.expectedGrant)
			}
		})
	}

	t.Run("Fail Closed saat PermissionSet bernilai nil", func(t *testing.T) {
		got := service.CanAccessStudent(userOther, studentOwnerID, nil, "student:read:any")
		if got != false {
			t.Errorf("CanAccessStudent(nil) = %v, want false (Fail Closed)", got)
		}
	})
}
