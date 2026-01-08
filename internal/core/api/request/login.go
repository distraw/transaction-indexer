package request

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/distraw/transaction-indexer/internal/core/api/ctx"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

func issueSignedJWT(secret []byte, subject string) (string, error) {
	const (
		expirationPeriod = time.Hour * 1
		issuer           = "transaction-indexer"
	)

	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expirationPeriod)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    issuer,
		Subject:   subject,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedJWT, err := token.SignedString(secret)
	if err != nil {
		return "", errors.Wrap(err, "failed unexpectedly to sign json web token")
	}

	return signedJWT, nil
}

func Login(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	log := ctx.Logger(c)

	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, "Use http authorization header to provide credentials", http.StatusUnauthorized)
		return
	}

	db := ctx.DB(c)
	user, err := db.Get(username)
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
			WithField("user.Password", string(user.Password)).
			WithField("db.Password", password).
			Error("password comparison failed unexpectedly")
		http.Error(w, "500 internal server error", http.StatusInternalServerError)
		return
	}

	signedJWT, err := issueSignedJWT(ctx.JWTSecret(c), username)
	if err != nil {
		log.WithError(err).Error("failed unexpectedly to issue signed jwt")
		http.Error(w, "Failed unexpectedly to issue jwt", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"token": signedJWT,
	})
}
