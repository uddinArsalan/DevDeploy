package services

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/moby/moby/client"
	"github.com/sio/coolname"
	"github.com/uddinArsalan/devdeploy/internals/domain"
	"github.com/uddinArsalan/devdeploy/internals/handlers/dto"
	"github.com/uddinArsalan/devdeploy/internals/repository"
)

type DeployService struct {
	client  *client.Client
	projectRepo repository.ProjectRepository
	deployRepo repository.DeploymentRepository
}

func NewDeployService(client *client.Client,projectRepo repository.ProjectRepository,deployRepo repository.DeploymentRepository) *DeployService {
	return &DeployService{
		client:  client,
		projectRepo: projectRepo,
		deployRepo : deployRepo,
	}
}

func (ds *DeployService) Deploy(ctx context.Context, imageTag string,projectID int64) (*dto.DeployResponse, error) {
	project,err := ds.projectRepo.GetProjectByID(ctx,projectID)
	if err != nil {
		return nil,err
	}
	slug, err := coolname.Slug()
	if err != nil {
		return nil, err
	}

	dom := os.Getenv("DOMAIN")
	hostName := fmt.Sprintf("%v.%v", slug, dom)


	deployID ,err := ds.deployRepo.CreateDeploymentRecord(ctx,hostName,projectID)
	if err != nil{
		return nil,err
	}

	// Pass job to the worker to process
	job := domain.BuildJob{
		GitURL: project.GitUrl,
		ProjectID: projectID,
		DeployID : deployID,
		Slug: slug,
	}

	return &dto.DeployResponse{
		DeployID: deployID,
		URL: hostName,
	}, err
}

func (ds *DeployService) StopDeploy(ctx context.Context, deployID int64) error {
	deployment,err := ds.deployRepo.GetDeploymentByID(ctx,deployID)
	if err != nil {
		return nil
	}
	_, err = ds.client.ContainerStop(ctx, deployment.ContainerID, client.ContainerStopOptions{})
	if err != nil {
		return errors.New("There was an error stopping deploy request")
	}
	if err = ds.deployRepo.UpdateDeploymentStatus(ctx,deployID,domain.StatusStopped);err != nil{
		return err
	}
	return nil
}
