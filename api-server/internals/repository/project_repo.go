package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/uddinArsalan/devdeploy/internals/domain"
)

type ProjectRepository struct {
	db *pgxpool.Pool
}

func NewProjectRepo(db *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{
		db: db,
	}
}

var ErrProjectNotFound = errors.New("project not found")

func (repo *ProjectRepository) CreateProject(ctx context.Context, name string, gitUrl string) (int64, error) {
	query :=
		"INSERT INTO projects (name,git_url) VALUES ($1,$2) RETURNING id;"
	var projectID int64
	if err := repo.db.QueryRow(ctx, query, name, gitUrl).Scan(&projectID); err != nil {
		return -1, err
	}
	return projectID, nil
}

func (repo *ProjectRepository) GetProjectByID(ctx context.Context, projectID int64) (domain.Project, error) {
	query := "SELECT id,name,git_url,created_at FROM projects WHERE id=$1"
	var project domain.Project
	if err := repo.db.QueryRow(ctx, query, projectID).Scan(&project.ID, &project.Name, &project.GitUrl, &project.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Project{}, ErrProjectNotFound
		}
		return domain.Project{}, err
	}
	return project, nil
}
