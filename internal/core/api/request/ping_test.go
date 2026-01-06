package request

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPing(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err, "failed to initialize new http request")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Ping)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	assert.Equal(t, "application/json", rr.Result().Header.Get("Content-Type"))
	assert.Equal(t, "{\"alive\":true}\n", rr.Body.String())
}
