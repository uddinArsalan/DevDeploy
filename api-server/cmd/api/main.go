package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"github.com/moby/moby/client"
	"github.com/uddinArsalan/devdeploy/internals/adapters/cache"
	queue "github.com/uddinArsalan/devdeploy/internals/adapters/messenger"
	"github.com/uddinArsalan/devdeploy/internals/db"
	"github.com/uddinArsalan/devdeploy/internals/handlers"
	"github.com/uddinArsalan/devdeploy/internals/repository"
	"github.com/uddinArsalan/devdeploy/internals/services"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()
	mux := http.NewServeMux()

	newClient, err := client.New(client.FromEnv)
	if err != nil {
		log.Fatalf("Error setting up docker client %q\n", err)
	}

	dbClient := db.NewDB(ctx)

	queue, err := queue.NewRabbitMQClient(ctx)
	if err != nil {
		log.Fatalf("RabbitMQ client error %v\n", err)
	}

	cache,err := cache.NewRedisClient(ctx)
	if err != nil{
		log.Fatalf("Redis client error %v\n", err)
	}

	projectRepo := repository.NewProjectRepo(dbClient)
	deployRepo := repository.NewDeploymentRepo(dbClient)
	proxyService := services.NewProxyService(cache)

	deployService := services.NewDeployService(newClient, *projectRepo, *deployRepo, queue)
	projectService := services.NewProjectService(*projectRepo)

	projectHandler := handlers.NewProjectHandler(projectService)
	deployHandler := handlers.NewDeployHandler(deployService)
	proxyHandler := handlers.NewProxyHandler(proxyService)

	mux.Handle("POST /deploy", http.HandlerFunc(deployHandler.Deploy))
	mux.Handle("POST /project", http.HandlerFunc(projectHandler.CreateProject))
	mux.Handle("GET /", http.HandlerFunc(proxyHandler.ReverseHandler))

	server := &http.Server{
		Addr:    ":3000",
		Handler: mux,
	}
	server.ListenAndServe()
}
