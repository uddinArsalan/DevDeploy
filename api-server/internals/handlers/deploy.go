package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
)

type UserURL struct{
	GitURL string `json:"git_url"`
}

type DeployHandler struct{
	client *client.Client
}

func NewClient(client *client.Client) *DeployHandler{
	return &DeployHandler{
		client: client,
	}
}

func(h *DeployHandler) Deploy(w http.ResponseWriter,r *http.Request){
	var url UserURL
	var imageTag = os.Getenv("IMAGE_TAG")
	var projectID = 123 // needs to be unique per project
	var dynamicPort = "8080"
	port, _ := network.PortFrom(5173,"tcp")
	err := json.NewDecoder(r.Body).Decode(&url)
	if err != nil{
		http.Error(w,"Invalid url",http.StatusBadRequest)
		return
	}
	res,err := h.client.ContainerCreate(r.Context(),client.ContainerCreateOptions{
		Image: imageTag,
		Config: &container.Config{
			Env: []string{fmt.Sprintf("GIT_URL=%s",url.GitURL),fmt.Sprintf("PROJECT_ID=%v",projectID)},
			
		},
		HostConfig: &container.HostConfig{
			PortBindings: network.PortMap{
				port : []network.PortBinding{
					{
						HostPort: dynamicPort,
					},
				},
			},
			Binds: []string{"/var/run/docker.sock.raw:/var/run/docker.sock"},
		},
	})
	if err != nil{
		fmt.Printf("Error creating builder container %v",err)
		http.Error(w,err.Error(),http.StatusInternalServerError)
		return
	}
	_,err = h.client.ContainerStart(r.Context(),res.ID,client.ContainerStartOptions{

	})
	if err != nil{
		fmt.Printf("Error starting builder container %v",err)
		http.Error(w,err.Error(),http.StatusInternalServerError)
		return
	}

	finalRes,err := h.client.ContainerCreate(r.Context(),client.ContainerCreateOptions{
		Image: fmt.Sprintf("deployment-%v",projectID),
		HostConfig: &container.HostConfig{
			PortBindings: network.PortMap{
				port : []network.PortBinding{
					{
						HostPort: dynamicPort, // need to dynamic per deployment
					},
				},
			},
		},
	})
	if err != nil{
		fmt.Printf("Error creating deployment container %v",err)
		http.Error(w,err.Error(),http.StatusInternalServerError)
		return
	}
	_,err = h.client.ContainerStart(r.Context(),finalRes.ID,client.ContainerStartOptions{

	})
	if err != nil{
		fmt.Printf("Error starting deployment container %v",err)
		http.Error(w,err.Error(),http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := map[string]string{
		"message": "Deploy started, The app will be available at the following url",
		"url" : fmt.Sprintf("http://localhost:%v",dynamicPort),
	}

	json.NewEncoder(w).Encode(resp)
}