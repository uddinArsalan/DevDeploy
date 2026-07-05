package services

import (
	"context"

	"github.com/uddinArsalan/devdeploy/internals/domain"
	"github.com/uddinArsalan/devdeploy/internals/repository"
)

type AuthService struct {
	userRepo *repository.UserRepo
}

func NewAuthService(userRepo *repository.UserRepo) *AuthService {
	return &AuthService{
		userRepo,
	}
}

func (a *AuthService) CreateUser(ctx context.Context, userDetails domain.CreateUser) error {
	return a.userRepo.CreateUser(ctx, userDetails)
}
