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
	"github.com/uddinArsalan/devdeploy/internals/adapters/token"
	"github.com/uddinArsalan/devdeploy/internals/db"
	"github.com/uddinArsalan/devdeploy/internals/handlers"
	"github.com/uddinArsalan/devdeploy/internals/middlewares"
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
	userRepo := repository.NewUserRepo(dbClient)
	refreshTokenRepo := repository.NewRefreshTokenRepo(dbClient)
	jwtTokenStore := token.NewJwtToken()

	middleware := middlewares.NewAuthMiddleware(jwtTokenStore)

	public := http.NewServeMux()
	private := http.NewServeMux()

	deployService := services.NewDeployService(newClient, projectRepo, deployRepo, queue, cache)
	projectService := services.NewProjectService(projectRepo)
	envService := services.NewEnvService(envRepo, projectRepo)
	proxyService := services.NewProxyService(cache)
	logStreamService := services.NewLogService(cache, sse, observers, deployRepo, projectRepo)
	authService := services.NewAuthService(userRepo, refreshTokenRepo, jwtTokenStore)

	proxyHandler := handlers.NewProxyHandler(proxyService)
	projectHandler := handlers.NewProjectHandler(projectService)
	deployHandler := handlers.NewDeployHandler(deployService)
	logStreamHandler := handlers.NewLogHandler(logStreamService)
	envHandler := handlers.NewEnvHandler(envService)
	authHandler := handlers.NewAuthHandler(authService)

	private.Handle("POST /project", http.HandlerFunc(projectHandler.CreateProject))

	private.Handle("GET /stream/{deployID}", http.HandlerFunc(logStreamHandler.StreamLogsHandler))

	private.Handle("GET /projects/{projectID}/deployments", http.HandlerFunc(deployHandler.GetDeployments))
	private.Handle("POST /projects/{projectID}/deployments", http.HandlerFunc(deployHandler.Deploy))
	private.Handle("POST /deployments/{deployID}/start", http.HandlerFunc(deployHandler.StartDeploy))
	private.Handle("POST /deployments/{deployID}/stop", http.HandlerFunc(deployHandler.StopDeploy))

	private.Handle("GET /projects/{projectID}/envs", http.HandlerFunc(envHandler.GetProjectEnvs))
	private.Handle("POST /projects/{projectID}/envs", http.HandlerFunc(envHandler.CreateEnvs))
	private.Handle("PATCH /projects/{projectID}/envs", http.HandlerFunc(envHandler.UpdateEnvs))
	private.Handle("DELETE /projects/{projectID}/envs/{id}", http.HandlerFunc(envHandler.DeleteEnv))

	public.Handle("POST /auth/signup", http.HandlerFunc(authHandler.Register))
	public.Handle("POST /auth/signin", http.HandlerFunc(authHandler.Login))
	public.Handle("POST /auth/refresh", http.HandlerFunc(authHandler.Refresh))

	mainMux := http.NewServeMux()

	mainMux.Handle("/api/v1/public/", http.StripPrefix("/api/v1/public", public))

	mainMux.Handle("/api/v1/private/", middleware.AuthMiddleware(http.StripPrefix("/api/v1/private", private)))

	mainMux.Handle("/", http.HandlerFunc(proxyHandler.ReverseHandler))

	server := &http.Server{
		Addr:    ":3000",
		Handler: mainMux,
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
