package handlers

import (
	"encoding/json"
	"errors"
	"log"
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
		Path:     "/",
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
	if err := a.authService.CreateUser(r.Context(), domain.CreateUser{
		Name:     userReqDTO.Name,
		Email:    userReqDTO.Email,
		Password: userReqDTO.Password,
	}); err != nil {
		utils.FAIL(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	utils.SUCCESS(w, http.StatusCreated, "user created successfully", nil)
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
			utils.FAIL(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		utils.FAIL(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	SetCookie(w, r, "access_token", token.AccessToken, services.AccessTokenTTL)
	SetCookie(w, r, "refresh_token", token.RefreshToken, services.RefreshTokenTTL)

	utils.SUCCESS(w, http.StatusOK, "user login successfully", nil)
}

func (a *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := r.Cookie("refresh_token")
	if err != nil {
		utils.FAIL(w, http.StatusBadRequest, "invalid cookies")
		return
	}
	token, err := a.authService.Refresh(r.Context(), refreshToken.Value)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidToken):
			utils.FAIL(w, http.StatusUnauthorized, "invalid refresh token")

		case errors.Is(err, services.ErrExpiredToken):
			utils.FAIL(w, http.StatusUnauthorized, "refresh token expired")

		case errors.Is(err, services.ErrRefreshTokenRevoked):
			utils.FAIL(w, http.StatusUnauthorized, "refresh token revoked")

		default:
			log.Printf("refresh error: %v", err)
			utils.FAIL(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	SetCookie(w, r, "access_token", token.AccessToken, services.AccessTokenTTL)
	SetCookie(w, r, "refresh_token", token.RefreshToken, services.RefreshTokenTTL)

	utils.SUCCESS(w, http.StatusAccepted, "tokens refreshed successfully", nil)
}
