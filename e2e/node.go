package e2e

import (
	"strconv"
	"testing"
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

func setupRegtestNode(t *testing.T) (addr string) {
	t.Helper()
	run(t, "bitcoin-cli", append(standardArgs, "createwallet", "wallet_0")...)
	return run(t, "bitcoin-cli", append(standardArgs, "getnewaddress")...)
}

func generateToAddress(t *testing.T, addr string, blocks int) {
	t.Helper()
	run(t, "bitcoin-cli", append(standardArgs, "generatetoaddress", strconv.Itoa(blocks), addr)...)
}
