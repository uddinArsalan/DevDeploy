package token

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/uddinArsalan/devdeploy/internals/domain"
)

type JwtToken struct {
	secret        string
	signingMethod jwt.SigningMethod
}

func NewJwtToken() *JwtToken {
	secret := os.Getenv("JWT_SECRET")
	signingMethod := jwt.SigningMethodHS256
	return &JwtToken{
		secret:        secret,
		signingMethod: signingMethod,
	}
}

type jwtClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func (jw *JwtToken) GenerateToken(userID int64, role string) (string, error) {
	token := jwt.NewWithClaims(jw.signingMethod, jwtClaims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})
	return token.SignedString([]byte(jw.secret))
}

func (jw *JwtToken) ParseToken(tokenString string) (*domain.UserClaims, error) {
	claims := &jwtClaims{}
	token, err := jwt.ParseWithClaims(tokenString,
		claims,
		func(t *jwt.Token) (any, error) {
			return []byte(jw.secret), nil
		},
		jwt.WithValidMethods([]string{jw.signingMethod.Alg()}))

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	sub, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return nil, err
	}

	return &domain.UserClaims{
		UserID: sub,
		Role:   claims.Role,
	}, nil
}
