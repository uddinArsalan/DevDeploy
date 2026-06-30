package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/moby/moby/client"
	"github.com/uddinArsalan/devdeploy/internals/adapters/cache"
	queue "github.com/uddinArsalan/devdeploy/internals/adapters/messenger"
	"github.com/uddinArsalan/devdeploy/internals/db"
	"github.com/uddinArsalan/devdeploy/internals/handlers"
	"github.com/uddinArsalan/devdeploy/internals/repository"
	"github.com/uddinArsalan/devdeploy/internals/services"
	"github.com/uddinArsalan/devdeploy/internals/sse"
	"github.com/uddinArsalan/devdeploy/internals/sse/observer"
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

	cache, err := cache.NewRedisClient(ctx)
	if err != nil {
		log.Fatalf("Redis client error %v\n", err)
	}
	sse := sse.NewSSE()
	observers := []observer.Observer{sse}

	projectRepo := repository.NewProjectRepo(dbClient)
	deployRepo := repository.NewDeploymentRepo(dbClient)
	envRepo := repository.NewEnvRepo(dbClient)

	deployService := services.NewDeployService(newClient, *projectRepo, *deployRepo, queue, cache)
	projectService := services.NewProjectService(*projectRepo)
	envService := services.NewEnvService(*envRepo)
	proxyService := services.NewProxyService(cache)
	logStreamService := services.NewLogService(cache, sse, observers)

	proxyHandler := handlers.NewProxyHandler(proxyService)
	projectHandler := handlers.NewProjectHandler(projectService)
	deployHandler := handlers.NewDeployHandler(deployService)
	logStreamHandler := handlers.NewLogHandler(logStreamService)
	envHandler := handlers.NewEnvHandler(envService)

	mux.Handle("/", http.HandlerFunc(proxyHandler.ReverseHandler))

	mux.Handle("POST /project", http.HandlerFunc(projectHandler.CreateProject))

	mux.Handle("GET /stream/{deployID}", http.HandlerFunc(logStreamHandler.StreamLogsHandler))

	mux.Handle("POST /projects/{projectID}/deployments", http.HandlerFunc(deployHandler.Deploy))
	mux.Handle("POST /deployments/{deployID}/start", http.HandlerFunc(deployHandler.StartDeploy))
	mux.Handle("POST /deployments/{deployID}/stop", http.HandlerFunc(deployHandler.StopDeploy))

	mux.Handle("GET /projects/{projectID}/envs", http.HandlerFunc(envHandler.GetProjectEnvs))
	mux.Handle("POST /projects/{projectID}/envs", http.HandlerFunc(envHandler.CreateEnvs))

	server := &http.Server{
		Addr:    ":3000",
		Handler: mux,
	}
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatalf("Server Exit %v", err)
		}
	}()

	log.Println("Server listening on port 3000")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)
	signal.Notify(sigChan, os.Interrupt)

	sign := <-sigChan
	log.Printf("Gracefully Shutdown , Received Signal : %v", sign)

	ctx, cancel = context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	server.Shutdown(ctx)
}
