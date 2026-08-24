package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/uddinArsalan/devdeploy/internals/dto"
	"github.com/uddinArsalan/devdeploy/internals/middlewares"
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
	userClaim , ok := middlewares.UserFromContext(r.Context())
	if !ok {
		utils.FAIL(w, http.StatusUnauthorized, "authentication required")
		return
	}
	var projectReq dto.ProjectReqDTO
	err := json.NewDecoder(r.Body).Decode(&projectReq)
	if err != nil {
		utils.FAIL(w, http.StatusBadRequest, "Invalid Project Details")
		return
	}
	projectRes, err := h.ps.CreateProject(r.Context(),userClaim.UserID, projectReq.Name, projectReq.GitUrl)
	if err != nil {
		utils.FAIL(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	utils.SUCCESS(w, http.StatusOK, "project created successfully", dto.ProjectResDTO{
		ProjectID: projectRes.ProjectID,
	})
}
