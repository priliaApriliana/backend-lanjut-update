package helper

import (
	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
)

func Ok(c *fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusOK).JSON(model.Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func OkWithMeta(c *fiber.Ctx, message string, data any, meta *model.Meta) error {
	return c.Status(fiber.StatusOK).JSON(model.Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

func Created(c *fiber.Ctx, message string, data any, location string) error {
	c.Set("Location", location)
	return c.Status(fiber.StatusCreated).JSON(model.Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func Fail(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(model.Response{
		Success: false,
		Message: message,
	})
}

// dignkn khusus buat 422 (error per field)
func FailValidation(c *fiber.Ctx, errs map[string]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(model.Response{
		Success: false,
		Message: "validasi gagal",
		Errors:  errs,
	})
}

// Success menggantikan Ok — sekarang menerima status secara eksplisit,
// karena tidak semua respons sukses memakai 200 (contoh: 201 Created).
func Success(c *fiber.Ctx, status int, message string, data any) error {
	return c.Status(status).JSON(model.Response{
		Success: true,
		Message: message,
		Data: data,
	})
}

// SuccessCursor membalas endpoint list berbasis cursor pagination.
// Tidak memakai model.Response karena field Meta di sana bertipe *Meta
// (khusus offset-pagination) — bentuknya beda dari CursorMeta.
func SuccessCursor(c *fiber.Ctx, message string, data any, meta *model.CursorMeta) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
		"meta":    meta,
	})
}