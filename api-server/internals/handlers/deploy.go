package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	port, _ := network.PortFrom(5173,"tcp")
	err := json.NewDecoder(r.Body).Decode(&url)
	if err != nil{
		http.Error(w,"Invalid url",http.StatusBadRequest)
		return
	}
	res,err := h.client.ContainerCreate(r.Context(),client.ContainerCreateOptions{
		Image: "devdeploy-image:9",
		Config: &container.Config{
			Env: []string{fmt.Sprintf("GIT_URL=%s",url.GitURL)},
		},
		HostConfig: &container.HostConfig{
			PortBindings: network.PortMap{
				port : []network.PortBinding{
					{
						HostPort: "5173",
					},
				},
			},
		},
	})
	if err != nil{
		fmt.Printf("Error creating container %v",err)
		http.Error(w,err.Error(),http.StatusInternalServerError)
		return
	}
	_,err = h.client.ContainerStart(r.Context(),res.ID,client.ContainerStartOptions{

	})
	if err != nil{
		fmt.Printf("Error starting container %v",err)
		http.Error(w,err.Error(),http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := map[string]string{
		"message": "success",
		"url" : "http://localhost:5173",
	}

	json.NewEncoder(w).Encode(resp)
}