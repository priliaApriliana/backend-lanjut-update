package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/model"
)

// UserRepository mendefinisikan operasi pada tabel users.
// Modul 6: ditambahkan FindAll, Update, Delete, UpdateRole.
type UserRepository interface {
	Create(ctx context.Context, u model.User) (model.User, error)
	FindAll(ctx context.Context) ([]model.User, error)
	FindByID(ctx context.Context, id int) (model.User, error)
	FindByUsername(ctx context.Context, username string) (model.User, error)
	Update(ctx context.Context, u model.User) (model.User, error)
	Delete(ctx context.Context, id int) error
	UpdateRole(ctx context.Context, id int, role string) (model.User, error)
	FindAfterCursor(ctx context.Context, q model.CursorQuery) ([]model.User, error)
}

type userPostgresRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository mengembalikan interface, bukan struct konkret.
func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userPostgresRepository{pool: pool}
}

func (r *userPostgresRepository) Create(
	ctx context.Context, u model.User,
) (model.User, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (username, email, password, role)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`,
		u.Username, u.Email, u.Password, u.Role,
	).Scan(&u.ID, &u.CreatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return model.User{}, ErrDuplicate
		}
		return model.User{}, fmt.Errorf("menyimpan user: %w", err)
	}

	// Kosongkan password dari struct yang dikembalikan.
	u.Password = ""
	return u, nil
}

// FindAll mengembalikan seluruh user tanpa password.
func (r *userPostgresRepository) FindAll(
	ctx context.Context,
) ([]model.User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, username, email, role, is_active, created_at
		 FROM users ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar user: %w", err)
	}
	defer rows.Close()

	hasil := []model.User{}
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.IsActive, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("membaca baris user: %w", err)
		}
		hasil = append(hasil, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca hasil query user: %w", err)
	}

	return hasil, nil
}

func (r *userPostgresRepository) FindByID(
	ctx context.Context, id int,
) (model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, email, role, is_active, created_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.IsActive, &u.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mengambil user: %w", err)
	}
	return u, nil
}

// FindAfterCursor mengambil satu halaman memakai keyset pagination.
//
// id ikut dibandingkan karena created_at TIDAK dijamin unik. Bila dua
// baris dibuat pada mikrodetik yang sama dan hanya created_at yang
// dibandingkan, salah satu baris akan terlewat atau terkirim dua kali.
//
// Jumlah yang diminta sengaja limit+1. Baris tambahan itu tidak dikirim
// ke client; keberadaannya hanya dipakai untuk menjawab "masih ada
// halaman berikutnya?" tanpa perlu COUNT(*) atas seluruh tabel.
func (r *userPostgresRepository) FindAfterCursor(
	ctx context.Context, q model.CursorQuery,
) ([]model.User, error) {
	args := []any{}
	where := ""

	if q.After != nil {
		args = append(args, q.After.CreatedAt, q.After.ID)
		where = fmt.Sprintf(" WHERE (created_at, id) < ($%d, $%d)",
			len(args)-1, len(args))
	}

	args = append(args, q.Limit+1)
	query := fmt.Sprintf(
		`SELECT id, username, email, role, is_active, created_at
		 FROM users%s ORDER BY created_at DESC, id DESC LIMIT $%d`,
		where, len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar user: %w", err)
	}
	defer rows.Close()

	hasil := []model.User{}
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.IsActive, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("membaca baris user: %w", err)
		}
		hasil = append(hasil, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca hasil query user: %w", err)
	}

	return hasil, nil
}

// FindByUsername mengembalikan User BESERTA password hash.
// Dipakai khusus untuk proses login; password hash diperlukan
// agar bisa dibandingkan dengan bcrypt.
func (r *userPostgresRepository) FindByUsername(
	ctx context.Context, username string,
) (model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, email, password, role, is_active, created_at
		 FROM users WHERE LOWER(username) = LOWER($1)`, username,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mengambil user berdasarkan username: %w", err)
	}
	return u, nil
}

// Update memperbarui data profil user (username, email, is_active).
// TIDAK mengubah password atau role — keduanya punya jalur sendiri.
func (r *userPostgresRepository) Update(
	ctx context.Context, u model.User,
) (model.User, error) {
	err := r.pool.QueryRow(ctx,
		`UPDATE users SET username = $1, email = $2, is_active = $3
		 WHERE id = $4
		 RETURNING id, username, email, role, is_active, created_at`,
		u.Username, u.Email, u.IsActive, u.ID,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.IsActive, &u.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		if isUniqueViolation(err) {
			return model.User{}, ErrDuplicate
		}
		return model.User{}, fmt.Errorf("memperbarui user: %w", err)
	}
	return u, nil
}

// Delete menghapus user berdasarkan id.
func (r *userPostgresRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("menghapus user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateRole sengaja dipisahkan dari Update. Mengubah role adalah tindakan
// istimewa yang dijaga permission tersendiri, sehingga tidak boleh ikut
// terbawa oleh endpoint perubahan data biasa.
func (r *userPostgresRepository) UpdateRole(
	ctx context.Context, id int, role string,
) (model.User, error) {
	var updated model.User
	err := r.pool.QueryRow(ctx,
		`UPDATE users SET role = $1 WHERE id = $2
		 RETURNING id, username, email, role, is_active, created_at`,
		role, id,
	).Scan(&updated.ID, &updated.Username, &updated.Email,
		&updated.Role, &updated.IsActive, &updated.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mengubah role user: %w", err)
	}
	return updated, nil
}
