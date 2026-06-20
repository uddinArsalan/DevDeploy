package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"github.com/uddinArsalan/devdeploy/internals/domain"
	"github.com/uddinArsalan/devdeploy/internals/handlers/dto"
)

type DeployService struct {
	client *client.Client
}

func NewDeployService(client *client.Client) *DeployService {
	return &DeployService{
		client: client,
	}
}

func (ds *DeployService) Deploy(ctx context.Context, imageTag string, url dto.UserURLReqDTO) (*domain.DeployResponse, error) {
	var projectID = 123 // needs to be unique per project
	var dynamicPort = "8081"
	port, _ := network.PortFrom(8000, "tcp")

	// BUILDER CONTAINER :
	// this container will stop after building the image
	res, err := ds.client.ContainerCreate(ctx, client.ContainerCreateOptions{
		Image: imageTag,
		Config: &container.Config{
			Env: []string{fmt.Sprintf("GIT_URL=%s", url.GitURL), fmt.Sprintf("PROJECT_ID=%v", projectID)},
			Tty: true,
			
		},
		HostConfig: &container.HostConfig{
			Binds: []string{"/var/run/docker.sock.raw:/var/run/docker.sock"},
		},
	})

	if err != nil {
		fmt.Printf("Error creating builder container %v", err)
		return nil, err
	}
	_, err = ds.client.ContainerStart(ctx, res.ID, client.ContainerStartOptions{})
	if err != nil {
		fmt.Printf("Error starting builder container %v", err)
		return nil, err
	}

	buildLogs, err := ds.client.ContainerLogs(ctx, res.ID, client.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})

	if err != nil {
		fmt.Printf("Error getting logs %v", err)
		return nil, err
	}
	var buildLogsBuilder strings.Builder
	for {
		buff := make([]byte, 512)

		n, err := buildLogs.Read(buff)
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		buildLogsBuilder.Write(buff[:n])
	}

	fmt.Println(buildLogsBuilder.String())

	waitRes := client.APIClient.ContainerWait(ds.client, ctx, res.ID, client.ContainerWaitOptions{})

	select {
	case <-waitRes.Result:
		fmt.Print("Build container completed successfully")
	case <-waitRes.Error:
		fmt.Printf("Build container error %v", waitRes.Error)
	}

	// APPLICATION CONTAINER :
	finalRes, err := ds.client.ContainerCreate(ctx, client.ContainerCreateOptions{
		Image: fmt.Sprintf("deployment-image-%v", projectID),
		Config: &container.Config{
			Env: []string{"PORT=8000"},
		},
		HostConfig: &container.HostConfig{
			PortBindings: network.PortMap{
				port: []network.PortBinding{
					{
						HostPort: dynamicPort, // need to dynamic per deployment
					},
				},
			},
		},
	})
	if err != nil {
		fmt.Printf("Error creating deployment container %v", err)
		return nil, err
	}
	_, err = ds.client.ContainerStart(ctx, finalRes.ID, client.ContainerStartOptions{})
	if err != nil {
		fmt.Printf("Error starting deployment container %v", err)
		return nil, err
	}

	appLogs, err := ds.client.ContainerLogs(ctx, finalRes.ID, client.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})

	if err != nil {
		fmt.Printf("Error getting logs %v", err)
		return nil, err
	}
	var appLogsBuilder strings.Builder
	for {
		buff := make([]byte, 512)

		n, err := appLogs.Read(buff)
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		appLogsBuilder.Write(buff[:n])
	}

	fmt.Println(appLogsBuilder.String())

	// curl -X POST http://localhost:3000/deploy -d '{"git_url" : "https://github.com/hkirat/react-boilerplate" }'
	return &domain.DeployResponse{
		DynamicPort:  dynamicPort,
		DeploymentID: finalRes.ID,
	}, err
}

func (ds *DeployService) StopDeploy(ctx context.Context, deployID string) error {
	_, err := ds.client.ContainerStop(ctx, deployID, client.ContainerStopOptions{})
	if err != nil {
		return errors.New("There was an error stopping deploy request")
	}
	return nil
}
