package services

import (
	"context"

	"github.com/uddinArsalan/devdeploy/internals/adapters/cache"
	"github.com/uddinArsalan/devdeploy/internals/domain"
	"github.com/uddinArsalan/devdeploy/internals/repository"
	"github.com/uddinArsalan/devdeploy/internals/sse"
	"github.com/uddinArsalan/devdeploy/internals/sse/observer"
)

type LogService struct {
	cache       cache.Cache
	sse         *sse.LogChan
	observers   []observer.Observer
	deployRepo  *repository.DeploymentRepository
	projectRepo *repository.ProjectRepository
}

func NewLogService(
	cache cache.Cache,
	sse *sse.LogChan,
	observers []observer.Observer,
	deployRepo *repository.DeploymentRepository,
	projectRepo *repository.ProjectRepository) *LogService {

	return &LogService{
		cache:       cache,
		sse:         sse,
		observers:   observers,
		deployRepo:  deployRepo,
		projectRepo: projectRepo,
	}
}

func (ls *LogService) StreamLogs(ctx context.Context, userID int64, lastID string, deployID int64) (<-chan domain.LogEvent, error) {
	deployment, err := ls.deployRepo.GetDeploymentByID(ctx, deployID)
	if err != nil {
		return nil, err
	}

	_, err = ls.projectRepo.GetProjectByID(ctx, userID, deployment.ProjectID)
	if err != nil {
		return nil, err
	}
	ch := ls.sse.AddUser(deployID)
	go func() {
		// this runs the continuous XRead loop and notifies via observer
		ls.cache.ReadEntriesFromStream(ctx, lastID, deployID, ls.observers)
		ls.sse.Done(deployID)
	}()
	return ch, nil
}
