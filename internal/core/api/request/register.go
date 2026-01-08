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
		http.Error(w, "Use http authorization header to provide credentials", http.StatusUnauthorized)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Provided password is too long (72 characters max)", http.StatusBadRequest)
	}

	db := ctx.DB(c)
	err = db.Insert(data.User{
		Username: username,
		Password: hashedPassword,
	})

	if err == data.ErrAlreadyExists {
		http.Error(w, "Username was already taken", http.StatusConflict)
		return
	}

	if err != nil {
		log.WithError(err).Error()
		http.Error(w, "Failed to save credentials on server", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}
