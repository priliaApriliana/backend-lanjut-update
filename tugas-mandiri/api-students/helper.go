package main

import "github.com/gofiber/fiber/v2"

type Response struct {
	Success bool		`json:"success"`
	Message string 		`json:"message"`
	Data 	any			`json:"data,omitempty"`
	Meta 	*Meta		`json:"meta,omitempty"`
	Errors	any			`json:"errors,omitempty"`
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
		Data: data,
	})
}

func okWithMeta(c *fiber.Ctx, message string, data any, meta *Meta) error {
	return c.Status(fiber.StatusOK).JSON(Response{
		Success: true,
		Message: message,
		Data: data,
		Meta: meta,
	})
}

func created(c *fiber.Ctx, message string, data any, location string) error {
	c.Set("Location", location)
	return c.Status(fiber.StatusCreated).JSON(Response{
		Success: true,
		Message: message,
		Data: data,
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
		Errors: errs,
	})
}