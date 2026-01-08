package request

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/distraw/transaction-indexer/internal/core/api/ctx"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/distributed_lab/logan/v3"
)

// Only GET method is tested as it would be enforced by chi router anyway
func TestRegister(t *testing.T) {
	const (
		noContentType   = ""
		contentTypeText = "text/plain; charset=utf-8"

		mockedUsername = "user_0"
		mockedPassword = "password_0"
	)

	tests := map[string]struct {
		mockDB            data.UsersQMock
		inputUsername     string
		inputPassword     string
		wantStatus        int
		wantContentHeader string
		wantResponseBody  string
	}{
		"should return 200 (OK) if credentials are correct": {
			mockDB:            data.NewUsersQMock(),
			inputUsername:     mockedUsername,
			inputPassword:     mockedPassword,
			wantStatus:        http.StatusOK,
			wantContentHeader: noContentType,
			wantResponseBody:  "",
		},
		"should return 401 (Unauthorized) if credentials are missing": {
			mockDB:            data.NewUsersQMock(),
			inputUsername:     "",
			inputPassword:     "",
			wantStatus:        http.StatusUnauthorized,
			wantContentHeader: contentTypeText,
			wantResponseBody:  "Use http authorization header to provide credentials\n",
		},
		"should return 400 (Bad request) if provided password is too long": {
			mockDB:            data.NewUsersQMock(),
			inputUsername:     mockedUsername,
			inputPassword:     strings.Repeat("x", 73),
			wantStatus:        http.StatusBadRequest,
			wantContentHeader: contentTypeText,
			wantResponseBody:  "Provided password is too long (72 characters max)\n",
		},
		"should return 409 (Conflict) if username was already taken": {
			mockDB: data.UsersQMock{Data: map[string]data.User{
				mockedUsername: {
					Username: mockedUsername,
					Password: []byte(mockedPassword),
				},
			},
			},
			inputUsername:     mockedUsername,
			inputPassword:     mockedPassword,
			wantStatus:        http.StatusConflict,
			wantContentHeader: contentTypeText,
			wantResponseBody:  "Username was already taken\n",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			//HTTP method is enforced by chi router -> no difference which one is tested here
			req, err := http.NewRequest(http.MethodGet, "/register", nil)
			require.NoError(t, err, "failed to initialize new http request")

			req.SetBasicAuth(tt.inputUsername, tt.inputPassword)

			// Put mocked DB and logger into request context
			req = req.WithContext(ctx.DBProvider(&tt.mockDB)(req.Context()))
			req = req.WithContext(ctx.LoggerProvider(logan.New())(req.Context()))

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(Register)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			assert.Equal(t, tt.wantContentHeader, rr.Result().Header.Get("Content-Type"))
			assert.Equal(t, tt.wantResponseBody, rr.Body.String())
		})
	}
}
