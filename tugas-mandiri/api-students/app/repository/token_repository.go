package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/model"
)

// TokenRepository mendefinisikan operasi pada tabel refresh_tokens.
type TokenRepository interface {
	Save(ctx context.Context, rt model.RefreshToken) error
	FindActive(ctx context.Context, tokenHash string) (model.RefreshToken, error)
	Revoke(ctx context.Context, id int64) error
	RevokeAllForUser(ctx context.Context, userID int) error
}

type tokenPostgresRepository struct {
	pool *pgxpool.Pool
}

// NewTokenRepository mengembalikan interface, bukan struct konkret.
func NewTokenRepository(pool *pgxpool.Pool) TokenRepository {
	return &tokenPostgresRepository{pool: pool}
}

func (r *tokenPostgresRepository) Save(
	ctx context.Context, rt model.RefreshToken,
) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		rt.UserID, rt.TokenHash, rt.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("menyimpan refresh token: %w", err)
	}
	return nil
}

// FindActive mencari refresh token yang belum expired DAN belum di-revoke.
func (r *tokenPostgresRepository) FindActive(
	ctx context.Context, tokenHash string,
) (model.RefreshToken, error) {
	var rt model.RefreshToken
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		 FROM refresh_tokens
		 WHERE token_hash = $1
		   AND revoked_at IS NULL
		   AND expires_at > NOW()`, tokenHash,
	).Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.RevokedAt, &rt.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.RefreshToken{}, ErrNotFound
		}
		return model.RefreshToken{}, fmt.Errorf("mencari refresh token aktif: %w", err)
	}
	return rt, nil
}

// Revoke menandai satu refresh token sebagai dicabut.
func (r *tokenPostgresRepository) Revoke(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at = NOW() WHERE id = $1 AND revoked_at IS NULL`, id,
	)
	if err != nil {
		return fmt.Errorf("mencabut refresh token: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// RevokeAllForUser mencabut semua refresh token milik user tertentu.
// Dipakai saat logout.
func (r *tokenPostgresRepository) RevokeAllForUser(ctx context.Context, userID int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at = NOW()
		 WHERE user_id = $1 AND revoked_at IS NULL`, userID,
	)
	if err != nil {
		return fmt.Errorf("mencabut semua refresh token user: %w", err)
	}
	return nil
}
