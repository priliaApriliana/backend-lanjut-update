package model

import "time"

// RegisterRequest adalah body untuk POST /auth/register.
// TIDAK ADA field Role di sini — role selalu di-set "user" oleh server.
// Ini mencegah mass assignment: klien tidak bisa mengirim "role":"admin".
type RegisterRequest struct { 
    Username string `json:"username" validate:"required,min=3,max=30,username"` 
    Email    string `json:"email"    validate:"required,email,max=120"` 
    Password string `json:"password" validate:"required,max=72,strongpassword"` 
} 

// LoginRequest adalah body untuk POST /auth/login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RefreshRequest adalah body untuk POST /auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// TokenPair dikembalikan saat login dan refresh berhasil.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// RefreshToken memetakan tabel refresh_tokens di database.
type RefreshToken struct {
	ID        int64      `json:"id"`
	UserID    int        `json:"user_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// AuthUser adalah identitas ringkas yang disimpan di c.Locals
// setelah middleware RequireAuth berhasil memverifikasi token.
type AuthUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}
