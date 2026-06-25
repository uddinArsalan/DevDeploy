package services

import (
	"context"

	"github.com/uddinArsalan/devdeploy/internals/domain"
	"github.com/uddinArsalan/devdeploy/internals/repository"
)

type ProjectService struct {
	projectRepo repository.ProjectRepository
}

func NewProjectService(projectRepo repository.ProjectRepository) *ProjectService {
	return &ProjectService{
		projectRepo,
	}
}

func (ps *ProjectService) CreateProject(ctx context.Context, name string, gitUrl string) (domain.CreateProjectResult, error) {
	projectId, err := ps.projectRepo.CreateProject(ctx, name, gitUrl)
	if err != nil {
		return domain.CreateProjectResult{}, err
	}
	return domain.CreateProjectResult{
		ProjectID: projectId,
	}, nil
}
