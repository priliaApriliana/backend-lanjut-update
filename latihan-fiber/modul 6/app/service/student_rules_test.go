package service

import (
	"testing"
	"time"

	"api-students/app/model"
)

// Unit test ini menguji Business Rules Murni secara terisolasi.
// Tidak memerlukan koneksi database PostgreSQL, server HTTP, maupun objek fiber.Ctx.

func TestCountTotalPages(t *testing.T) {
	cases := []struct {
		total, limit, want int
	}{
		{0, 10, 0},
		{1, 10, 1},
		{10, 10, 1},
		{11, 10, 2},
		{35, 10, 4},
		{100, 20, 5},
		{101, 20, 6},
	}

	for _, tc := range cases {
		if got := CountTotalPages(tc.total, tc.limit); got != tc.want {
			t.Errorf("total=%d limit=%d: diharapkan %d, diperoleh %d",
				tc.total, tc.limit, tc.want, got)
		}
	}
}

func TestValidateCreate(t *testing.T) {
	// Kasus 1: Input valid
	validReq := model.CreateStudentRequest{
		NIM:      "434241022",
		Name:     "Putri Apriliana",
		Grade:    95.0,
		IsActive: true,
	}
	if errs := ValidateCreate(validReq); len(errs) != 0 {
		t.Fatalf("seharusnya tidak ada error validasi untuk input valid, diperoleh: %v", errs)
	}

	// Kasus 2: Input tidak valid (NIM kosong, nama kosong, grade di luar batas)
	invalidReq := model.CreateStudentRequest{
		NIM:      "",
		Name:     "   ",
		Grade:    120.0,
		IsActive: true,
	}
	errs := ValidateCreate(invalidReq)
	if len(errs) != 3 {
		t.Fatalf("diharapkan 3 field error, diperoleh %d: %v", len(errs), errs)
	}
	if errs["nim"] != "NIM wajib diisi" {
		t.Errorf("pesan error nim tidak cocok: %v", errs["nim"])
	}
	if errs["name"] != "nama mahasiswa wajib diisi" {
		t.Errorf("pesan error name tidak cocok: %v", errs["name"])
	}
	if errs["grade"] != "nilai grade harus berada pada rentang 0.0 sampai 100.0" {
		t.Errorf("pesan error grade tidak cocok: %v", errs["grade"])
	}
}

func TestValidateReplace(t *testing.T) {
	// Kasus 1: Input valid untuk PUT
	validReq := model.ReplaceStudentRequest{
		NIM:      "434241022",
		Name:     "Putri Apriliana S.Tr.Kom",
		Grade:    98.5,
		IsActive: false,
	}
	if errs := ValidateReplace(validReq); len(errs) != 0 {
		t.Fatalf("seharusnya tidak ada error validasi untuk PUT valid, diperoleh: %v", errs)
	}

	// Kasus 2: Input tidak valid (field kosong pada PUT)
	invalidReq := model.ReplaceStudentRequest{
		NIM:   "",
		Name:  "",
		Grade: -5.0,
	}
	errs := ValidateReplace(invalidReq)
	if errs["nim"] != "NIM wajib diisi pada PUT" {
		t.Errorf("pesan error nim PUT tidak cocok: %v", errs["nim"])
	}
	if errs["name"] != "nama mahasiswa wajib diisi pada PUT" {
		t.Errorf("pesan error name PUT tidak cocok: %v", errs["name"])
	}
	if errs["grade"] != "nilai grade harus berada pada rentang 0.0 sampai 100.0" {
		t.Errorf("pesan error grade PUT tidak cocok: %v", errs["grade"])
	}
}

func TestApplyPatch(t *testing.T) {
	initial := model.Student{
		ID:        1,
		NIM:       "434241022",
		Name:      "Putri Apriliana",
		Grade:     90.0,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	// Uji 1: Hanya update Grade dan IsActive
	newGrade := 99.5
	newActive := false
	patchReq := model.PatchStudentRequest{
		Grade:    &newGrade,
		IsActive: &newActive,
	}

	result, errs := ApplyPatch(initial, patchReq)
	if len(errs) != 0 {
		t.Fatalf("tidak boleh ada error validasi pada patch yang benar: %v", errs)
	}

	// Field yang dikirim harus berubah
	if result.Grade != 99.5 {
		t.Errorf("grade seharusnya 99.5, diperoleh %.2f", result.Grade)
	}
	if result.IsActive != false {
		t.Errorf("is_active seharusnya false, diperoleh %v", result.IsActive)
	}

	// Field yang TIDAK dikirim harus tetap utuh (tidak berubah)
	if result.NIM != "434241022" {
		t.Errorf("NIM tidak boleh berubah, diperoleh %v", result.NIM)
	}
	if result.Name != "Putri Apriliana" {
		t.Errorf("Name tidak boleh berubah, diperoleh %v", result.Name)
	}

	// Uji 2: Patch dengan grade tidak valid (< 0)
	invalidGrade := -10.0
	_, errsInvalid := ApplyPatch(initial, model.PatchStudentRequest{Grade: &invalidGrade})
	if errsInvalid["grade"] == "" {
		t.Error("seharusnya menghasilkan error validasi grade negatif pada PATCH")
	}
}

func TestIsEmptyPatch(t *testing.T) {
	empty := model.PatchStudentRequest{}
	if !IsEmptyPatch(empty) {
		t.Error("IsEmptyPatch seharusnya mengembalikan true untuk struct kosong")
	}

	val := "434241022"
	nonEmpty := model.PatchStudentRequest{NIM: &val}
	if IsEmptyPatch(nonEmpty) {
		t.Error("IsEmptyPatch seharusnya mengembalikan false ketika ada field yang dikirim")
	}
}
