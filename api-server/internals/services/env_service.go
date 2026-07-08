package services

import (
	"context"

	"github.com/uddinArsalan/devdeploy/internals/domain"
	"github.com/uddinArsalan/devdeploy/internals/repository"
)

type EnvService struct {
	projectRepo *repository.ProjectRepository
	envRepo     *repository.EnvRepo
}

func NewEnvService(envRepo *repository.EnvRepo, projectRepo *repository.ProjectRepository) *EnvService {
	return &EnvService{
		envRepo:     envRepo,
		projectRepo: projectRepo,
	}
}

func (e *EnvService) CreateEnvs(ctx context.Context, envs []domain.Env) error {
	return e.envRepo.InsertEnvs(ctx, envs)
}

func (e *EnvService) GetProjectEnvs(ctx context.Context, userID, projectID int64) ([]domain.Env, error) {
	userProject, err := e.projectRepo.GetProjectByID(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}
	return e.envRepo.GetProjectEnvs(ctx, userProject.ID)
}

func (e *EnvService) UpdateEnvs(ctx context.Context, userID, projectID int64, updatedEnvs []domain.UpdateEnv) error {
	userProject, err := e.projectRepo.GetProjectByID(ctx, userID, projectID)
	if err != nil {
		return err
	}
	return e.envRepo.UpdateEnvs(ctx, userProject.ID, updatedEnvs)
}

func (e *EnvService) DeleteEnv(ctx context.Context, userID, projectID int64, id int64) error {
	userProject, err := e.projectRepo.GetProjectByID(ctx, userID, projectID)
	if err != nil {
		return err
	}
	return e.envRepo.DeleteEnv(ctx, userProject.ID, id)
}
