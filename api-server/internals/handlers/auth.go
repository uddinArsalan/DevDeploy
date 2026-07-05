package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/uddinArsalan/devdeploy/internals/domain"
	"github.com/uddinArsalan/devdeploy/internals/dto"
	"github.com/uddinArsalan/devdeploy/internals/services"
	"github.com/uddinArsalan/devdeploy/internals/utils"
)

type AuthHandler struct{
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler{
	return &AuthHandler{
		authService,
	}
}

func (a *AuthHandler) Register(w http.ResponseWriter,r *http.Request){
	var userReqDTO dto.CreateUserDTO
	if err := json.NewDecoder(r.Body).Decode(&userReqDTO);err != nil{
		utils.FAIL(w,http.StatusBadRequest,"invalid user req")
		return
	}
	a.authService.CreateUser(r.Context(),domain.CreateUser{
		Name: userReqDTO.Name,
		Email: userReqDTO.Email,
	})
}

func Login(w http.ResponseWriter,r *http.Request){
	
}