package services

import (
	"context"

	"github.com/uddinArsalan/devdeploy/internals/domain"
	"github.com/uddinArsalan/devdeploy/internals/repository"
)

type EnvService struct {
	envRepo repository.EnvRepo
}

func NewEnvService(envRepo repository.EnvRepo) *EnvService {
	return &EnvService{
		envRepo: envRepo,
	}
}

func (e *EnvService) CreateEnvs(ctx context.Context, envs []domain.Env) error {
	return e.envRepo.InsertEnvs(ctx, envs)
}

func (e *EnvService) GetProjectEnvs(ctx context.Context, projectID int64) ([]domain.Env, error) {
	return e.envRepo.GetProjectEnvs(ctx, projectID)
}
