package request

import (
	"encoding/json"
	"net/http"

	"github.com/distraw/transaction-indexer/internal/core/api/ctx"
	"github.com/distraw/transaction-indexer/internal/data"
	"golang.org/x/crypto/bcrypt"
)

func Register(w http.ResponseWriter, r *http.Request) {
	c := r.Context()

	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, "Use http authorization header to provide credentials.", http.StatusUnauthorized)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to generate hashcode from given password.", http.StatusInternalServerError)
	}

	db := ctx.DB(c)
	_, err = db.Insert(data.User{
		Username: username,
		Password: hashedPassword,
	})

	if err == data.ErrAlreadyExists {
		http.Error(w, "Username was already taken.", http.StatusConflict)
		return
	}

	if err != nil {
		http.Error(w, "Failed to save credentials on server.", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Write empty body upon successful registration
	json.NewEncoder(w).Encode(map[string]any{})
}
