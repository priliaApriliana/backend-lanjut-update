package main

import (
    "context"
    "strconv"
    "strings"
    "time"

    "github.com/gofiber/fiber/v2"

    "api-students/app/model"
)

type Response struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
    Data    any    `json:"data,omitempty"`
    Meta    *Meta  `json:"meta,omitempty"`
    Errors  any    `json:"errors,omitempty"`
}

type Meta struct {
    Page       int `json:"page"`
    Limit      int `json:"limit"`
    Total      int `json:"total"`
    TotalPages int `json:"total_pages"`
}

func ok(c *fiber.Ctx, message string, data any) error {
    return c.Status(fiber.StatusOK).JSON(Response{
        Success: true,
        Message: message,
        Data:    data,
    })
}

func okWithMeta(c *fiber.Ctx, message string, data any, meta *Meta) error {
    return c.Status(fiber.StatusOK).JSON(Response{
        Success: true,
        Message: message,
        Data:    data,
        Meta:    meta,
    })
}

func created(c *fiber.Ctx, message string, data any, location string) error {
    c.Set("Location", location)
    return c.Status(fiber.StatusCreated).JSON(Response{
        Success: true,
        Message: message,
        Data:    data,
    })
}

func fail(c *fiber.Ctx, status int, message string) error {
    return c.Status(status).JSON(Response{
        Success: false,
        Message: message,
    })
}

// dignkn khusus buat 422 (error per field)
func failValidation(c *fiber.Ctx, errs map[string]string) error {
    return c.Status(fiber.StatusUnprocessableEntity).JSON(Response{
        Success: false,
        Message: "validasi gagal",
        Errors:  errs,
    })
}

// reqCtx memberi batas waktu untuk setiap operasi basis data.
// Tanpa batas waktu, satu query yang menggantung bisa menahan koneksi
// selamanya dan lama-lama menghabiskan seluruh isi pool.
func reqCtx(c *fiber.Ctx) (context.Context, context.CancelFunc) {
    return context.WithTimeout(c.UserContext(), 5*time.Second)
}

// parseListQuery mengambil dan menormalkan seluruh query parameter untuk
// endpoint List. Logikanya sama seperti pertemuan 2, cuma sekarang
// hasilnya dibungkus jadi model.ListQuery untuk diteruskan ke repository.
func parseListQuery(c *fiber.Ctx) model.ListQuery {
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
