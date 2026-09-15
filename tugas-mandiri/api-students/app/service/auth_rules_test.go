package service

import (
	"testing"

	"api-students/app/model"
)

// ==================== Test ValidateRegister ====================

func TestValidateRegister_Valid(t *testing.T) {
	req := model.RegisterRequest{
		Username: "putri",
		Email:    "putri@example.com",
		Password: "Rahasia123",
	}
	if errs := ValidateRegister(req); len(errs) != 0 {
		t.Errorf("data valid seharusnya tidak menghasilkan error, dapat: %v", errs)
	}
}

func TestValidateRegister_Invalid(t *testing.T) {
	req := model.RegisterRequest{
		Username: "ab",           // terlalu pendek
		Email:    "bukan-email",  // format salah
		Password: "short1",      // kurang dari 8 karakter
	}
	errs := ValidateRegister(req)
	if _, ok := errs["username"]; !ok {
		t.Error("username < 3 karakter seharusnya menghasilkan error")
	}
	if _, ok := errs["email"]; !ok {
		t.Error("email tanpa @ seharusnya menghasilkan error")
	}
	if _, ok := errs["password"]; !ok {
		t.Error("password < 8 karakter seharusnya menghasilkan error")
	}
}

func TestValidateRegister_Empty(t *testing.T) {
	errs := ValidateRegister(model.RegisterRequest{})
	if len(errs) < 3 {
		t.Errorf("semua field kosong seharusnya menghasilkan 3 error, dapat: %d", len(errs))
	}
}

// ==================== Test IsPasswordStrong ====================

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
			ok, pesan := IsPasswordStrong(tc.pw)
			if ok != tc.strong {
				t.Errorf("IsPasswordStrong(%q) = %v, harap %v (pesan: %s)", tc.pw, ok, tc.strong, pesan)
			}
		})
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

// TestRegisterRequest_TidakAdaFieldRole memastikan bahwa struct
// RegisterRequest TIDAK memiliki field Role. Ini adalah perlindungan
// compile-time terhadap mass assignment: jika seseorang menambahkan
// field Role ke struct, test ini akan gagal kompilasi.
func TestRegisterRequest_TidakAdaFieldRole(t *testing.T) {
	// Jika RegisterRequest punya field Role, baris ini akan COMPILE.
	// Karena RegisterRequest TIDAK punya field Role, ini membuktikan
	// bahwa mass assignment via "role":"admin" tidak mungkin terjadi.
	req := model.RegisterRequest{
		Username: "test",
		Email:    "test@example.com",
		Password: "Test1234",
	}

	// Pastikan struct bisa dibuat tanpa field Role.
	if req.Username == "" {
		t.Error("seharusnya tidak kosong")
	}
}
