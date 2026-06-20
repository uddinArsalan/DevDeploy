package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/uddinArsalan/devdeploy/internals/handlers/dto"
	"github.com/uddinArsalan/devdeploy/internals/services"
	"github.com/uddinArsalan/devdeploy/internals/utils"
)

type DeployHandler struct {
	ds *services.DeployService
}

func NewDeployHandler(ds *services.DeployService) *DeployHandler {
	return &DeployHandler{
		ds,
	}
}

func (h *DeployHandler) Deploy(w http.ResponseWriter, r *http.Request) {
	// the respective code should move into a service
	var url dto.UserURLReqDTO
	var imageTag = os.Getenv("IMAGE_TAG")

	err := json.NewDecoder(r.Body).Decode(&url)

	if err != nil {
		http.Error(w, "Invalid url", http.StatusBadRequest)
		return
	}

	deployRes, err := h.ds.Deploy(r.Context(), imageTag, url)

	if err != nil {
		utils.FAIL(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SUCCESS(w, http.StatusOK, "Deploy started, The app will be available at the following url", dto.DeployResponse{
		Url:      fmt.Sprintf("http://localhost:%v", deployRes.DynamicPort),
		DeployID: deployRes.DeploymentID,
	})
}

func (h *DeployHandler) StopDeploy(w http.ResponseWriter, r *http.Request) {
	var deployIDReq dto.DeployReqDTO
	err := json.NewDecoder(r.Body).Decode(&deployIDReq)
	if err != nil {
		http.Error(w, "Error reading deploy id", http.StatusBadRequest)
		return
	}
	err = h.ds.StopDeploy(r.Context(), deployIDReq.DeployID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SUCCESS(w, http.StatusAccepted, "Deploy stopped", nil)
}
