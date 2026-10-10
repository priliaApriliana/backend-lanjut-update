package model

import "time"

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,username"`
	Email    string `json:"email" validate:"required,email,max=120"`
	Password string `json:"password" validate:"required,max=72,strongpassword"`
	// Perhatikan: TIDAK ADA field Role di sini.
	// Bila ada, siapa pun dapat mendaftar sebagai admin (kerentanan mass assignment).
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"` // detik
}

// RefreshToken adalah row pada tabel refresh_tokens.
// Yang disimpan adalah TokenHash, bukan token aslinya.
type RefreshToken struct {
	ID         int64
	UserID     int
	TokenHash  string
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	CreatedAt  time.Time
}

// AuthUser adalah identitas yang dibawa access token.
// Isinya minimal: hanya yang benar-benar diperlukan middleware dan handler.
type AuthUser struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}
