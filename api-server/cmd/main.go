package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/moby/moby/client"
	"github.com/uddinArsalan/devdeploy/internals/handlers"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	mux := http.NewServeMux()
	newClient, _ := client.New(client.FromEnv)
	deployHandler := handlers.NewClient(newClient)
	mux.Handle("POST /deploy", http.HandlerFunc(deployHandler.Deploy))
	server := &http.Server{
		Addr:    ":3000",
		Handler: mux,
	}
	server.ListenAndServe()
}
