package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/uddinArsalan/devdeploy/internals/domain"
)

type DeploymentRepository struct {
	db *pgxpool.Pool
}

func NewDeploymentRepo(db *pgxpool.Pool) *DeploymentRepository {
	return &DeploymentRepository{
		db: db,
	}
}

var ErrDeploymentNotFound = errors.New("deployment not found")

func (repo *DeploymentRepository) CreateDeploymentRecord(ctx context.Context, hostname string, projectID int64) (int64, error) {
	query :=
		"INSERT INTO deployments (project_id,hostname) VALUES ($1,$2) RETURNING id;"
	var deployID int64
	if err := repo.db.QueryRow(ctx, query, projectID, hostname).Scan(&deployID); err != nil {
		return -1, err
	}
	return deployID, nil
}

func (repo *DeploymentRepository) UpdateDeploymentStatus(ctx context.Context, deployID int64, status domain.DeploymentStatus) error {
	query :=
		`UPDATE deployments
			SET status = $1,
					updated_at = NOW()
				WHERE id = $2;
			`
	if _, err := repo.db.Exec(ctx, query, status, deployID); err != nil {
		return err
	}
	return nil
}

func (repo *DeploymentRepository) GetDeploymentByID(ctx context.Context, deployID int64) (domain.Deployment, error) {
	query := `
		SELECT
    		id,
    		project_id,
  			hostname,
    		port,
    		container_id,
    		status,
    		retry_count,
    		created_at,
    		updated_at
		FROM deployments
		WHERE id = $1
`
	var deployment domain.Deployment
	if err := repo.db.QueryRow(ctx, query, deployID).Scan(
		&deployment.ID,
		&deployment.ProjectID,
		&deployment.HostName,
		&deployment.Port,
		&deployment.ContainerID,
		&deployment.Status,
		&deployment.RetryCount,
		&deployment.CreatedAt,
		&deployment.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Deployment{}, ErrDeploymentNotFound
		}
		return domain.Deployment{}, err
	}
	return deployment, nil
}
