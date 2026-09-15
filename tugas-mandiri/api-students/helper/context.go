package helper

import (
	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
)

// ContextKeyAuth adalah kunci tunggal untuk menyimpan dan membaca
// identitas pengguna di c.Locals. Didefinisikan sebagai konstanta
// di satu tempat agar tidak ada hardcode string tersebar di banyak file.
const ContextKeyAuth = "auth_user"

// CurrentUser membaca identitas pengguna yang sudah diverifikasi
// dari c.Locals. Mengembalikan nil jika belum login (middleware
// belum menyimpan data).
func CurrentUser(c *fiber.Ctx) *model.AuthUser {
	user, ok := c.Locals(ContextKeyAuth).(*model.AuthUser)
	if !ok {
		return nil
	}
	return user
}
