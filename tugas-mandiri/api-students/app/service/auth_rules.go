package service

import (
	"strings"
	// "unicode"

	"api-students/app/model"
	"api-students/helper"
)

// File ini berisi business rules MURNI untuk autentikasi.
// Tidak menyentuh fiber.Ctx, tidak menyentuh database,
// dan tidak tahu apa pun tentang HTTP.
// Pola ini sama persis dengan student_rules.go.


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
	} else if ok, pesan := helper.IsPasswordStrong(req.Password); !ok {
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
