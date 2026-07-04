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

func (e *EnvRepo) UpdateEnvs(ctx context.Context, projectID int64, updatedEnvs []domain.UpdateEnv) error {
	if len(updatedEnvs) == 0 {
		return nil
	}
	var values []string
	var args []any

	for i, env := range updatedEnvs {
		values = append(values, fmt.Sprintf("($%d::bigint, $%d::bytea)", 2*i+2, 2*i+3),)
		args = append(args, env.ID, env.EncryptedValue)
	}
	query := fmt.Sprintf(`
			UPDATE project_env_vars p
			SET encrypted_value = u.encrypted_value,
				updated_at = NOW()
				FROM (
    				VALUES %v
					) AS u(id, encrypted_value)
			WHERE p.project_id = $1
  					AND p.id = u.id;
			`, strings.Join(values, ","))

	args = append([]any{projectID}, args...)

	_, err := e.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	return nil
}

func (e *EnvRepo) DeleteEnv(ctx context.Context, projectID int64, id int64) error {
	query := `
	DELETE FROM project_env_vars
		WHERE id = $1
  			AND project_id = $2;
	`
	_, err := e.db.Exec(ctx, query, id, projectID)
	if err != nil {
		return err
	}
	return nil
}
