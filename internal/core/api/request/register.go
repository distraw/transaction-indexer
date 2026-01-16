package request

import (
	"net/http"

	"github.com/distraw/transaction-indexer/internal/core/api/ctx"
	"github.com/distraw/transaction-indexer/internal/data"
	"golang.org/x/crypto/bcrypt"
)

func Register(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	log := ctx.Logger(c)

	username, password, ok := r.BasicAuth()
	if !ok || len(username) == 0 || len(password) == 0 {
		http.Error(w, "401 unauthorized (invalid authorization header)", http.StatusUnauthorized)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "400 bad request (password must be less than 72 symbols)", http.StatusBadRequest)
	}

	db := ctx.DB(c)
	_, err = db.Insert(data.User{
		Username: username,
		Password: hashedPassword,
	})

	if err == data.ErrAlreadyExists {
		http.Error(w, "409 conflict (username was already taken)", http.StatusConflict)
		return
	}

	if err != nil {
		log.WithError(err).Error()
		http.Error(w, "500 internal server error", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}
