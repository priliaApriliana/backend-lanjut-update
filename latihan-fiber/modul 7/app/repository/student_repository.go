package repository

import (
	"context"
	"errors"
	"fmt"

	"api-students/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Sentinel error: error milik lapisan repository, bukan error milik driver basis data.
var (
	ErrNotFound  = errors.New("data tidak ditemukan")
	ErrDuplicate = errors.New("data sudah ada")
)

// StudentRepository adalah KONTRAK penyimpanan data mahasiswa (Layer Use Case boundary).
// Tidak mengimpor Fiber atau HTTP library apa pun.
type StudentRepository interface {
	FindAfterCursor(ctx context.Context, q model.CursorQuery, ownerID *int) ([]model.Student, error)
	FindAll(ctx context.Context, q model.ListStudentQuery) ([]model.Student, int, error)
	FindByID(ctx context.Context, id int) (model.Student, error)
	Create(ctx context.Context, s model.Student) (model.Student, error)
	Update(ctx context.Context, s model.Student) (model.Student, error)
	Delete(ctx context.Context, id int) error
}

// kolomUrut adalah daftar putih (whitelist) nama kolom yang diizinkan untuk klausa ORDER BY.
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

// NewStudentRepository mengembalikan interface StudentRepository dengan implementasi PostgreSQL.
func NewStudentRepository(pool *pgxpool.Pool) StudentRepository {
	return &studentPostgresRepository{pool: pool}
}

// buildFilter menyusun klausa WHERE dinamis menggunakan query parameter aman ($1, $2, ...).
func buildFilter(q model.ListStudentQuery) (string, []any) {
	where := " WHERE 1 = 1"
	args := []any{}

	if q.Search != "" {
		where += fmt.Sprintf(" AND (name ILIKE $%d OR nim ILIKE $%d)",
			len(args)+1, len(args)+1)
		args = append(args, "%"+q.Search+"%")
	}

	if q.IsActive != nil {
		where += fmt.Sprintf(" AND is_active = $%d", len(args)+1)
		args = append(args, *q.IsActive)
	}

	if q.MinGrade != nil {
		where += fmt.Sprintf(" AND grade >= $%d", len(args)+1)
		args = append(args, *q.MinGrade)
	}

	if q.MaxGrade != nil {
		where += fmt.Sprintf(" AND grade <= $%d", len(args)+1)
		args = append(args, *q.MaxGrade)
	}

	return where, args
}

func (r *studentPostgresRepository) FindAll(
	ctx context.Context, q model.ListStudentQuery,
) ([]model.Student, int, error) {
	where, args := buildFilter(q)

	// 1. Hitung total baris yang cocok sebelum dipaginasi
	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM students"+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("menghitung total mahasiswa: %w", err)
	}

	// 2. Tentukan arah pengurutan aman
	arah := "ASC"
	if q.Order == "desc" {
		arah = "DESC"
	}

	colName, ok := kolomUrut[q.Sort]
	if !ok {
		colName = "id"
	}

	// 3. Ambil data satu halaman menggunakan LIMIT dan OFFSET di sisi basis data
	sqlText := fmt.Sprintf(
		`SELECT id, nim, name, grade, is_active, created_at, COALESCE(owner_id, 0)
		FROM students%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d`,
		where, colName, arah, len(args)+1, len(args)+2,
	)
	args = append(args, q.Limit, q.Offset())

	rows, err := r.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("mengambil daftar mahasiswa: %w", err)
	}
	defer rows.Close() // Mencegah kebocoran koneksi (connection pool starvation)

	hasil := []model.Student{}
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(&s.ID, &s.NIM, &s.Name, &s.Grade, &s.IsActive, &s.CreatedAt, &s.OwnerID); err != nil {
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
		`SELECT id, nim, name, grade, is_active, created_at, COALESCE(owner_id, 0)
		FROM students WHERE id = $1`, id,
	).Scan(&s.ID, &s.NIM, &s.Name, &s.Grade, &s.IsActive, &s.CreatedAt, &s.OwnerID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("mengambil mahasiswa: %w", err)
	}

	return s, nil
}

func (r *studentPostgresRepository) Create(
	ctx context.Context, s model.Student,
) (model.Student, error) {
	var ownerArg any = nil
	if s.OwnerID > 0 {
		ownerArg = s.OwnerID
	}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO students (nim, name, grade, is_active, owner_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`,
		s.NIM, s.Name, s.Grade, s.IsActive, ownerArg,
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
		RETURNING id, nim, name, grade, is_active, created_at, COALESCE(owner_id, 0)`,
		s.NIM, s.Name, s.Grade, s.IsActive, s.ID,
	).Scan(&s.ID, &s.NIM, &s.Name, &s.Grade, &s.IsActive, &s.CreatedAt, &s.OwnerID)

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

// isUniqueViolation memeriksa apakah error berasal dari batasan UNIQUE PostgreSQL (23505).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

const studentColumns = "id, nim, name, grade, is_active, created_at, COALESCE(owner_id, 0)"

func (r *studentPostgresRepository) FindAfterCursor(
	ctx context.Context, q model.CursorQuery, ownerID *int,
) ([]model.Student, error) {
	args := []any{}
	where := " WHERE 1 = 1"

	if ownerID != nil {
		args = append(args, *ownerID)
		where += fmt.Sprintf(" AND owner_id = $%d", len(args))
	}

	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		where += fmt.Sprintf(" AND (name ILIKE $%d OR nim ILIKE $%d)", len(args), len(args))
	}

	if q.IsActive != nil {
		args = append(args, *q.IsActive)
		where += fmt.Sprintf(" AND is_active = $%d", len(args))
	}

	if q.After != nil {
		args = append(args, q.After.CreatedAt, q.After.ID)
		where += fmt.Sprintf(" AND (created_at, id) < ($%d, $%d)", len(args)-1, len(args))
	}

	args = append(args, q.Limit+1)
	query := fmt.Sprintf(
		"SELECT %s FROM students%s ORDER BY created_at DESC, id DESC LIMIT $%d",
		studentColumns, where, len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar mahasiswa cursor: %w", err)
	}
	defer rows.Close()

	result := []model.Student{}
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(&s.ID, &s.NIM, &s.Name, &s.Grade, &s.IsActive, &s.CreatedAt, &s.OwnerID); err != nil {
			return nil, fmt.Errorf("membaca row mahasiswa: %w", err)
		}
		result = append(result, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca hasil query mahasiswa: %w", err)
	}

	return result, nil
}

