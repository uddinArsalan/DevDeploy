package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/uddinArsalan/devdeploy/internals/domain"
	"github.com/uddinArsalan/devdeploy/internals/dto"
	"github.com/uddinArsalan/devdeploy/internals/services"
	"github.com/uddinArsalan/devdeploy/internals/utils"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService,
	}
}

func SetCookie(w http.ResponseWriter, r *http.Request, name string, value string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})
}

func (a *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var userReqDTO dto.CreateUserDTO
	if err := json.NewDecoder(r.Body).Decode(&userReqDTO); err != nil {
		utils.FAIL(w, http.StatusBadRequest, "invalid user req")
		return
	}
	a.authService.CreateUser(r.Context(), domain.CreateUser{
		Name:  userReqDTO.Name,
		Email: userReqDTO.Email,
	})
}

func (a *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var loginReq dto.LoginUserDTO
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		utils.FAIL(w, http.StatusBadRequest, "invalid user req")
		return
	}
	token, err := a.authService.LoginUser(r.Context(), domain.LoginUser{
		Email:    loginReq.Email,
		Password: loginReq.Password,
	})
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			utils.FAIL(w, http.StatusBadRequest, "invalid credentials")
			return
		}
		utils.FAIL(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	SetCookie(w, r, "access_token", token.AccessToken, 15*time.Minute)
	SetCookie(w, r, "refresh_token", token.RefreshToken, 7*24*time.Hour)

	utils.SUCCESS(w, http.StatusAccepted, "user login successfully", nil)
}
