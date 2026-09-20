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

// AssignRoleRequest dipakai endpoint PATCH /users/:id/role.
type AssignRoleRequest struct {
	Role string `json:"role"`
}

// UpdateUserRequest dipakai endpoint PUT /users/:id.
type UpdateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
}

// PatchUserRequest dipakai endpoint PATCH /users/:id.
// Field pointer: nil berarti "tidak diubah".
type PatchUserRequest struct {
	Username *string `json:"username,omitempty"`
	Email    *string `json:"email,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
}
