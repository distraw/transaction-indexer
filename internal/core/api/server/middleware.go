package server

import (
	"net/http"
	"strings"

	"github.com/distraw/transaction-indexer/internal/core/api/ctx"
	"github.com/distraw/transaction-indexer/internal/core/api/token"
	"github.com/distraw/transaction-indexer/internal/data"
)

func AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				http.Error(w,
					"400 bad request (invalid authorization header format)",
					http.StatusBadRequest,
				)
				return
			}
			auth = strings.TrimPrefix(auth, "Bearer ")

			c := r.Context()

			jwtClaims, err := token.Parse(auth, ctx.JWTSecret(c))
			if err != nil {
				http.Error(w, "401 unauthorized", http.StatusUnauthorized)
				return
			}

			if len(jwtClaims.Subject) == 0 ||
				jwtClaims.Issuer != token.Issuer {
				http.Error(w, "401 unauthorized", http.StatusUnauthorized)
				return
			}

			user, err := ctx.DB(c).Get(jwtClaims.Subject)
			if err == data.ErrNotFound {
				http.Error(w, "401 unauthorized", http.StatusUnauthorized)
				return
			}
			if err != nil {
				ctx.Logger(c).WithError(err).Error("failed to check user existence in db")
				http.Error(w, "500 internal server error", http.StatusInternalServerError)
				return
			}

			c = ctx.UserIDProvider(&user.ID)(c)
			next.ServeHTTP(w, r.WithContext(c))
		})
	}
}
