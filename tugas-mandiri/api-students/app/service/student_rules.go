package service

import (
	"strings"

	"api-students/app/model"
)

// File ini berisi business rules MURNI: tidak menyentuh fiber.Ctx,
// tidak menyentuh database, dan tidak tahu apa pun tentang HTTP.

func ValidateCreate(req model.CreateStudentRequest) map[string]string {
	errs := map[string]string{}
	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "nim tidak boleh kosong"
	}
	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "name tidak boleh kosong"
	}
	if req.Grade < 0 || req.Grade > 4.0 {
		errs["grade"] = "grade harus antara 0 dan 4.0"
	}
	return errs
}

func ValidateReplace(req model.UpdateStudentRequest) map[string]string {
	errs := map[string]string{}
	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "nim tidak boleh kosong"
	}
	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "name tidak boleh kosong"
	}
	if req.Grade < 0 || req.Grade > 4.0 {
		errs["grade"] = "grade harus antara 0 dan 4.0"
	}
	return errs
}

func ApplyPatch(
	current model.Student, req model.PatchStudentRequest,
) (model.Student, map[string]string) {
	errs := map[string]string{}

	if req.NIM != nil {
		if strings.TrimSpace(*req.NIM) == "" {
			errs["nim"] = "nim tidak boleh kosong"
		} else {
			current.NIM = *req.NIM
		}
	}
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			errs["name"] = "name tidak boleh kosong"
		} else {
			current.Name = *req.Name
		}
	}
	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 4.0 {
			errs["grade"] = "grade harus antara 0 dan 4.0"
		} else {
			current.Grade = *req.Grade
		}
	}
	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}

	return current, errs
}

func IsEmptyPatch(req model.PatchStudentRequest) bool {
	return req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil
}

func CountTotalPages(total, limit int) int {
	if limit <= 0 {
		return 0
	}
	return (total + limit - 1) / limit
}