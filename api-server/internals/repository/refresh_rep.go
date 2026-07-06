package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshTokenRepo struct {
	db *pgxpool.Pool
}

func NewRefreshTokenRepo(db *pgxpool.Pool) *RefreshTokenRepo {
	return &RefreshTokenRepo{
		db,
	}
}

func (rt *RefreshTokenRepo) CreateRefreshToken(
	ctx context.Context,
	userID int64,
	tokenHash string,
	expiresAt time.Time,
) error {
	query := `
        INSERT INTO refresh_tokens (token_hash, user_id, expires_at)
        VALUES ($1, $2, $3)
    `
	_, err := rt.db.Exec(ctx, query, tokenHash, userID, expiresAt)
	return err
}
