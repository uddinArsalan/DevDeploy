package services

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/uddinArsalan/devdeploy/internals/adapters/token"
	"github.com/uddinArsalan/devdeploy/internals/domain"
	"github.com/uddinArsalan/devdeploy/internals/repository"

	"crypto/rand"
	"crypto/subtle"

	"golang.org/x/crypto/argon2"
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
		userRepo,
		refreshTokenRepo,
		tokenStore,
	}
}

func (a *AuthService) CreateUser(ctx context.Context, userDetails domain.CreateUser) error {
	hashedPassword, err := hashAndEncodePassword(userDetails.Password)
	if err != nil {
		return err
	}
	return a.userRepo.CreateUser(ctx, userDetails.Name, userDetails.Email, hashedPassword)
}

var ErrInvalidCredentials = errors.New("invalid credentials")

func (a *AuthService) LoginUser(ctx context.Context, loginDetails domain.LoginUser) (*Token, error) {
	user, err := a.userRepo.FindUserByEmail(ctx, loginDetails.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	//password - check
	parsedPass, err := parsePassword(user.PasswordHash)
	if err != nil {
		return nil, err
	}
	params := Params{
		Memory:      parsedPass.Memory,
		Iterations:  parsedPass.Iterations,
		Parallelism: parsedPass.Parallelism,
		KeyLen:      parsedPass.KeyLen,
	}
	storedHashPass := parsedPass.PasswordHash
	storedSalt := parsedPass.Salt

	userProvidedPasswordHash := hashPassword(loginDetails.Password, storedSalt, params)

	if subtle.ConstantTimeCompare(storedHashPass, userProvidedPasswordHash) != 1 {
		return nil, ErrInvalidCredentials
	}

	accessToken, err := a.tokenStore.GenerateToken(user.ID, user.Role, 15*time.Minute)
	if err != nil {
		return nil, err
	}
	tokenHash, err := a.tokenStore.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	if err = a.refreshTokenRepo.CreateRefreshToken(ctx, user.ID, tokenHash, time.Now().Add(7*24*time.Hour)); err != nil {
		return nil, err
	}
	return &Token{
		AccessToken:  accessToken,
		RefreshToken: tokenHash,
	}, nil
}

func generateSalt(size uint32) ([]byte, error) {
	salt := make([]byte, size)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

type Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	KeyLen      uint32
}

func hashPassword(password string, salt []byte, params Params) []byte {
	key := argon2.IDKey([]byte(password), salt, params.Iterations, params.Memory, params.Parallelism, params.KeyLen)
	return key
}

func hashAndEncodePassword(password string) (string, error) {
	salt, err := generateSalt(16)
	if err != nil {
		return "", err
	}
	params := Params{
		Memory:      20 * 1024,
		Iterations:  2,
		Parallelism: 1,
		KeyLen:      32,
	}
	key := hashPassword(password, salt, params)

	b64Key := base64.RawStdEncoding.EncodeToString(key)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)

	hash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, params.Memory, params.Iterations, params.Parallelism, b64Salt, b64Key)
	return hash, nil
}

type ParsedPassword struct {
	PasswordHash []byte
	Salt         []byte
	Params
}

var ErrInvalidHash = errors.New("invalid hash")

func parsePassword(password string) (ParsedPassword, error) {
	parts := strings.Split(password, "$")

	if parts[1] != "argon2id" {
		return ParsedPassword{}, ErrInvalidHash
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return ParsedPassword{}, ErrInvalidHash
	}

	if version != argon2.Version {
		return ParsedPassword{}, ErrInvalidHash
	}

	if len(parts) != 6 {
		return ParsedPassword{}, ErrInvalidHash
	}
	config := strings.Split(parts[3], ",")
	if len(config) != 3 {
		return ParsedPassword{}, ErrInvalidHash
	}
	var (
		memory      uint32
		iterations  uint32
		parallelism uint8
	)
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return ParsedPassword{}, ErrInvalidHash
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return ParsedPassword{}, ErrInvalidHash
	}
	passwordHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return ParsedPassword{}, ErrInvalidHash
	}
	return ParsedPassword{
		PasswordHash: passwordHash,
		Salt:         salt,
		Params: Params{
			Memory:      memory,
			Iterations:  iterations,
			Parallelism: parallelism,
			KeyLen:      uint32(len(passwordHash)),
		},
	}, ErrInvalidHash
}

// ""
// argon2id
// v=""
// m=,t=,p=
// ""
// ""
