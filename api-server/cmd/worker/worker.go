package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"github.com/uddinArsalan/devdeploy/internals/adapters/cache"
	queue "github.com/uddinArsalan/devdeploy/internals/adapters/messenger"
	"github.com/uddinArsalan/devdeploy/internals/domain"
	"github.com/uddinArsalan/devdeploy/internals/repository"
	"github.com/uddinArsalan/devdeploy/internals/utils"
)

type DeployWorker struct {
	Id         int
	wg         *sync.WaitGroup
	client     *client.Client
	portMap    *utils.PortMap
	deployRepo *repository.DeploymentRepository
	envRepo    *repository.EnvRepo
	queue      queue.Queue
	cache      cache.Cache
}

func (w *DeployWorker) DeployBuildWorker(ctx context.Context) {
	defer w.wg.Done()
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Stopping deploy worker %d\n", w.Id)
			return
		default:
			consumer, err := w.queue.NewConsumer(ctx)
			if err != nil {
				fmt.Printf("Error creating consumer %v\n", err.Error())
				return
			}
			job, delivery, err := consumer.ConsumeMessage(ctx)

			if err != nil {
				fmt.Printf("Error consuming message %v\n", err.Error())
				continue
			}

			if err := w.publishStatus(ctx, job.DeployID, domain.StatusBuilding); err != nil {
				fmt.Printf("Error updating deployment status %v\n", err.Error())
				_ = delivery.Retry(ctx)
				continue
			}

			err = w.processBuildJob(ctx, job)
			if err != nil {
				fmt.Printf("Error processing build job by worker %d err %v\n", w.Id, err)
				_ = w.deployRepo.UpdateDeploymentStatus(ctx, job.DeployID, domain.StatusFailed)

				_ = w.publishStatus(ctx, job.DeployID, domain.StatusFailed)

				_ = delivery.Retry(ctx)
				continue
			}
			if err = delivery.Ack(ctx); err != nil {
				fmt.Printf("Error accepting the message %v", err)
				continue
			}
		}
	}
}

func (w *DeployWorker) processBuildJob(ctx context.Context, job domain.BuildJob) error {
	fmt.Printf("Processing worker by job %d\n", w.Id)

	imageTag := os.Getenv("IMAGE_TAG")

	// need to validate job payload

	envs, err := w.envRepo.GetProjectEnvs(ctx, job.ProjectID)
	if err != nil {
		return err
	}
	var isPortSpecified bool
	var containerEnvs []string
	var appPort = 8000
	for _, env := range envs {
		value, err := utils.Decrypt(env.EncryptedValue)
		if err != nil {
			continue
		}
		if env.Key == "PORT" {
			p, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("invalid PORT value: %q", value)
			}
			appPort = p
			isPortSpecified = true
		}
		containerEnvs = append(containerEnvs, fmt.Sprintf("%v=%v", env.Key, value))
	}
	if !isPortSpecified {
		containerEnvs = append(containerEnvs, fmt.Sprintf("PORT=%v", appPort))
	}
	tcpPort, _ := network.PortFrom(uint16(appPort), "tcp")

	port := w.portMap.GetPort()
	if port == -1 {
		return errors.New("No available ports to listen to")
	}

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

	go w.streamLogs(ctx, res.ID, job.DeployID)

	waitRes := w.client.ContainerWait(ctx, res.ID, client.ContainerWaitOptions{})

	select {
	case result := <-waitRes.Result:
		if result.StatusCode != 0 {
			return fmt.Errorf("builder container exited with code %d", result.StatusCode)
		}
	case err := <-waitRes.Error:
		return fmt.Errorf("builder container error: %v", err)
	}

	// APPLICATION CONTAINER :
	finalRes, err := w.client.ContainerCreate(ctx, client.ContainerCreateOptions{
		Image: fmt.Sprintf("deployment-image-%v", job.ProjectID),
		Config: &container.Config{
			Env:          containerEnvs,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          true,
		},
		HostConfig: &container.HostConfig{
			PortBindings: network.PortMap{
				tcpPort: []network.PortBinding{
					{
						HostPort: strconv.Itoa(port),
					},
				},
			},
		},
	})
	if err != nil {
		fmt.Printf("Error creating deployment container %v\n", err)
		return err
	}
	_, err = w.client.ContainerStart(ctx, finalRes.ID, client.ContainerStartOptions{})
	if err != nil {
		fmt.Printf("Error starting deployment container %v", err)
		return err
	}

	go w.streamLogs(ctx, finalRes.ID, job.DeployID)

	if err := w.cache.SetHostName(ctx, job.Hostname, port); err != nil {
		fmt.Printf("Error setting hostname %v", err)
		return err
	}

	if err := w.publishStatus(ctx, job.DeployID, domain.StatusRunning); err != nil {
		fmt.Printf("Error updating deployment status %v\n", err.Error())
		return err
	}
	return w.deployRepo.UpdateDeploymentRunning(ctx, port, finalRes.ID, domain.StatusRunning, job.DeployID)
}

func (w *DeployWorker) streamLogs(ctx context.Context, containerID string, deployID int64) {
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
		if scanner.Err() != nil {
			fmt.Printf("Error getting logs %v", err)
			return
		}
		for scanner.Scan() {
			w.cache.AppendLogsAndStatus(ctx, domain.Log, scanner.Text(), deployID)
		}
	}()
}

func (w *DeployWorker) publishStatus(ctx context.Context, deployID int64, status domain.DeploymentStatus) error {
	if err := w.cache.AppendLogsAndStatus(ctx, domain.Status, string(status), deployID); err != nil {
		return err
	}
	return w.cache.SetStatus(ctx, deployID, status)
}
