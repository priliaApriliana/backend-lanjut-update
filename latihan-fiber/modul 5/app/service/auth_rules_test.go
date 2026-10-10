package service_test

import (
	"testing"

	"api-students/app/model"
	"api-students/app/service"
)

func TestCheckPasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected string
	}{
		{
			name:     "Password terlalu pendek (< 8)",
			password: "pass1",
			expected: "minimal 8 karakter",
		},
		{
			name:     "Password hanya huruf tanpa angka",
			password: "passwordrahasia",
			expected: "harus memuat huruf dan angka",
		},
		{
			name:     "Password hanya angka tanpa huruf",
			password: "1234567890",
			expected: "harus memuat huruf dan angka",
		},
		{
			name:     "Password masuk daftar weak password (password1)",
			password: "password1",
			expected: "password terlalu umum",
		},
		{
			name:     "Password masuk daftar weak password (admin123)",
			password: "admin123",
			expected: "password terlalu umum",
		},
		{
			name:     "Password kuat dan valid",
			password: "rahasia123",
			expected: "",
		},
		{
			name:     "Password kuat dengan simbol",
			password: "SuperSecret_2026",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.CheckPasswordStrength(tt.password)
			if got != tt.expected {
				t.Errorf("CheckPasswordStrength(%q) = %q, want %q", tt.password, got, tt.expected)
			}
		})
	}
}

func TestValidateRegister(t *testing.T) {
	tests := []struct {
		name       string
		req        model.RegisterRequest
		expectErrs bool
		errKey     string
	}{
		{
			name: "Registrasi Valid",
			req: model.RegisterRequest{
				Username: "putri_apr",
				Email:    "putri@example.com",
				Password: "rahasia123",
			},
			expectErrs: false,
		},
		{
			name: "Username Kosong",
			req: model.RegisterRequest{
				Username: "",
				Email:    "putri@example.com",
				Password: "rahasia123",
			},
			expectErrs: true,
			errKey:     "username",
		},
		{
			name: "Username Terlalu Pendek",
			req: model.RegisterRequest{
				Username: "ab",
				Email:    "putri@example.com",
				Password: "rahasia123",
			},
			expectErrs: true,
			errKey:     "username",
		},
		{
			name: "Username Karakter Ilegal",
			req: model.RegisterRequest{
				Username: "user@name#",
				Email:    "putri@example.com",
				Password: "rahasia123",
			},
			expectErrs: true,
			errKey:     "username",
		},
		{
			name: "Email Tidak Valid",
			req: model.RegisterRequest{
				Username: "putri_apr",
				Email:    "bukan-email",
				Password: "rahasia123",
			},
			expectErrs: true,
			errKey:     "email",
		},
		{
			name: "Password Lemah",
			req: model.RegisterRequest{
				Username: "putri_apr",
				Email:    "putri@example.com",
				Password: "password1",
			},
			expectErrs: true,
			errKey:     "password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := service.ValidateRegister(tt.req)
			if tt.expectErrs {
				if len(errs) == 0 {
					t.Fatalf("ValidateRegister(%v) diharapkan menghasilkan error, tetapi kosong", tt.req)
				}
				if tt.errKey != "" && errs[tt.errKey] == "" {
					t.Errorf("ValidateRegister(%v) diharapkan memiliki error key %q", tt.req, tt.errKey)
				}
			} else {
				if len(errs) > 0 {
					t.Fatalf("ValidateRegister(%v) menghasilkan error tak terduga: %v", tt.req, errs)
				}
			}
		})
	}
}

func TestValidateLogin(t *testing.T) {
	tests := []struct {
		name       string
		req        model.LoginRequest
		expectErrs bool
	}{
		{
			name: "Login Valid",
			req: model.LoginRequest{
				Username: "sari",
				Password: "rahasia123",
			},
			expectErrs: false,
		},
		{
			name: "Login Username Kosong",
			req: model.LoginRequest{
				Username: "  ",
				Password: "rahasia123",
			},
			expectErrs: true,
		},
		{
			name: "Login Password Kosong",
			req: model.LoginRequest{
				Username: "sari",
				Password: "",
			},
			expectErrs: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := service.ValidateLogin(tt.req)
			if (len(errs) > 0) != tt.expectErrs {
				t.Errorf("ValidateLogin(%v) got errs=%v, want expectErrs=%v", tt.req, errs, tt.expectErrs)
			}
		})
	}
}

func TestIsValidUsername(t *testing.T) {
	valid := []string{"putri", "putri_123", "user.name", "User_99"}
	for _, u := range valid {
		if !service.IsValidUsername(u) {
			t.Errorf("IsValidUsername(%q) = false, want true", u)
		}
	}

	invalid := []string{"user name", "user@host", "user#12", "admin!"}
	for _, u := range invalid {
		if service.IsValidUsername(u) {
			t.Errorf("IsValidUsername(%q) = true, want false", u)
		}
	}
}
