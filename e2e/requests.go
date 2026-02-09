package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

type healthcheckResponse struct {
	Alive     bool `json:"alive"`
	Launched  bool `json:"launched"`
	CatchedUp bool `json:"catchedUp"`
}

func healthcheck(t *testing.T) healthcheckResponse {
	resp, err := http.Get(indexerEndpoint + "healthcheck")
	if err != nil {
		return healthcheckResponse{
			Alive:     false,
			Launched:  false,
			CatchedUp: false,
		}
	}
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var healthcheck healthcheckResponse

	json.Unmarshal(body, &healthcheck)
	return healthcheck
}

func postAddress(t *testing.T, token string, addr string) {
	reqBody, err := json.Marshal(map[string]interface{}{
		"addr": addr,
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, indexerEndpoint+"addresses", bytes.NewReader(reqBody))
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
}

func register(t *testing.T, username string, password string) {
	req, err := http.NewRequest(http.MethodPost, indexerEndpoint+"register", nil)
	require.NoError(t, err)
	req.SetBasicAuth(username, password)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func login(t *testing.T, username string, password string) (jwtToken string) {
	req, err := http.NewRequest(http.MethodPost, indexerEndpoint+"login", nil)
	require.NoError(t, err)
	req.SetBasicAuth(username, password)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	defer resp.Body.Close()

	var token struct {
		Value string `json:"token"`
	}
	err = json.Unmarshal(respBody, &token)

	return token.Value
}

func launch(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, indexerEndpoint+"launch", nil)
	require.NoError(t, err, "failed to launch indexer")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func balance(t *testing.T, token string, addr string) int {
	req, err := http.NewRequest(http.MethodGet, indexerEndpoint+"addresses/"+addr+"/balance", nil)
	require.NoError(t, err)
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	respBody = respBody[:len(respBody)-1]

	balance, err := strconv.Atoi(string(respBody))
	require.NoError(t, err)

	return balance
}
