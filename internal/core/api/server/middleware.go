package server

import (
	"net/http"
	"strings"

	"github.com/distraw/transaction-indexer/internal/core/api/ctx"
	"github.com/distraw/transaction-indexer/internal/core/api/token"
)

const (
	authSchemeBearer = "Bearer "

	unauthorizedErrBody = "401 unauthorized"
)

func AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, authSchemeBearer) {
				http.Error(w,
					"400 bad request (invalid authorization header format)",
					http.StatusBadRequest,
				)
				return
			}
			auth = strings.TrimPrefix(auth, authSchemeBearer)

			c := r.Context()

			jwtClaims, err := token.Parse(auth, ctx.JWTSecret(c))
			if err != nil {
				http.Error(w, unauthorizedErrBody, http.StatusUnauthorized)
				return
			}

			if len(jwtClaims.Subject) == 0 ||
				jwtClaims.Issuer != token.Issuer ||
				!ctx.DB(c).Exists(jwtClaims.Subject) {
				http.Error(w, unauthorizedErrBody, http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
