package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/uddinArsalan/devdeploy/internals/domain"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

func (u *UserRepo) CreateUser(ctx context.Context, name string, email string, passwordHash string) error {
	query := `
			INSERT INTO users(name,email,password_hash) VALUES ($1,$2,$3)
			`
	_, err := u.db.Exec(ctx, query, name, email, passwordHash)
	return err
}

func (u *UserRepo) FindUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	query := `
			SELECT id,name,email,role,password_hash,created_at FROM users
				WHERE email = $1
			`
	if err := u.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Role,
		&user.PasswordHash,
		&user.CreatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *UserRepo) FindUserByID(ctx context.Context, id int64) (*domain.User, error) {
	var user domain.User
	query := `
			SELECT id,name,email,role,password_hash,created_at
			 	FROM users
					WHERE id = $1
			`
	if err := u.db.QueryRow(ctx, query, id).Scan(&user.ID,
		&user.Name,
		&user.Email,
		&user.Role,
		&user.PasswordHash,
		&user.CreatedAt,); err != nil {
		return nil, err
	}
	return &user, nil

}
