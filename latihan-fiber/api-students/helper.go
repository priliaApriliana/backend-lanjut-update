package main

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// ok mengirim respons JSON status 200 OK dengan format WebResponse standar.
func ok(c *fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusOK).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// okList mengirim respons JSON status 200 OK lengkap dengan data koleksi dan metadata paginasi.
func okList(c *fiber.Ctx, message string, data any, meta *Meta) error {
	return c.Status(fiber.StatusOK).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// created mengirim respons JSON status 201 Created dengan header Location URL sumber daya baru.
func created(c *fiber.Ctx, message string, data any, location string) error {
	c.Set("Location", location)
	return c.Status(fiber.StatusCreated).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// noContent mengirim status 204 No Content tanpa body sama sekali.
func noContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// fail mengirim respons kegagalan dengan kode status HTTP tertentu.
func fail(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(WebResponse{
		Success: false,
		Message: message,
	})
}

// failValidation mengirim status 422 Unprocessable Entity beserta rincian error per field masukan.
func failValidation(c *fiber.Ctx, errs map[string]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(WebResponse{
		Success: false,
		Message: "validasi gagal",
		Errors:  errs,
	})
}

// allowedSort adalah daftar putih (whitelist) atribut yang diizinkan untuk pengurutan.
// Whitelisting sangat krusial untuk mencegah penyusupan identifier yang tidak valid atau SQL injection.
var allowedSort = map[string]bool{
	"id":         true,
	"nim":        true,
	"name":       true,
	"grade":      true,
	"created_at": true,
}

// parseListStudentQuery mengekstrak dan memvalidasi seluruh query parameter dengan nilai default yang aman.
func parseListStudentQuery(c *fiber.Ctx) ListStudentQuery {
	q := ListStudentQuery{
		Page:   c.QueryInt("page", 1),
		Limit:  c.QueryInt("limit", 10),
		Search: strings.TrimSpace(c.Query("search")),
		Sort:   strings.ToLower(c.Query("sort", "id")),
		Order:  strings.ToLower(c.Query("order", "asc")),
	}

	// 1. Validasi nilai page
	if q.Page < 1 {
		q.Page = 1
	}

	// 2. Validasi nilai limit dengan batas bawah (1) dan batas atas aman (50).
	// Alasan: Membatasi limit mencegah kehabisan memori server (OOM) dan serangan DoS jika klien
	// meminta data dalam jumlah masif (?limit=999999).
	if q.Limit < 1 {
		q.Limit = 10
	}
	if q.Limit > 50 {
		q.Limit = 50
	}

	// 3. Validasi daftar putih field pengurutan
	if !allowedSort[q.Sort] {
		q.Sort = "id"
	}

	// 4. Validasi arah pengurutan (hanya boleh "asc" atau "desc")
	if q.Order != "desc" {
		q.Order = "asc"
	}

	// 5. Filter status is_active jika disediakan klien
	if raw := c.Query("is_active"); raw != "" {
		if v, err := strconv.ParseBool(raw); err == nil {
			q.IsActive = &v
		}
	}

	// 6. Filter rentang nilai (min_grade & max_grade) jika disediakan klien
	if rawMin := c.Query("min_grade"); rawMin != "" {
		if v, err := strconv.ParseFloat(rawMin, 64); err == nil {
			q.MinGrade = &v
		}
	}
	if rawMax := c.Query("max_grade"); rawMax != "" {
		if v, err := strconv.ParseFloat(rawMax, 64); err == nil {
			q.MaxGrade = &v
		}
	}

	return q
}
