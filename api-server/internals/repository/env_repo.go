package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/uddinArsalan/devdeploy/internals/domain"
)

type EnvRepo struct {
	db *pgxpool.Pool
}

func NewEnvRepo(db *pgxpool.Pool) *EnvRepo {
	return &EnvRepo{
		db: db,
	}
}

func (e *EnvRepo) InsertEnvs(ctx context.Context, envArray []domain.Env) error {
	if len(envArray) == 0 {
		return errors.New("no envs")
	}
	var values []string
	var args []any
	for i, env := range envArray {
		values = append(values, fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3))
		args = append(args, env.ProjectID, env.Key, env.EncryptedValue)
	}
	query := `INSERT INTO project_env_vars (project_id,key_name,encrypted_value) VALUES ` + strings.Join(values, ",")
	_, err := e.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	return nil
}

func (e *EnvRepo) GetProjectEnvs(ctx context.Context, projectID int64) ([]domain.Env, error) {
	query := `
	SELECT id, project_id, key_name, encrypted_value, created_at, updated_at
	FROM project_env_vars
	WHERE project_id = $1
`
	rows, err := e.db.Query(ctx, query, projectID)
	if err != nil {
		return []domain.Env{}, err
	}
	defer rows.Close()
	var envArr []domain.Env
	for rows.Next() {
		var env domain.Env
		err = rows.Scan(&env.ID, &env.ProjectID, &env.Key, &env.EncryptedValue, &env.CreatedAt, &env.UpdatedAt)
		if err != nil {
			return []domain.Env{}, fmt.Errorf("error scanning row: %w", err)
		}
		envArr = append(envArr, env)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return envArr, nil
}
