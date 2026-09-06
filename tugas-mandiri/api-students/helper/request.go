package helper

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
)

// RequestContext memberi batas waktu untuk setiap operasi basis data.
func RequestContext(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}

// ParamID membaca parameter :id dari jalur.
func ParamID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return 0, false
	}
	return id, true
}

// ParseListQuery mengambil dan menormalkan seluruh query parameter untuk
// endpoint List.
func ParseListQuery(c *fiber.Ctx) model.ListQuery {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	if limit > 100 {
		limit = 100
	}
	if page < 1 {
		page = 1
	}

	q := model.ListQuery{
		Page:   page,
		Limit:  limit,
		Search: c.Query("search", ""),
		Sort:   c.Query("sort", "id"),
		Order:  c.Query("order", "asc"),
		Grade:  strings.ToUpper(c.Query("grade", "")),
	}

	if raw := c.Query("is_active", ""); raw != "" {
		if parsed, err := strconv.ParseBool(raw); err == nil {
			q.IsActive = &parsed
		}
	}

	return q
}