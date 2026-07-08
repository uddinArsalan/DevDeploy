package middlewares

import (
	"context"
	"net/http"

	"github.com/uddinArsalan/devdeploy/internals/adapters/token"
	"github.com/uddinArsalan/devdeploy/internals/domain"
	"github.com/uddinArsalan/devdeploy/internals/utils"
)

type Auth struct {
	token token.TokenStore
}

func NewAuthMiddleware(token token.TokenStore) *Auth {
	return &Auth{
		token: token,
	}
}

type authUserKey struct{}

func (a *Auth) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken, err := r.Cookie("access_token")
		if err != nil || accessToken.Value == "" {
			utils.FAIL(w, http.StatusUnauthorized, "authentication required")
			return
		}
		claims, err := a.token.ParseToken(accessToken.Value)
		if err != nil {
			utils.FAIL(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		ctx := context.WithValue(r.Context(), authUserKey{}, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserFromContext(ctx context.Context) (*domain.UserClaims, bool) {
	claim, ok := ctx.Value(authUserKey{}).(*domain.UserClaims)
	return claim, ok
}
