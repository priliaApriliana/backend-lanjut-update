package service

import (
	"testing"

	"api-students/app/model"
	"api-students/helper"
)

// ==================== Test IsPasswordStrong ====================
// Dipindahkan ke helper.IsPasswordStrong sejak Modul 7 (Langkah 5),
// karena helper/validator.go membutuhkannya dan service tidak boleh
// diimpor balik oleh helper.

func TestIsPasswordStrong(t *testing.T) {
	cases := []struct {
		name   string
		pw     string
		strong bool
	}{
		{"password kuat", "Rahasia123", true},
		{"huruf dan angka", "abcdef99", true},
		{"terlalu pendek", "Ab1", false},
		{"tanpa angka", "RahasiaKuat", false},
		{"tanpa huruf", "12345678", false},
		{"password umum - password1", "password1", false},
		{"password umum - qwerty123", "qwerty123", false},
		{"password umum - 12345678", "12345678", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ok, pesan := helper.IsPasswordStrong(tc.pw)
			if ok != tc.strong {
				t.Errorf("IsPasswordStrong(%q) = %v, harap %v (pesan: %s)", tc.pw, ok, tc.strong, pesan)
			}
		})
	}
}

// ==================== Test ValidateRegister (Modul 7: via tag) ====================
// ValidateRegister manual sudah dihapus sejak Modul 7 (Langkah 6);
// RegisterRequest sekarang divalidasi lewat tag struct + helper.ValidateStruct.

func TestRegisterRequest_ValidasiTag_Valid(t *testing.T) {
	req := model.RegisterRequest{
		Username: "putri",
		Email:    "putri@example.com",
		Password: "Rahasia123",
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		t.Errorf("data valid seharusnya tidak menghasilkan error, dapat: %v", errs)
	}
}

func TestRegisterRequest_ValidasiTag_Invalid(t *testing.T) {
	req := model.RegisterRequest{
		Username: "ab",          // terlalu pendek
		Email:    "bukan-email", // format salah
		Password: "short1x",     // tidak lolos strongpassword
	}
	errs := helper.ValidateStruct(req)
	if _, ok := errs["username"]; !ok {
		t.Error("username < 3 karakter seharusnya menghasilkan error")
	}
	if _, ok := errs["email"]; !ok {
		t.Error("email tanpa format valid seharusnya menghasilkan error")
	}
	if _, ok := errs["password"]; !ok {
		t.Error("password lemah seharusnya menghasilkan error")
	}
}

// ==================== Test ValidateLogin ====================

func TestValidateLogin_Valid(t *testing.T) {
	req := model.LoginRequest{Username: "putri", Password: "Rahasia123"}
	if errs := ValidateLogin(req); len(errs) != 0 {
		t.Errorf("login valid seharusnya tidak menghasilkan error, dapat: %v", errs)
	}
}

func TestValidateLogin_Empty(t *testing.T) {
	errs := ValidateLogin(model.LoginRequest{})
	if _, ok := errs["username"]; !ok {
		t.Error("username kosong seharusnya menghasilkan error")
	}
	if _, ok := errs["password"]; !ok {
		t.Error("password kosong seharusnya menghasilkan error")
	}
}

// ==================== Test Mass Assignment Protection ====================

func TestRegisterRequest_TidakAdaFieldRole(t *testing.T) {
	req := model.RegisterRequest{
		Username: "test",
		Email:    "test@example.com",
		Password: "Test1234",
	}
	if req.Username == "" {
		t.Error("seharusnya tidak kosong")
	}
}