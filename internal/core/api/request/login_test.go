package request

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/distraw/transaction-indexer/internal/core/api/ctx"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/distributed_lab/logan/v3"
	"golang.org/x/crypto/bcrypt"
)

func parseJWT(t *testing.T, body []byte, secret []byte) *jwt.RegisteredClaims {
	var parsedBody struct {
		Token string `json:"token"`
	}

	err := json.Unmarshal(body, &parsedBody)
	require.NoError(t, err)

	token, err := jwt.ParseWithClaims(parsedBody.Token, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	require.NoError(t, err)

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	require.True(t, ok, "failed to fetch claims from token")

	return claims
}

func TestLogin(t *testing.T) {
	const (
		noContentType   = ""
		contentTypeText = "text/plain; charset=utf-8"
		contentTypeJSON = "application/json"

		mockedUsername = "user_0"
		mockedPassword = "password_0"

		mockedSecret = "1234567890abcdefghijklmnopqrstuv"

		mockedIssuer  = "transaction-indexer"
		mockedSubject = mockedUsername
	)

	mockedHashedPassword, err := bcrypt.GenerateFromPassword([]byte(mockedPassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	tests := map[string]struct {
		mockDB                   data.UsersQMock
		inputUsername            string
		inputPassword            string
		wantStatus               int
		wantContentHeader        string
		wantResponseTokenIssuer  string
		wantResponseTokenSubject string
	}{
		"should return 200 (OK) and JWT when user with given credentials exists": {
			mockDB: data.NewUsersQMock().WithUser(data.User{
				Username: mockedUsername,
				Password: mockedHashedPassword,
			}),
			inputUsername:            mockedUsername,
			inputPassword:            mockedPassword,
			wantStatus:               http.StatusOK,
			wantContentHeader:        contentTypeJSON,
			wantResponseTokenIssuer:  mockedIssuer,
			wantResponseTokenSubject: mockedUsername,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			//HTTP method is enforced by chi router -> no difference which one is tested here
			req, err := http.NewRequest(http.MethodPost, "/login", nil)
			require.NoError(t, err, "failed to initialize new http request")

			req.SetBasicAuth(tt.inputUsername, tt.inputPassword)

			// Put mocked DB and logger into request context
			req = req.WithContext(ctx.DBProvider(&tt.mockDB)(req.Context()))
			req = req.WithContext(ctx.LoggerProvider(logan.New())(req.Context()))
			req = req.WithContext(ctx.JWTSecretProvider([]byte(mockedSecret))(req.Context()))

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(Login)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			assert.Equal(t, tt.wantContentHeader, rr.Result().Header.Get("Content-Type"))

			body, err := io.ReadAll(rr.Result().Body)
			require.NoError(t, err)

			tokenClaims := parseJWT(t, body, []byte(mockedSecret))

			assert.Equal(t, mockedIssuer, tokenClaims.Issuer)
			assert.Equal(t, mockedSubject, tokenClaims.Subject)

			now := time.Now()
			assert.False(t, tokenClaims.IssuedAt.Equal(tokenClaims.ExpiresAt.Time))
			assert.True(t, tokenClaims.ExpiresAt.After(now) || tokenClaims.ExpiresAt.Equal(now))
			assert.True(t, tokenClaims.IssuedAt.Before(now) || tokenClaims.IssuedAt.Equal(now))
		})
	}
}
