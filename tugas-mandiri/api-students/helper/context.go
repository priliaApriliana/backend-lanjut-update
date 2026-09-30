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
// dari c.Locals. ok bernilai false jika belum login.
func CurrentUser(c *fiber.Ctx) (model.AuthUser, bool) {
	user, ok := c.Locals(ContextKeyAuth).(*model.AuthUser)
	if !ok || user == nil{
		return model.AuthUser{}, false
	}
	return *user, true
}

// RequestID mengambil request ID dari Locals middleware requestid.
func RequestID(c *fiber.Ctx) string {
	id, _ := c.Locals("requestid").(string)
	return id
}