package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/moby/moby/client"
	"github.com/uddinArsalan/devdeploy/internals/handlers"
	"github.com/uddinArsalan/devdeploy/internals/services"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	mux := http.NewServeMux()
	newClient, _ := client.New(client.FromEnv)
	deployService := services.NewDeployService(newClient)
	deployHandler := handlers.NewDeployHandler(deployService)
	mux.Handle("POST /deploy", http.HandlerFunc(deployHandler.Deploy))
	server := &http.Server{
		Addr:    ":3000",
		Handler: mux,
	}
	server.ListenAndServe()
}
