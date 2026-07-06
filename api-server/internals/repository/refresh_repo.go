package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/uddinArsalan/devdeploy/internals/domain"
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
) (int64, error) {
	var refreshTokenID int64
	query := `
        INSERT INTO refresh_tokens (token_hash, user_id, expires_at)
        VALUES ($1, $2, $3)
		returning id
    `
	if err := rt.db.QueryRow(ctx, query, tokenHash, userID, expiresAt).Scan(&refreshTokenID); err != nil {
		return -1, err
	}
	return refreshTokenID, nil
}

func (rt *RefreshTokenRepo) FindRefreshToken(
	ctx context.Context,
	tokenHash string,
) (*domain.RefreshToken, error) {

	var refreshToken domain.RefreshToken
	query := `
        SELECT id, user_id, token_hash, revoked_at, replaced_by_token_id ,expires_at ,
				created_at, updated_at
				FROM refresh_tokens 
					WHERE token_hash = $1
    `
	err := rt.db.QueryRow(ctx, query, tokenHash).Scan(
		&refreshToken.ID,
		&refreshToken.UserID,
		&refreshToken.TokenHash,
		&refreshToken.RevokedAt,
		&refreshToken.ReplacedByTokenID,
		&refreshToken.ExpiresAt,
		&refreshToken.CreatedAt,
		&refreshToken.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &refreshToken, nil
}

func (rt *RefreshTokenRepo) UpdateRefreshToken(
	ctx context.Context,
	id int64,
	replacedByTokenID int64) error {

	query := `
        UPDATE refresh_tokens
			SET revoked_at = NOW(),
				replaced_by_token_id = $1,
				updated_at = NOW()
			WHERE id = $2
    `
	_, err := rt.db.Exec(ctx, query, replacedByTokenID, id)
	return err
}

func (rt *RefreshTokenRepo) CreateAndUpdateToken(ctx context.Context, oldTokenID int64, tokenHash string, userID int64, expiresAt time.Time) error {
	return pgx.BeginFunc(ctx, rt.db, func(tx pgx.Tx) error {
		var refreshTokenID int64
		query1 := `
        INSERT INTO refresh_tokens (token_hash, user_id, expires_at)
        VALUES ($1, $2, $3)
		returning id
    `
		if err := tx.QueryRow(ctx, query1, tokenHash, userID, expiresAt).Scan(&refreshTokenID); err != nil {
			return err
		}

		query2 := `
        UPDATE refresh_tokens
			SET revoked_at = NOW(),
				replaced_by_token_id = $1,
				updated_at = NOW()
			WHERE id = $2
    `
		_, err := tx.Exec(ctx, query2, refreshTokenID, oldTokenID)
		return err
	})
}
