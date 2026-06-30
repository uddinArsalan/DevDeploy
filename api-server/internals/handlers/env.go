package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/uddinArsalan/devdeploy/internals/handlers/dto"
	"github.com/uddinArsalan/devdeploy/internals/handlers/dto/mapping"
	"github.com/uddinArsalan/devdeploy/internals/services"
	"github.com/uddinArsalan/devdeploy/internals/utils"
)

type EnvHandler struct {
	envService *services.EnvService
}

func NewEnvHandler(envService *services.EnvService) *EnvHandler {
	return &EnvHandler{
		envService: envService,
	}
}

func (e *EnvHandler) CreateEnvs(w http.ResponseWriter, r *http.Request) {
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

	var userEnvs []dto.Env
	err = json.NewDecoder(r.Body).Decode(&userEnvs)
	if err != nil {
		utils.FAIL(w, http.StatusBadRequest, "Invalid envs")
		return
	}

	if err = e.envService.CreateEnvs(r.Context(), mapping.ToEnvDomain(projectIDInt, userEnvs)); err != nil {
		fmt.Printf("\nError creating envs %v\n", err)
		utils.FAIL(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	utils.SUCCESS(w, http.StatusOK, "envs created successfully", nil)
}

func (e *EnvHandler) GetProjectEnvs(w http.ResponseWriter, r *http.Request) {
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
	envs, err := e.envService.GetProjectEnvs(r.Context(), projectIDInt)
	if err != nil {
		fmt.Printf("\nError fetching envs %v\n", err)
		utils.FAIL(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	utils.SUCCESS(w, http.StatusOK, "envs fetched successfully", mapping.ToEnvsReponse(envs))
}
