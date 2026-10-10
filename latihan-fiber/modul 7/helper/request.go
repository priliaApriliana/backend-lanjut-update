package helper

import (
	"context"
	"strconv"
	"strings"
	"time"

	"api-students/app/model"

	"github.com/gofiber/fiber/v2"
)

// RequestContext memberi batas waktu (timeout) 5 detik untuk setiap operasi basis data.
func RequestContext(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}

// ParamID membaca dan memvalidasi parameter :id harus berupa integer positif.
func ParamID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

// allowedSort adalah whitelist kolom pengurutan di sisi helper.
var allowedSort = map[string]bool{
	"id":         true,
	"nim":        true,
	"name":       true,
	"grade":      true,
	"created_at": true,
}

// ParseListQuery mengekstrak dan memvalidasi query parameter dari *fiber.Ctx dengan nilai default aman.
func ParseListQuery(c *fiber.Ctx) model.ListStudentQuery {
	q := model.ListStudentQuery{
		Page:   c.QueryInt("page", 1),
		Limit:  c.QueryInt("limit", 10),
		Search: strings.TrimSpace(c.Query("search")),
		Sort:   strings.ToLower(c.Query("sort", "id")),
		Order:  strings.ToLower(c.Query("order", "asc")),
	}

	if q.Page < 1 {
		q.Page = 1
	}

	// Batas aman paginasi: maksimum 50 untuk mencegah kehabisan memori (OOM)
	if q.Limit < 1 {
		q.Limit = 10
	}
	if q.Limit > 50 {
		q.Limit = 50
	}

	if !allowedSort[q.Sort] {
		q.Sort = "id"
	}

	if q.Order != "desc" {
		q.Order = "asc"
	}

	if raw := c.Query("is_active"); raw != "" {
		if v, err := strconv.ParseBool(raw); err == nil {
			q.IsActive = &v
		}
	}

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

// RequestID mengambil request ID dari locals atau header HTTP.
func RequestID(c *fiber.Ctx) string {
	if id, ok := c.Locals("requestid").(string); ok && id != "" {
		return id
	}
	return c.Get(fiber.HeaderXRequestID)
}

// ParseCursorQuery mengekstrak dan memvalidasi parameter query untuk keyset pagination.
func ParseCursorQuery(c *fiber.Ctx) (model.CursorQuery, error) {
	limit := c.QueryInt("limit", 10)
	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	q := model.CursorQuery{
		Limit:  limit,
		Search: strings.TrimSpace(c.Query("search")),
	}

	if rawCursor := strings.TrimSpace(c.Query("cursor")); rawCursor != "" {
		cur, err := DecodeCursor(rawCursor)
		if err != nil {
			return q, BadRequest("cursor tidak valid")
		}
		q.After = &cur
	}

	if raw := strings.TrimSpace(c.Query("is_active")); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return q, BadRequest("is_active harus berupa boolean (true/false)")
		}
		q.IsActive = &v
	}

	return q, nil
}

