package service_test

import (
	"testing"

	"api-students/app/model"
	"api-students/helper"
)

func TestDeclarativeValidation_Register(t *testing.T) {
	tests := []struct {
		name      string
		req       model.RegisterRequest
		wantField string
	}{
		{
			name: "Registrasi Valid",
			req: model.RegisterRequest{
				Username: "putri123",
				Email:    "putri@unair.ac.id",
				Password: "UnairHebat2026",
			},
			wantField: "",
		},
		{
			name: "Username Kurang dari 3 karakter",
			req: model.RegisterRequest{
				Username: "ab",
				Email:    "putri@unair.ac.id",
				Password: "Password123",
			},
			wantField: "username",
		},
		{
			name: "Username Karakter Ilegal",
			req: model.RegisterRequest{
				Username: "a b!",
				Email:    "putri@unair.ac.id",
				Password: "Password123",
			},
			wantField: "username",
		},
		{
			name: "Email Tidak Valid",
			req: model.RegisterRequest{
				Username: "putri123",
				Email:    "bukan-email",
				Password: "Password123",
			},
			wantField: "email",
		},
		{
			name: "Password Terlalu Pendek",
			req: model.RegisterRequest{
				Username: "putri123",
				Email:    "putri@unair.ac.id",
				Password: "abc",
			},
			wantField: "password",
		},
		{
			name: "Password Tanpa Angka",
			req: model.RegisterRequest{
				Username: "putri123",
				Email:    "putri@unair.ac.id",
				Password: "tanpaangka",
			},
			wantField: "password",
		},
		{
			name: "Password Terlalu Umum",
			req: model.RegisterRequest{
				Username: "putri123",
				Email:    "putri@unair.ac.id",
				Password: "password123",
			},
			wantField: "password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := helper.ValidateStruct(tt.req)
			if tt.wantField == "" {
				if len(errs) != 0 {
					t.Errorf("diharapkan valid, diperoleh errors: %v", errs)
				}
			} else {
				if _, ok := errs[tt.wantField]; !ok {
					t.Errorf("diharapkan field %q mengalami error validasi, diperoleh: %v", tt.wantField, errs)
				}
			}
		})
	}
}

func TestDeclarativeValidation_Login(t *testing.T) {
	valid := model.LoginRequest{
		Username: "admin",
		Password: "Password123",
	}
	if errs := helper.ValidateStruct(valid); len(errs) != 0 {
		t.Errorf("diharapkan login valid tidak error: %v", errs)
	}

	empty := model.LoginRequest{}
	errs := helper.ValidateStruct(empty)
	if _, ok := errs["username"]; !ok {
		t.Errorf("diharapkan username wajib diisi")
	}
	if _, ok := errs["password"]; !ok {
		t.Errorf("diharapkan password wajib diisi")
	}
}
