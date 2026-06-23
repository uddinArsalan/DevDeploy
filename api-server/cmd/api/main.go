package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"github.com/moby/moby/client"
	"github.com/uddinArsalan/devdeploy/internals/db"
	"github.com/uddinArsalan/devdeploy/internals/handlers"
	"github.com/uddinArsalan/devdeploy/internals/repository"
	"github.com/uddinArsalan/devdeploy/internals/services"
	"github.com/uddinArsalan/devdeploy/internals/utils"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()
	mux := http.NewServeMux()
	portMap := utils.NewPortMap(10000, 20000)
	newClient, _ := client.New(client.FromEnv)
	dbClient := db.NewDB(ctx)
	projectRepo := repository.NewProjectRepo(dbClient)
	deployRepo := repository.NewDeploymentRepo(dbClient)
	proxyService := services.NewProxyService(portMap)
	deployService := services.NewDeployService(newClient, *projectRepo, *deployRepo)
	deployHandler := handlers.NewDeployHandler(deployService)
	proxyHandler := handlers.NewProxyHandler(proxyService)
	mux.Handle("POST /deploy", http.HandlerFunc(deployHandler.Deploy))
	mux.Handle("GET /", http.HandlerFunc(proxyHandler.ReverseHandler))
	server := &http.Server{
		Addr:    ":3000",
		Handler: mux,
	}
	server.ListenAndServe()
}
