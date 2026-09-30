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
	Username string `json:"username" validate:"required,min=3,max=30,alphanum"`
	Email    string `json:"email" validate:"required,email,max=120"`
	IsActive bool   `json:"is_active"`
}

// ErrorResponse adalah bentuk response kegagalan terpusat (Modul 7).
type ErrorResponse struct {
	Success   bool              `json:"success"`
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
}

// Mulai pertemuan 7, aturan validasi ditulis sebagai tag pada struct. 
// Aturan dan bentuk data berada pada baris yang sama, sehingga menambah 
type CreateUserRequest struct { 
    Username string `json:"username" validate:"required,min=3,max=30,alphanum"` 
    Email    string `json:"email"    validate:"required,email,max=120"` 
    Password string `json:"password" validate:"required,min=8,max=72,nospace"` 
}  
  
// Pada PATCH, pointer membedakan "tidak dikirim" (nil) dari "dikirim 
// bernilai kosong". omitnil dipilih karena ia menyatakan maksud yang 
// sebenarnya: lewati hanya bila nil. 
type PatchUserRequest struct { 
    Username *string  `json:"username,omitempty"  validate:"omitnil,min=3,max=30,alphanum"` 
    Email    *string `json:"email,omitempty"     validate:"omitnil,email,max=120"` 
    IsActive *bool   `json:"is_active,omitempty"` 
} 

// cursor menandai posisi terakhir yang sudah dibaca client.
// dua kolom (CreatedAt + ID) dipakai bersama karena CreatedAt saja
// tidak dijamin unik - dua baris bisa dibuat diwaktu yang sama persis
type Cursor struct {
	CreatedAt time.Time
	ID        int
}

// cursorQuery adalah parameter untuk endpoint list berbasis cursor.
type CursorQuery struct {
	Limit int
	After *Cursor
}

// CursorMeta menggantikan Meta pada endpoint yang memakai cursor. 
// 
// Perhatikan tidak adanya Total dan TotalPages. Keduanya tidak dapat 
// disediakan tanpa COUNT(*) atas seluruh tabel — persis biaya yang ingin 
// dihindari oleh pagination berbasis cursor. 
type CursorMeta struct { 
    Limit      int    `json:"limit"` 
    NextCursor string `json:"next_cursor,omitempty"` 
    HasMore    bool   `json:"has_more"` 
}