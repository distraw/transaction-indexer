package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

const (
	ExpirationPeriod = time.Hour * 1
	Issuer           = "transaction-indexer"
)

var (
	signingMethod = jwt.SigningMethodHS256
)

// Parse unmarshals token string into json web token and extracts claims
//
// Returns ErrTokenParsingFailed if token is invalid
func Parse(tokenString string, secret []byte) (*jwt.RegisteredClaims, error) {
	claims := &jwt.RegisteredClaims{}

	_, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		},
		jwt.WithValidMethods([]string{signingMethod.Name}),
		jwt.WithIssuer(Issuer),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithLeeway(30*time.Second),
	)

	return claims, err
}

func Sign(secret []byte, subject string) (string, error) {
	now := time.Now()

	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(ExpirationPeriod)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		Issuer:    Issuer,
		Subject:   subject,
	}

	token := jwt.NewWithClaims(signingMethod, claims)
	signedJWT, err := token.SignedString(secret)
	if err != nil {
		return "", errors.Wrap(err, "failed unexpectedly to sign json web token")
	}

	return signedJWT, nil
}
