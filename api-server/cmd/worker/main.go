package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
	"github.com/moby/moby/client"
	"github.com/uddinArsalan/devdeploy/internals/adapters/cache"
	queue "github.com/uddinArsalan/devdeploy/internals/adapters/messenger"
	"github.com/uddinArsalan/devdeploy/internals/db"
	"github.com/uddinArsalan/devdeploy/internals/repository"
	"github.com/uddinArsalan/devdeploy/internals/utils"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbClient := db.NewDB(ctx)

	queue, err := queue.NewRabbitMQClient(ctx)
	if err != nil {
		log.Fatalf("RabbitMQ client error %v", err)
	}

	newClient, err := client.New(client.FromEnv)
	if err != nil {
		log.Fatalf("Error setting up docker client %q\n", err)
	}

	cache,err := cache.NewRedisClient(ctx)
	if err != nil{
		log.Fatalf("Redis client error %v\n", err)
	}

	portMap := utils.NewPortMap(10000, 20000)

	deployRepo := repository.NewDeploymentRepo(dbClient)

	numOfWorkers := os.Getenv("NUM_OF_WORKERS")
	numOfWorkersInt, err := strconv.Atoi(numOfWorkers)
	if err != nil {
		log.Fatalf("Num of workers not set %q", err)
	}

	wg := sync.WaitGroup{}
	dispatcher := NewDispatcher(&wg, ctx, numOfWorkersInt, newClient, portMap, deployRepo, queue,cache)
	dispatcher.Start()
	wg.Wait()
}
