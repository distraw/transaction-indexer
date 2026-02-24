//	End-to-end test.
//	Checks if indexer synchronizes, handles reorgs and returns correct
// 	information regarding tracked addresses
//
//	Because of the nature of golang testing package, logs are not shown
//	during testing. Tests run at average for 120 seconds and during that
//	it may seem that test froze

package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	imageBuilt = false
)

func TestIndexer_Syncing(t *testing.T) {
	if !imageBuilt {
		fmt.Println("Building docker image...")
		buildImage(t)
		imageBuilt = true
	}

	fmt.Println("Setupping compose...")
	setupCompose(t)
	defer run(t, "docker", "compose", "down", "-v")

	fmt.Println("Setupping regtest node...")
	setupRegtestNode(t)

	addr := getNewAddress(t)

	fmt.Println("Generating genesis block...")
	generateToAddress(t, addr, 1)

	require.True(t, healthcheck(t).Alive, "indexer is not alive")

	fmt.Println("Authorizing into indexer...")
	register(t, "user", "123")
	token := login(t, "user", "123")

	postAddress(t, token, addr)
	start(t)
	waitFor(t, time.Second*11, func() bool {
		return healthcheck(t).CatchedUp
	})

	const mineBlocks int = 10
	for i := 0; i < mineBlocks; i++ {
		generateToAddress(t, addr, 1)
		time.Sleep(time.Second)
	}

	assert.Equal(t, 550, balance(t, token, addr))
}

func TestIndexer_CatchUp(t *testing.T) {
	if !imageBuilt {
		fmt.Println("Building docker image...")
		buildImage(t)
		imageBuilt = true
	}

	fmt.Println("Setupping compose...")
	setupCompose(t)
	defer run(t, "docker", "compose", "down", "-v")

	fmt.Println("Setupping regtest node...")
	setupRegtestNode(t)

	addr := getNewAddress(t)

	fmt.Println("Generating blocks...")
	generateToAddress(t, addr, 101)

	require.True(t, healthcheck(t).Alive, "indexer is not alive")

	fmt.Println("Authorizing into indexer...")
	register(t, "user", "123")
	token := login(t, "user", "123")

	postAddress(t, token, addr)
	start(t)
	waitFor(t, time.Second*10, func() bool {
		return healthcheck(t).CatchedUp
	})

	assert.Equal(t, 5050, balance(t, token, addr))
}

func TestIndexer_Reorganizing(t *testing.T) {
	if !imageBuilt {
		fmt.Println("Building docker image...")
		buildImage(t)
		imageBuilt = true
	}

	fmt.Println("Setupping compose...")
	setupCompose(t)
	defer run(t, "docker", "compose", "down", "-v")

	fmt.Println("Setupping regtest node...")
	setupRegtestNode(t)

	addr := getNewAddress(t)

	fmt.Println("Generating blocks...")
	hashes := generateToAddress(t, addr, 10)

	require.True(t, healthcheck(t).Alive, "indexer is not alive")

	fmt.Println("Authorizing into indexer...")
	register(t, "user", "123")
	token := login(t, "user", "123")

	postAddress(t, token, addr)
	start(t)
	waitFor(t, time.Second*10, func() bool {
		return healthcheck(t).CatchedUp
	})

	assert.Equal(t, 500, balance(t, token, addr))

	fmt.Println("Invalidating 5th block...")
	invalidateBlock(t, hashes[len(hashes)-5])

	time.Sleep(time.Second * 15)
	assert.Equal(t, 250, balance(t, token, addr))

	generateToAddress(t, addr, 3)
	time.Sleep(time.Second * 15)
	assert.Equal(t, 400, balance(t, token, addr))
}

func TestIndexer_MultipleUsersAddresses(t *testing.T) {
	if !imageBuilt {
		fmt.Println("Building docker image...")
		buildImage(t)
		imageBuilt = true
	}

	fmt.Println("Setupping compose...")
	setupCompose(t)
	defer run(t, "docker", "compose", "down", "-v")

	fmt.Println("Setupping regtest node...")
	setupRegtestNode(t)

	addr0 := getNewAddress(t)
	addr1 := getNewAddress(t)

	fmt.Println("Generating blocks...")
	generateToAddress(t, addr0, 3)
	generateToAddress(t, addr1, 4)

	require.True(t, healthcheck(t).Alive, "indexer is not alive")

	register(t, "user_0", "password_0")
	register(t, "user_1", "password_1")

	token0 := login(t, "user_0", "password_0")
	token1 := login(t, "user_1", "password_1")

	postAddress(t, token0, addr0)
	postAddress(t, token1, addr0)

	postAddress(t, token0, addr1)
	postAddress(t, token1, addr1)

	start(t)
	waitFor(t, time.Second*10, func() bool {
		return healthcheck(t).CatchedUp
	})

	assert.Equal(t, 150, balance(t, token0, addr0))
	assert.Equal(t, 200, balance(t, token0, addr1))
	assert.Equal(t, 150, balance(t, token1, addr0))
	assert.Equal(t, 200, balance(t, token1, addr1))
}
