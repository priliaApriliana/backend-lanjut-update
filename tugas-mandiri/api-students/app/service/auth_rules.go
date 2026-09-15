package service

import (
	"strings"
	"unicode"

	"api-students/app/model"
)

// File ini berisi business rules MURNI untuk autentikasi.
// Tidak menyentuh fiber.Ctx, tidak menyentuh database,
// dan tidak tahu apa pun tentang HTTP.
// Pola ini sama persis dengan student_rules.go.

// passwordUmum adalah daftar kecil password yang sering dipakai dan
// harus ditolak meskipun memenuhi syarat panjang dan karakter.
var passwordUmum = map[string]bool{
	"password1":  true,
	"password12": true,
	"12345678":   true,
	"123456789":  true,
	"1234567890": true,
	"qwerty123":  true,
	"abcdef123":  true,
	"abc12345":   true,
	"admin123":   true,
	"welcome1":   true,
}

// ValidateRegister memeriksa isi permintaan register.
func ValidateRegister(req model.RegisterRequest) map[string]string {
	errs := map[string]string{}

	username := strings.TrimSpace(req.Username)
	if username == "" {
		errs["username"] = "username tidak boleh kosong"
	} else if len(username) < 3 {
		errs["username"] = "username minimal 3 karakter"
	}

	email := strings.TrimSpace(req.Email)
	if email == "" {
		errs["email"] = "email tidak boleh kosong"
	} else if !isValidEmail(email) {
		errs["email"] = "format email tidak valid"
	}

	if req.Password == "" {
		errs["password"] = "password tidak boleh kosong"
	} else if ok, pesan := IsPasswordStrong(req.Password); !ok {
		errs["password"] = pesan
	}

	return errs
}

// ValidateLogin memeriksa isi permintaan login.
func ValidateLogin(req model.LoginRequest) map[string]string {
	errs := map[string]string{}
	if strings.TrimSpace(req.Username) == "" {
		errs["username"] = "username tidak boleh kosong"
	}
	if req.Password == "" {
		errs["password"] = "password tidak boleh kosong"
	}
	return errs
}

// IsPasswordStrong memeriksa kekuatan password.
// Syarat: minimal 8 karakter, mengandung huruf DAN angka,
// dan bukan termasuk daftar password umum.
func IsPasswordStrong(pw string) (bool, string) {
	if len(pw) < 8 {
		return false, "password minimal 8 karakter"
	}

	if passwordUmum[strings.ToLower(pw)] {
		return false, "password terlalu umum, gunakan yang lebih unik"
	}

	adaHuruf := false
	adaAngka := false
	for _, ch := range pw {
		if unicode.IsLetter(ch) {
			adaHuruf = true
		}
		if unicode.IsDigit(ch) {
			adaAngka = true
		}
	}

	if !adaHuruf {
		return false, "password harus mengandung huruf"
	}
	if !adaAngka {
		return false, "password harus mengandung angka"
	}

	return true, ""
}

// isValidEmail melakukan validasi sederhana: harus ada tepat satu '@'
// dan minimal satu '.' setelah '@'.
func isValidEmail(email string) bool {
	at := strings.Index(email, "@")
	if at < 1 {
		return false
	}
	domain := email[at+1:]
	if len(domain) < 3 {
		return false
	}
	return strings.Contains(domain, ".")
}
