package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/uddinArsalan/devdeploy/internals/domain"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) UserRepo {
	return UserRepo{
		db: db,
	}
}

func (u *UserRepo) CreateUser(ctx context.Context, userDetails domain.CreateUser) error {
	query := `
			INSERT INTO users(name,email) VALUES ($1,$2)
			`
	_, err := u.db.Exec(ctx, query, userDetails.Name, userDetails.Email)
	return err
}
