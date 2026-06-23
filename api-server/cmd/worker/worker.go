package worker

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"github.com/uddinArsalan/devdeploy/internals/domain"
	"github.com/uddinArsalan/devdeploy/internals/repository"
	"github.com/uddinArsalan/devdeploy/internals/utils"
)

type DeployWorker struct {
	Id           string
	wg           *sync.WaitGroup
	jobBuildChan chan domain.BuildJob
	client       *client.Client
	portMap      *utils.PortMap
	deployRepo   *repository.DeploymentRepository
}

func (w *DeployWorker) DeployBuildWorker(ctx context.Context) {
	defer w.wg.Done()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Stopping deploy worker")
			return
		case job, ok := <-w.jobBuildChan:
			if !ok {
				fmt.Printf("Worker %d: jobs channel closed, exiting\n", w.Id)
				return
			}
			if err := w.deployRepo.UpdateDeploymentStatus(ctx, job.DeployID, domain.StatusBuilding); err != nil {
				fmt.Printf("Error updating deployment status %v", err.Error())
				return
			}
			err := w.processBuildJob(ctx, job)
			if err != nil {
				fmt.Printf("Error processing build job by worker %d", w.Id)
				if err := w.deployRepo.UpdateDeploymentStatus(ctx, job.DeployID, domain.StatusFailed); err != nil {
					fmt.Printf("Error updating deployment status %v", err.Error())
				}
				return
			}
		}
	}
}

func (w *DeployWorker) processBuildJob(ctx context.Context, job domain.BuildJob) error {
	imageTag := os.Getenv("IMAGE_TAG")
	dom := os.Getenv("DOMAIN")
	appPort, _ := network.PortFrom(8000, "tcp")
	port := w.portMap.GetPort()
	if port == -1 {
		return errors.New("No available ports to listen to")
	}

	hostName := fmt.Sprintf("%v.%v", job.Slug, dom)

	// BUILDER CONTAINER :
	// this container will stop after building the image
	res, err := w.client.ContainerCreate(ctx, client.ContainerCreateOptions{
		Image: imageTag,
		Config: &container.Config{
			Env:          []string{fmt.Sprintf("GIT_URL=%s", job.GitURL), fmt.Sprintf("PROJECT_ID=%v", job.ProjectID)},
			AttachStdout: true,
			AttachStderr: true,
			Tty:          true,
		},
		HostConfig: &container.HostConfig{
			Binds: []string{"/var/run/docker.sock.raw:/var/run/docker.sock"},
		},
	})

	if err != nil {
		fmt.Printf("Error creating builder container %v", err)
		return err
	}
	_, err = w.client.ContainerStart(ctx, res.ID, client.ContainerStartOptions{})
	if err != nil {
		fmt.Printf("Error starting builder container %v", err)
		return err
	}

	go w.streamLogs(res.ID)

	waitRes := w.client.ContainerWait(ctx, res.ID, client.ContainerWaitOptions{})

	select {
	case <-waitRes.Result:
		fmt.Print("\nBuild container completed successfully\n")
	case <-waitRes.Error:
		fmt.Printf("Build container error %v", waitRes.Error)
	}

	// APPLICATION CONTAINER :
	finalRes, err := w.client.ContainerCreate(ctx, client.ContainerCreateOptions{
		Image: fmt.Sprintf("deployment-image-%v", job.ProjectID),
		Config: &container.Config{
			Env:          []string{fmt.Sprintf("PORT=%v", appPort.Port())},
			AttachStdout: true,
			AttachStderr: true,
			Tty:          true,
		},
		HostConfig: &container.HostConfig{
			PortBindings: network.PortMap{
				appPort: []network.PortBinding{
					{
						HostPort: string(port),
					},
				},
			},
		},
	})
	if err != nil {
		fmt.Printf("Error creating deployment container %v", err)
		return err
	}
	_, err = w.client.ContainerStart(ctx, finalRes.ID, client.ContainerStartOptions{})
	if err != nil {
		fmt.Printf("Error starting deployment container %v", err)
		return err
	}

	go w.streamLogs(finalRes.ID)

	w.portMap.AssignProjectIDToDomain(job.ProjectID, hostName, finalRes.ID, int(port))
	return nil
}

func (w *DeployWorker) streamLogs(containerID string) {
	go func() {
		logs, err := w.client.ContainerLogs(context.Background(), containerID, client.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
			Timestamps: false,
		})
		if err != nil {
			fmt.Printf("Error getting logs %v", err)
			return
		}
		defer logs.Close()

		scanner := bufio.NewScanner(logs)
		for scanner.Scan() {
			// send logs through a Redis pub/sub
			fmt.Println(scanner.Text())
		}
	}()
}
