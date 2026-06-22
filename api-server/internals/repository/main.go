package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) UserRepository {
	return UserRepository{
		db: db,
	}
}

func (repo *UserRepository) CreateProject(ctx context.Context, name string, gitUrl string) error {
	query := `INSERT INTO projects (name,git_url) VALUES ($1,$2)`
	if _, err := repo.db.Exec(ctx, query, name, gitUrl); err != nil {
		return err
	}
	return nil
}

func (repo *UserRepository) CreateDeploymentRecord(ctx context.Context, hostname string, projectID int64) error {
	query := `INSERT INTO deployments (project_id,hostname) VALUES ($1,$2)`
	if _, err := repo.db.Exec(ctx, query, projectID, hostname); err != nil {
		return err
	}
	return nil
}
