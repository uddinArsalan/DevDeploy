package token

import (
	"time"

	"github.com/uddinArsalan/devdeploy/internals/domain"
)

type TokenStore interface {
	GenerateToken(userID int64, role domain.UserRoles, expiry time.Duration) (string, error)
	GenerateRefreshToken() (string, error)
	ParseToken(tokenString string) (*domain.UserClaims, error)
}
