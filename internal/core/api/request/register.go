package request

import (
	"net/http"

	"github.com/distraw/transaction-indexer/internal/core/api/ctx"
	"github.com/distraw/transaction-indexer/internal/data"
	"golang.org/x/crypto/bcrypt"
)

func Register(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok || len(username) == 0 || len(password) == 0 {
		http.Error(w, "invalid authorization header", http.StatusUnauthorized)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "password must be less than 72 symbols", http.StatusBadRequest)
		return
	}

	_, err = ctx.Storage(r.Context()).Users().Insert(data.User{
		Username: username,
		Password: hashedPassword,
	})

	if err == data.ErrAlreadyExists {
		http.Error(w, "username was already taken", http.StatusConflict)
		return
	}
	if err != nil {
		ctx.Logger(r.Context()).WithError(err).Error("failed to insert new user into storage")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
