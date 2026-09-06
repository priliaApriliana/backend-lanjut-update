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