package request

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthcheck(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err, "failed to initialize new http request")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Healthcheck)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Result().Header.Get("Content-Type"))
	assert.Equal(t, "{\"alive\":true}\n", rr.Body.String())
}
