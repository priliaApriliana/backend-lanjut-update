package service

import (
	"testing"

	"api-students/app/model"
	"api-students/helper"
)

// ==================== Test ValidateCreate (Modul 7: via struct tag) ====================
func TestValidateCreate(t *testing.T) {
	valid := model.CreateStudentRequest{NIM: "434241022", Name: "Putri", Grade: 3.5, IsActive: true}
	if errs := helper.ValidateStruct(valid); errs != nil {
		t.Errorf("data valid seharusnya tidak menghasilkan error, dapat: %v", errs)
	}

	invalid := model.CreateStudentRequest{NIM: "", Name: "  ", Grade: 4.5}
	errs := helper.ValidateStruct(invalid)
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

// ==================== Test ValidateReplace (Modul 7: via struct tag) ====================
func TestValidateReplace(t *testing.T) {
	invalid := model.UpdateStudentRequest{NIM: "434241022", Name: "Putri", Grade: -1}
	errs := helper.ValidateStruct(invalid)
	if _, ok := errs["grade"]; !ok {
		t.Error("grade negatif seharusnya menghasilkan error pada PUT")
	}
}

// ==================== Test ApplyPatch ====================
func TestApplyPatch(t *testing.T) {
	initial := model.Student{ID: 1, NIM: "434241022", Name: "Putri", Grade: 3.0, IsActive: true}
	inactive := false

	result := ApplyPatch(initial, model.PatchStudentRequest{IsActive: &inactive})
	if result.IsActive {
		t.Error("is_active seharusnya berubah menjadi false")
	}
	if result.Name != "Putri" {
		t.Error("field yang tidak dikirim seharusnya tidak berubah")
	}
}

// ==================== Test IsEmptyPatch ====================
func TestIsEmptyPatch(t *testing.T) {
	if !IsEmptyPatch(model.PatchStudentRequest{}) {
		t.Error("PatchStudentRequest kosong seharusnya dianggap empty patch")
	}
	name := "Putri"
	if IsEmptyPatch(model.PatchStudentRequest{Name: &name}) {
		t.Error("PatchStudentRequest dengan field terisi tidak boleh dianggap empty patch")
	}
}

// ==================== Test CountTotalPages ====================
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
