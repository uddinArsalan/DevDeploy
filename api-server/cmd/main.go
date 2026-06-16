package main

import (
	"net/http"
	"github.com/moby/moby/client"
	"github.com/uddinArsalan/devdeploy/internals/handlers"
)

func main() {
	mux := http.NewServeMux()
	newClient,_ := client.New(client.FromEnv)
	deployHandler := handlers.NewClient(newClient)
	mux.Handle("POST /deploy",http.HandlerFunc(deployHandler.Deploy))
	server := &http.Server{
		Addr: ":3000",
		Handler: mux,
	}
	server.ListenAndServe()
}
