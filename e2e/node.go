package e2e

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	user = "rpcuser_0"
	pass = "rpcpassword_0"

	indexerEndpoint = "http://localhost:8080/"
)

var (
	standardArgs = []string{"-regtest",
		"-rpcpassword=" + pass,
		"-rpcuser=" + user,
	}
)

func setupRegtestNode(t *testing.T) {
	t.Helper()
	run(t, "bitcoin-cli", append(standardArgs, "createwallet", "wallet_0")...)
}

func getNewAddress(t *testing.T) (addr string) {
	t.Helper()
	return run(t, "bitcoin-cli", append(standardArgs, "getnewaddress")...)
}

func invalidateBlock(t *testing.T, blockhash string) {
	t.Helper()
	run(t, "bitcoin-cli", append(standardArgs, "invalidateblock", blockhash)...)
}

func generateToAddress(t *testing.T, addr string, blocks int) []string {
	t.Helper()
	output := run(t, "bitcoin-cli", append(standardArgs, "generatetoaddress", strconv.Itoa(blocks), addr)...)

	var blocksHashes []string

	err := json.Unmarshal([]byte(output), &blocksHashes)
	require.NoError(t, err, "failed to unmarshal output from generatetoaddress")
	return blocksHashes
}
