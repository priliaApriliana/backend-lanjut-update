package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/model"
)

// Sentinel error milik lapisan repository, bukan error milik pgx.
// Handler cukup kenal dua ini, tidak perlu tahu basis datanya apa.
var (
	ErrNotFound  = errors.New("data tidak ditemukan")
	ErrDuplicate = errors.New("data sudah ada")
)

// interface
type StudentRepository interface {
	FindAll(ctx context.Context, q model.ListQuery) ([]model.Student, int, error)
	FindByID(ctx context.Context, id int) (model.Student, error)
	Create(ctx context.Context, s model.Student) (model.Student, error)
	Update(ctx context.Context, s model.Student) (model.Student, error)
	Delete(ctx context.Context, id int) error
	FindAfterCursor(ctx context.Context, q model.CursorQuery) ([]model.Student, error)
}

// kolomUrut adalah daftar putih: pemetaan dari nilai yang boleh dikirim
// klien ke nama kolom asli. ORDER BY tidak bisa diparameterkan, jadi nama
// kolom terpaksa disisipkan sebagai teks - daftar putih inilah satu-satunya
// yang mencegah SQL injection di titik ini.
var kolomUrut = map[string]string{
	"id":         "id",
	"nim":        "nim",
	"name":       "name",
	"grade":      "grade",
	"created_at": "created_at",
}

type studentPostgresRepository struct {
	pool *pgxpool.Pool
}

// NewStudentRepository mengembalikan interface, bukan struct konkret.
func NewStudentRepository(pool *pgxpool.Pool) StudentRepository {
	return &studentPostgresRepository{pool: pool}
}

// buildFilter menyusun bagian WHERE beserta argumennya.
// Nilai dari klien SELALU jadi argumen ($1, $2, ...), tidak pernah
// disambung langsung ke teks SQL.
func buildFilter(q model.ListQuery) (string, []any) {
	where := " WHERE 1 = 1"
	args := []any{}

	if q.Search != "" {
		where += fmt.Sprintf(" AND name ILIKE $%d", len(args)+1)
		args = append(args, "%"+q.Search+"%")
	}

	if q.IsActive != nil {
		where += fmt.Sprintf(" AND is_active = $%d", len(args)+1)
		args = append(args, *q.IsActive)
	}

	switch q.Grade {
	case "A":
		where += fmt.Sprintf(" AND grade >= $%d", len(args)+1)
		args = append(args, 3.50)
	case "B":
		where += fmt.Sprintf(" AND grade >= $%d AND grade < $%d", len(args)+1, len(args)+2)
		args = append(args, 3.00, 3.50)
	case "C":
		where += fmt.Sprintf(" AND grade >= $%d AND grade < $%d", len(args)+1, len(args)+2)
		args = append(args, 2.50, 3.00)
	case "D":
		where += fmt.Sprintf(" AND grade < $%d", len(args)+1)
		args = append(args, 2.50)
	}

	return where, args
}

func (r *studentPostgresRepository) FindAll(
	ctx context.Context, q model.ListQuery,
) ([]model.Student, int, error) {
	where, args := buildFilter(q)

	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM students"+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("menghitung mahasiswa: %w", err)
	}

	arah := "ASC"
	if q.Order == "desc" {
		arah = "DESC"
	}

	kolom, ok := kolomUrut[q.Sort]
	if !ok {
		kolom = "id"
	}

	sqlText := fmt.Sprintf(
		`SELECT id, nim, name, grade, is_active, owner_id, created_at
         FROM students%s
         ORDER BY %s %s
         LIMIT $%d OFFSET $%d`,
		where, kolom, arah, len(args)+1, len(args)+2,
	)
	args = append(args, q.Limit, q.Offset())

	rows, err := r.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("mengambil daftar mahasiswa: %w", err)
	}
	defer rows.Close()

	hasil := []model.Student{}
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(&s.ID, &s.NIM, &s.Name, &s.Grade, &s.IsActive,
			&s.OwnerID, &s.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("membaca baris mahasiswa: %w", err)
		}
		hasil = append(hasil, s)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("membaca hasil query: %w", err)
	}

	return hasil, total, nil
}

func (r *studentPostgresRepository) FindByID(
	ctx context.Context, id int,
) (model.Student, error) {
	var s model.Student
	err := r.pool.QueryRow(ctx,
		`SELECT id, nim, name, grade, is_active, owner_id, created_at
         FROM students WHERE id = $1`, id,
	).Scan(&s.ID, &s.NIM, &s.Name, &s.Grade, &s.IsActive,
		&s.OwnerID, &s.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("mengambil mahasiswa: %w", err)
	}
	return s, nil
}

// FindAfterCursor mengambil satu halaman mahasiswa memakai keyset
// pagination. Pasangan (created_at, id) dipakai sebagai penanda posisi:
// created_at saja tidak dijamin unik (dua mahasiswa bisa dibuat pada
// detik yang sama persis), sedangkan id (primary key) selalu unik,
// sehingga kombinasi keduanya menjamin urutan total yang konsisten.
func (r *studentPostgresRepository) FindAfterCursor(
	ctx context.Context, q model.CursorQuery,
) ([]model.Student, error) {
	args := []any{}
	where := ""

	if q.After != nil {
		args = append(args, q.After.CreatedAt, q.After.ID)
		where = fmt.Sprintf(" WHERE (created_at, id) < ($%d, $%d)",
			len(args)-1, len(args))
	}

	args = append(args, q.Limit+1)
	query := fmt.Sprintf(
		`SELECT id, nim, name, grade, is_active, owner_id, created_at
		 FROM students%s ORDER BY created_at DESC, id DESC LIMIT $%d`,
		where, len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar mahasiswa: %w", err)
	}
	defer rows.Close()

	hasil := []model.Student{}
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(&s.ID, &s.NIM, &s.Name, &s.Grade, &s.IsActive,
			&s.OwnerID, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("membaca baris mahasiswa: %w", err)
		}
		hasil = append(hasil, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca hasil query: %w", err)
	}

	return hasil, nil
}

func (r *studentPostgresRepository) Create(
	ctx context.Context, s model.Student,
) (model.Student, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO students (nim, name, grade, is_active, owner_id)
         VALUES ($1, $2, $3, $4, $5)
         RETURNING id, created_at`,
		s.NIM, s.Name, s.Grade, s.IsActive, s.OwnerID,
	).Scan(&s.ID, &s.CreatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, fmt.Errorf("menyimpan mahasiswa: %w", err)
	}
	return s, nil
}

func (r *studentPostgresRepository) Update(
	ctx context.Context, s model.Student,
) (model.Student, error) {
	err := r.pool.QueryRow(ctx,
		`UPDATE students SET nim = $1, name = $2, grade = $3, is_active = $4
         WHERE id = $5
         RETURNING id, nim, name, grade, is_active, owner_id, created_at`,
		s.NIM, s.Name, s.Grade, s.IsActive, s.ID,
	).Scan(&s.ID, &s.NIM, &s.Name, &s.Grade, &s.IsActive,
		&s.OwnerID, &s.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		if isUniqueViolation(err) {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, fmt.Errorf("memperbarui mahasiswa: %w", err)
	}
	return s, nil
}

func (r *studentPostgresRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM students WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("menghapus mahasiswa: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// isUniqueViolation memeriksa apakah error berasal dari pelanggaran
// batasan UNIQUE. Kode 23505 adalah kode resmi PostgreSQL untuk itu.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
