package service

import (
	"testing"

	"api-students/app/model"
)

func TestValidateCreate(t *testing.T) {
	valid := model.CreateStudentRequest{NIM: "434241022", Name: "Putri", Grade: 3.5, IsActive: true}
	if errs := ValidateCreate(valid); len(errs) != 0 {
		t.Errorf("data valid seharusnya tidak menghasilkan error, dapat: %v", errs)
	}

	invalid := model.CreateStudentRequest{NIM: "", Name: "  ", Grade: 4.5}
	errs := ValidateCreate(invalid)
	if _, ok := errs["nim"]; !ok {
		t.Error("nim kosong seharusnya menghasilkan error")
	}
	if _, ok := errs["name"]; !ok {
		t.Error("name kosong seharusnya menghasilkan error")
	}
	if _, ok := errs["grade"]; !ok {
		t.Error("grade > 4.0 seharusnya menghasilkan error")
	}
}

func TestValidateReplace(t *testing.T) {
	invalid := model.UpdateStudentRequest{NIM: "434241022", Name: "Putri", Grade: -1}
	errs := ValidateReplace(invalid)
	if _, ok := errs["grade"]; !ok {
		t.Error("grade negatif seharusnya menghasilkan error pada PUT")
	}
}

func TestApplyPatch(t *testing.T) {
	initial := model.Student{ID: 1, NIM: "434241022", Name: "Putri", Grade: 3.0, IsActive: true}
	inactive := false

	result, errs := ApplyPatch(initial, model.PatchStudentRequest{IsActive: &inactive})
	if len(errs) != 0 {
		t.Fatalf("tidak seharusnya ada error: %v", errs)
	}
	if result.IsActive {
		t.Error("is_active seharusnya berubah menjadi false")
	}
	if result.Name != "Putri" {
		t.Error("field yang tidak dikirim seharusnya tidak berubah")
	}
}

func TestIsEmptyPatch(t *testing.T) {
	if !IsEmptyPatch(model.PatchStudentRequest{}) {
		t.Error("PatchStudentRequest kosong seharusnya dianggap empty patch")
	}
	name := "Putri"
	if IsEmptyPatch(model.PatchStudentRequest{Name: &name}) {
		t.Error("PatchStudentRequest dengan field terisi tidak boleh dianggap empty patch")
	}
}

func TestCountTotalPages(t *testing.T) {
	cases := []struct{ total, limit, want int }{
		{0, 10, 0},
		{1, 10, 1},
		{10, 10, 1},
		{11, 10, 2},
		{137, 20, 7},
	}
	for _, tc := range cases {
		if got := CountTotalPages(tc.total, tc.limit); got != tc.want {
			t.Errorf("total=%d limit=%d: harap %d, dapat %d", tc.total, tc.limit, tc.want, got)
		}
	}
}
