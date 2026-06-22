package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/moby/moby/client"
	"github.com/uddinArsalan/devdeploy/internals/handlers"
	"github.com/uddinArsalan/devdeploy/internals/services"
	"github.com/uddinArsalan/devdeploy/internals/utils"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	mux := http.NewServeMux()
	portMap := utils.NewPortMap(int64(10000),int64(20000))
	newClient, _ := client.New(client.FromEnv)
	proxyService := services.NewProxyService(portMap)
	deployService := services.NewDeployService(newClient,portMap)
	deployHandler := handlers.NewDeployHandler(deployService)
	proxyHandler := handlers.NewProxyHandler(proxyService)
	mux.Handle("POST /deploy", http.HandlerFunc(deployHandler.Deploy))
	mux.Handle("GET /",http.HandlerFunc(proxyHandler.ReverseHandler))
	server := &http.Server{
		Addr:    ":3000",
		Handler: mux,
	}
	server.ListenAndServe()
}
