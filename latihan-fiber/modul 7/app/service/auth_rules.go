package service

import (
	"regexp"
	"strings"
	"unicode"
	"api-students/app/model"
)

const minPasswordLength = 8

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateRegister memeriksa keabsahan payload registrasi.
func ValidateRegister(req model.RegisterRequest) map[string]string {
	errs := map[string]string{}
	username := strings.TrimSpace(req.Username)
	switch {
	case username == "":
		errs["username"] = "wajib diisi"
	case len(username) < 3:
		errs["username"] = "minimal 3 karakter"
	case !IsValidUsername(username):
		errs["username"] = "hanya boleh huruf, angka, titik, dan garis bawah"
	}

	if !IsValidEmail(req.Email) {
		errs["email"] = "format email tidak valid"
	}

	if msg := CheckPasswordStrength(req.Password); msg != "" {
		errs["password"] = msg
	}

	return errs
}

// ValidateLogin hanya memeriksa kelengkapan field, BUKAN kekuatan password.
func ValidateLogin(req model.LoginRequest) map[string]string {
	errs := map[string]string{}
	if strings.TrimSpace(req.Username) == "" {
		errs["username"] = "wajib diisi"
	}
	if req.Password == "" {
		errs["password"] = "wajib diisi"
	}
	return errs
}

// CheckPasswordStrength memvalidasi panjang minimal 8 karakter, kombinasi huruf + angka,
// dan penolakan daftar password umum (weak passwords).
func CheckPasswordStrength(password string) string {
	if len(password) < minPasswordLength {
		return "minimal 8 karakter"
	}

	var hasLetter, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	if !hasLetter || !hasDigit {
		return "harus memuat huruf dan angka"
	}

	weak := map[string]bool{
		"password1":   true,
		"12345678":    true,
		"qwerty123":   true,
		"admin123":    true,
		"password123": true,
	}

	if weak[strings.ToLower(password)] {
		return "password terlalu umum"
	}

	return ""
}

// IsValidUsername memastikan username hanya terdiri dari karakter alfanumerik, titik, atau underscore.
func IsValidUsername(username string) bool {
	for _, r := range username {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '.' && r != '_' {
			return false
		}
	}
	return true
}

// IsValidEmail memvalidasi format email standar menggunakan regex RFC.
func IsValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if email == "" {
		return false
	}
	return emailRegex.MatchString(email)
}
