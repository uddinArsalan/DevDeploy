package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/uddinArsalan/devdeploy/internals/handlers/dto"
	"github.com/uddinArsalan/devdeploy/internals/services"
	"github.com/uddinArsalan/devdeploy/internals/utils"
)

type ProjectHandler struct {
	ps *services.ProjectService
}

func NewProjectHandler(ps *services.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		ps,
	}
}

func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var projectReq dto.ProjectReqDTO
	err := json.NewDecoder(r.Body).Decode(&projectReq)
	fmt.Printf("Project %v\n",err)
	if err != nil {
		utils.FAIL(w, http.StatusBadRequest, "Invalid Project Details")
		return
	}

	projectRes, err := h.ps.CreateProject(r.Context(), projectReq.Name, projectReq.GitUrl)
	if err != nil {
		utils.FAIL(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	utils.SUCCESS(w, http.StatusAccepted, "project created successfully", dto.ProjectResDTO{
		ProjectID: projectRes.ProjectID,
	})
}
