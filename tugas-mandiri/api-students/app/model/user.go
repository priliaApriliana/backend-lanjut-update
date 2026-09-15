package model

import "time"

// User memetakan tabel users di database.
// Field Password di-tag json:"-" agar tidak pernah muncul di response JSON.
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}
