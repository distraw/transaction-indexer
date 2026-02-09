package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndexer(t *testing.T) {
	fmt.Println("Building docker image...")
	buildImage(t)

	fmt.Println("Setupping compose...")
	setupCompose(t)
	defer run(t, "docker", "compose", "down", "-v")

	fmt.Println("Setupping regtest node...")
	addr := setupRegtestNode(t)

	fmt.Println("Generating blocks...")
	generateToAddress(t, addr, 101)

	require.True(t, healthcheck(t).Alive)

	fmt.Println("Authorizing into indexer...")
	register(t, "user", "123")
	token := login(t, "user", "123")

	postAddress(t, token, addr)
	launch(t)
	waitFor(t, time.Second*11, func() bool {
		return healthcheck(t).CatchedUp
	})

	assert.Equal(t, 5050, balance(t, token, addr))
}
