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
	"github.com/uddinArsalan/devdeploy/internals/utils"
)

type DeployService struct {
	client  *client.Client
	portMap *utils.PortMap
}

func NewDeployService(client *client.Client, portMap *utils.PortMap) *DeployService {
	return &DeployService{
		client:  client,
		portMap: portMap,
	}
}

func (ds *DeployService) Deploy(ctx context.Context, imageTag string, url dto.UserURLReqDTO) (*domain.DeployResponse, error) {
	slug, err := coolname.Slug()
	if err != nil {
		return nil, err
	}

	projectID, err := utils.GenerateRandomID()
	if err != nil {
		return nil, err
	}

	dom := os.Getenv("DOMAIN")
	hostName := fmt.Sprintf("%v.%v", slug, dom)


	// Pass job to the worker to process
	job := domain.BuildJob{
		GitURL: url.GitURL,
		ProjectID: projectID,
		Slug: slug,
	}

	return &domain.DeployResponse{
		Url: hostName,
	}, err
}

func (ds *DeployService) StopDeploy(ctx context.Context, deployID string) error {
	_, err := ds.client.ContainerStop(ctx, deployID, client.ContainerStopOptions{})
	if err != nil {
		return errors.New("There was an error stopping deploy request")
	}
	return nil
}
