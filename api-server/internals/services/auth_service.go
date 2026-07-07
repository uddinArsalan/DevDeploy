package services

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/uddinArsalan/devdeploy/internals/adapters/token"
	"github.com/uddinArsalan/devdeploy/internals/crypto"
	"github.com/uddinArsalan/devdeploy/internals/domain"
	"github.com/uddinArsalan/devdeploy/internals/repository"

	"crypto/subtle"
)

const (
	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 7 * 24 * time.Hour
)

type AuthService struct {
	userRepo         *repository.UserRepo
	refreshTokenRepo *repository.RefreshTokenRepo
	tokenStore       token.TokenStore
}

type Token struct {
	AccessToken  string
	RefreshToken string
}

func NewAuthService(userRepo *repository.UserRepo, refreshTokenRepo *repository.RefreshTokenRepo, tokenStore token.TokenStore) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		tokenStore:       tokenStore,
	}
}

func (a *AuthService) CreateUser(ctx context.Context, userDetails domain.CreateUser) error {
	hashedPassword, err := crypto.HashAndEncodePassword(userDetails.Password)
	if err != nil {
		return err
	}
	return a.userRepo.CreateUser(ctx, userDetails.Name, userDetails.Email, hashedPassword)
}

func (a *AuthService) LoginUser(ctx context.Context, loginDetails domain.LoginUser) (*Token, error) {
	user, err := a.userRepo.FindUserByEmail(ctx, loginDetails.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	//password - check
	parsedPass, err := crypto.ParsePassword(user.PasswordHash)
	if err != nil {
		return nil, err
	}
	params := crypto.Params{
		Memory:      parsedPass.Memory,
		Iterations:  parsedPass.Iterations,
		Parallelism: parsedPass.Parallelism,
		KeyLen:      parsedPass.KeyLen,
	}
	storedHashPass := parsedPass.PasswordHash
	storedSalt := parsedPass.Salt

	userProvidedPasswordHash := crypto.HashPassword(loginDetails.Password, storedSalt, params)

	if subtle.ConstantTimeCompare(storedHashPass, userProvidedPasswordHash) != 1 {
		return nil, ErrInvalidCredentials
	}

	accessToken, err := a.tokenStore.GenerateToken(user.ID, user.Role, AccessTokenTTL)
	if err != nil {
		return nil, err
	}
	refreshToken, err := a.tokenStore.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	if _, err = a.refreshTokenRepo.CreateRefreshToken(ctx, user.ID, crypto.HashToken(refreshToken), time.Now().Add(RefreshTokenTTL)); err != nil {
		return nil, err
	}
	return &Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (a *AuthService) Refresh(ctx context.Context, incomingRefreshToken string) (*Token, error) {
	if incomingRefreshToken == "" {
		return nil, ErrInvalidToken
	}
	tokenHash := crypto.HashToken(incomingRefreshToken)

	userRefreshToken, err := a.refreshTokenRepo.FindRefreshToken(ctx, tokenHash)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}

	if time.Now().After(userRefreshToken.ExpiresAt) {
		return nil, ErrExpiredToken
	}

	if userRefreshToken.RevokedAt != nil {
		return nil, ErrRefreshTokenRevoked
	}

	refreshToken, err := a.tokenStore.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	if err := a.refreshTokenRepo.CreateAndUpdateToken(ctx,
		userRefreshToken.ID,
		crypto.HashToken(refreshToken),
		userRefreshToken.UserID,
		time.Now().Add(RefreshTokenTTL)); err != nil {
		return nil, err
	}

	user, err := a.userRepo.FindUserByID(ctx, userRefreshToken.UserID)
	if err != nil {
		return nil, err
	}

	accessToken, err := a.tokenStore.GenerateToken(user.ID, user.Role, AccessTokenTTL)
	if err != nil {
		return nil, err
	}

	return &Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
