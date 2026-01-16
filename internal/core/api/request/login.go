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
		http.Error(w, "Use http authorization header to provide credentials", http.StatusUnauthorized)
		return
	}

	c := r.Context()
	log := ctx.Logger(c)
	user, err := ctx.DB(c).Get(username)
	if err == data.ErrUserNotFound {
		http.Error(w, "User with given credentials does not exist", http.StatusUnauthorized)
		return
	}
	if err != nil {
		log.WithError(err).Error("db failed unexpectedly")
		http.Error(w, "Failed unexpectedly to search for user with given credentials", http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword(user.Password, []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		http.Error(w, "User with given credentials does not exist", http.StatusUnauthorized)
		return
	}
	if err != nil {
		log.WithError(err).
			Error("password comparison failed unexpectedly")
		http.Error(w, "500 internal server error", http.StatusInternalServerError)
		return
	}

	signedJWT, err := token.Sign(ctx.JWTSecret(c), username)
	if err != nil {
		log.WithError(err).Error("failed unexpectedly to issue signed jwt")
		http.Error(w, "Failed unexpectedly to issue jwt", http.StatusInternalServerError)
	}

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"token": signedJWT,
	})
}
