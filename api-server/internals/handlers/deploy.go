package handlers

import (
	"net/http"
	"strconv"

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
	projectID := r.PathValue("projectID")
	if projectID == "" {
		utils.FAIL(w, http.StatusBadRequest, "missing project id")
		return
	}
	projectIDInt, err := strconv.ParseInt(projectID, 10, 64)

	if err != nil {
		utils.FAIL(w, http.StatusBadRequest, "invalid project id")
		return
	}

	deployRes, err := h.ds.Deploy(r.Context(), projectIDInt)

	if err != nil {
		utils.FAIL(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SUCCESS(w, http.StatusOK, "Deploy started, The app will be available at the following url", dto.DeployResponse{
		DeployID: deployRes.DeployID,
		URL:      deployRes.URL,
	})
}

func (h *DeployHandler) StartDeploy(w http.ResponseWriter, r *http.Request) {
	deployID := r.PathValue("deployID")
	if deployID == "" {
		utils.FAIL(w, http.StatusBadRequest, "missing deploy id")
		return
	}
	deployIDInt, err := strconv.ParseInt(deployID, 10, 64)

	if err != nil {
		utils.FAIL(w, http.StatusBadRequest, "invalid deploy id")
		return
	}
	err = h.ds.StartDeploy(r.Context(), deployIDInt)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SUCCESS(w, http.StatusAccepted, "Application started", nil)
}

func (h *DeployHandler) StopDeploy(w http.ResponseWriter, r *http.Request) {
	deployID := r.PathValue("deployID")
	if deployID == "" {
		utils.FAIL(w, http.StatusBadRequest, "missing deploy id")
		return
	}
	deployIDInt, err := strconv.ParseInt(deployID, 10, 64)

	if err != nil {
		utils.FAIL(w, http.StatusBadRequest, "invalid deploy id")
		return
	}

	err = h.ds.StopDeploy(r.Context(), deployIDInt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SUCCESS(w, http.StatusAccepted, "Deploy stopped", nil)
}
