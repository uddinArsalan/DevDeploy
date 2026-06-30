package services

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/moby/moby/client"
	"github.com/uddinArsalan/devdeploy/internals/adapters/cache"
	queue "github.com/uddinArsalan/devdeploy/internals/adapters/messenger"
	"github.com/uddinArsalan/devdeploy/internals/domain"
	"github.com/uddinArsalan/devdeploy/internals/handlers/dto"
	"github.com/uddinArsalan/devdeploy/internals/repository"
	"github.com/uddinArsalan/devdeploy/internals/utils"
)

type DeployService struct {
	client      *client.Client
	projectRepo repository.ProjectRepository
	deployRepo  repository.DeploymentRepository
	cache       cache.Cache
	queue       queue.Queue
}

func NewDeployService(
	client *client.Client,
	projectRepo repository.ProjectRepository,
	deployRepo repository.DeploymentRepository,
	queue queue.Queue,
	cache cache.Cache) *DeployService {

	return &DeployService{
		client:      client,
		projectRepo: projectRepo,
		deployRepo:  deployRepo,
		queue:       queue,
		cache:       cache,
	}
}

func (ds *DeployService) Deploy(ctx context.Context, projectID int64) (*dto.DeployResponse, error) {
	project, err := ds.projectRepo.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	randomID, err := utils.GenerateRandomID()
	if err != nil {
		return nil, err
	}
	dom := os.Getenv("DOMAIN")
	hostname := fmt.Sprintf("%s-%s.%s",
		utils.Slugify(project.Name),
		randomID,
		dom,
	)

	deployID, err := ds.deployRepo.CreateDeploymentRecord(ctx, hostname, projectID)
	if err != nil {
		return nil, err
	}

	// Pass job to the worker to process
	job := domain.BuildJob{
		GitURL:    project.GitUrl,
		ProjectID: projectID,
		DeployID:  deployID,
		Hostname:      hostname,
	}

	if err = ds.queue.PublishMessage(ctx, job); err != nil {
		return nil, err
	}

	return &dto.DeployResponse{
		DeployID: deployID,
		URL:      hostname,
	}, err
}

func (ds *DeployService) StopDeploy(ctx context.Context, deployID int64) error {
	deployment, err := ds.deployRepo.GetDeploymentByID(ctx, deployID)
	if err != nil {
		return err
	}
	_, err = ds.client.ContainerStop(ctx, deployment.ContainerID, client.ContainerStopOptions{})
	if err != nil {
		return errors.New("There was an error stopping deploy request")
	}
	if err = ds.deployRepo.UpdateDeploymentStatus(ctx, deployID, domain.StatusStopped); err != nil {
		return err
	}
	return ds.cache.DelHostName(ctx, deployment.HostName)
}

func (ds *DeployService) StartDeploy(ctx context.Context, deployID int64) error {
	deployment, err := ds.deployRepo.GetDeploymentByID(ctx, deployID)

	if err != nil {
		return err
	}
	_, err = ds.client.ContainerStart(ctx, deployment.ContainerID, client.ContainerStartOptions{})

	if err != nil {
		return errors.New("There was an error starting deploy request")
	}
	if err = ds.cache.SetHostName(ctx, deployment.HostName, deployment.Port); err != nil {
		return err
	}

	return nil
}
