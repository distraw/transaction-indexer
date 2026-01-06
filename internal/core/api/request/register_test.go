package request

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/distraw/transaction-indexer/internal/core/api/ctx"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err, "failed to initialize new http request")

	req.SetBasicAuth(data.MockedUser, data.MockedPassword)
	req = req.WithContext(ctx.DBProvider(&data.UsersQMock{})(req.Context()))

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Register)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	assert.Equal(t, "application/json", rr.Result().Header.Get("Content-Type"))
	assert.Equal(t, "{}\n", rr.Body.String())
}
