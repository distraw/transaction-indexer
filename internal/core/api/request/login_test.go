package request

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/distraw/transaction-indexer/internal/core/api/ctx"
	"github.com/distraw/transaction-indexer/internal/core/api/token"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/distributed_lab/logan/v3"
	"golang.org/x/crypto/bcrypt"
)

func TestLogin(t *testing.T) {
	mockedHashedPassword, err := bcrypt.GenerateFromPassword([]byte("password_0"), bcrypt.DefaultCost)
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
				Username: "user_0",
				Password: mockedHashedPassword,
			}),
			inputUsername:            "user_0",
			inputPassword:            "password_0",
			wantStatus:               http.StatusOK,
			wantContentHeader:        "application/json",
			wantResponseTokenIssuer:  "transaction-indexer",
			wantResponseTokenSubject: "user_0",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			//HTTP method is enforced by chi router -> no difference which one is tested here
			req, err := http.NewRequest(http.MethodPost, "/login", nil)
			require.NoError(t, err, "failed to initialize new http request")

			req.SetBasicAuth(tt.inputUsername, tt.inputPassword)

			mockedSecret := "1234567890abcdefghijklmnopqrstuv"

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

			var parsedBody struct {
				Token string `json:"token"`
			}

			err = json.Unmarshal(body, &parsedBody)
			require.NoError(t, err, "failed to fetch jwt from response body")

			tokenClaims, err := token.Parse(parsedBody.Token, []byte(mockedSecret))
			require.NoError(t, err, "failed unexpectedly to parse JWT")

			assert.Equal(t, "transaction-indexer", tokenClaims.Issuer)
			assert.Equal(t, "user_0", tokenClaims.Subject)

			now := time.Now()
			assert.False(t, tokenClaims.IssuedAt.Equal(tokenClaims.ExpiresAt.Time))
			assert.True(t, tokenClaims.ExpiresAt.After(now) || tokenClaims.ExpiresAt.Equal(now))
			assert.True(t, tokenClaims.IssuedAt.Before(now) || tokenClaims.IssuedAt.Equal(now))
		})
	}
}
