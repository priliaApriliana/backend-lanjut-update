package service_test

import (
	"testing"
	"time"

	"api-students/app/model"
	"api-students/app/service"
	"api-students/helper"
)

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
		if got := service.CountTotalPages(tc.total, tc.limit); got != tc.want {
			t.Errorf("total=%d limit=%d: diharapkan %d, diperoleh %d",
				tc.total, tc.limit, tc.want, got)
		}
	}
}

func TestStudentDeclarativeValidation_Create(t *testing.T) {
	// Kasus 1: Input valid
	validReq := model.CreateStudentRequest{
		NIM:      "434241022",
		Name:     "Putri Apriliana",
		Grade:    95.0,
		IsActive: true,
	}
	if errs := helper.ValidateStruct(validReq); errs != nil {
		t.Fatalf("seharusnya tidak ada error validasi untuk input valid, diperoleh: %v", errs)
	}

	// Kasus 2: NIM format salah (bukan digit), name kosong, grade di luar rentang
	invalidReq := model.CreateStudentRequest{
		NIM:      "bukan-digit",
		Name:     "",
		Grade:    120.0,
		IsActive: true,
	}
	errs := helper.ValidateStruct(invalidReq)
	if errs == nil {
		t.Fatalf("diharapkan error validasi, diperoleh nil")
	}
	if _, ok := errs["nim"]; !ok {
		t.Errorf("diharapkan error pada field nim")
	}
	if _, ok := errs["name"]; !ok {
		t.Errorf("diharapkan error pada field name")
	}
	if _, ok := errs["grade"]; !ok {
		t.Errorf("diharapkan error pada field grade")
	}
}

func TestStudentDeclarativeValidation_Patch(t *testing.T) {
	// Kasus 1: PATCH tanpa name (name nil) -> valid (dilewati oleh omitnil)
	grade := 90.0
	patchWithoutName := model.PatchStudentRequest{
		Grade: &grade,
	}
	if errs := helper.ValidateStruct(patchWithoutName); errs != nil {
		t.Fatalf("PATCH tanpa name seharusnya valid (dilewati omitnil), diperoleh: %v", errs)
	}

	// Kasus 2: PATCH dengan name bernilai kosong ("") -> DITOLAK (min=3)
	emptyName := ""
	patchWithEmptyName := model.PatchStudentRequest{
		Name: &emptyName,
	}
	errs := helper.ValidateStruct(patchWithEmptyName)
	if errs == nil {
		t.Fatalf("PATCH dengan name kosong seharusnya ditolak min=3, tetapi lolos")
	}
	if _, ok := errs["name"]; !ok {
		t.Errorf("diharapkan error pada field name saat dikirim string kosong")
	}
}

func TestApplyStudentPatch(t *testing.T) {
	initial := model.Student{
		ID:        1,
		NIM:       "434241022",
		Name:      "Putri",
		Grade:     85.0,
		IsActive:  true,
		CreatedAt: time.Now(),
		OwnerID:   1,
	}

	newName := "Putri Apriliana"
	newGrade := 98.0
	isAct := false

	req := model.PatchStudentRequest{
		Name:     &newName,
		Grade:    &newGrade,
		IsActive: &isAct,
	}

	updated := service.ApplyStudentPatch(initial, req)
	if updated.Name != newName {
		t.Errorf("Name gagal dimutasi: %s != %s", updated.Name, newName)
	}
	if updated.Grade != newGrade {
		t.Errorf("Grade gagal dimutasi: %.2f != %.2f", updated.Grade, newGrade)
	}
	if updated.IsActive != false {
		t.Errorf("IsActive gagal dimutasi: %v != false", updated.IsActive)
	}
	// NIM tidak diubah, harus tetap sama
	if updated.NIM != "434241022" {
		t.Errorf("NIM seharusnya tetap: %s", updated.NIM)
	}
}

func TestIsEmptyStudentPatch(t *testing.T) {
	empty := model.PatchStudentRequest{}
	if !service.IsEmptyStudentPatch(empty) {
		t.Errorf("diharapkan true untuk empty patch")
	}

	nim := "434241022"
	notEmpty := model.PatchStudentRequest{NIM: &nim}
	if service.IsEmptyStudentPatch(notEmpty) {
		t.Errorf("diharapkan false saat ada field yang dikirim")
	}
}
