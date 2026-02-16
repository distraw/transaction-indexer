package request

import (
	"encoding/json"
	"net/http"

	"github.com/distraw/transaction-indexer/internal/core/api/ctx"
	"github.com/distraw/transaction-indexer/internal/core/api/token"
	"github.com/distraw/transaction-indexer/internal/data"
	"golang.org/x/crypto/bcrypt"
)

func Login(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, "basic auth header is missing or invalid", http.StatusUnauthorized)
		return
	}

	user, err := ctx.Storage(r.Context()).Users().Get(username)
	if err == data.ErrNotFound {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	if err != nil {
		ctx.Logger(r.Context()).WithError(err).Error("db failed unexpectedly")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword(user.Password, []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	if err != nil {
		ctx.Logger(r.Context()).WithError(err).Error("password comparison failed unexpectedly")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	signedJWT, err := token.Sign(ctx.JWTSecret(r.Context()), username)
	if err != nil {
		ctx.Logger(r.Context()).WithError(err).Error("failed unexpectedly to issue signed jwt")
		http.Error(w, "", http.StatusInternalServerError)
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"token": signedJWT,
	})
}
