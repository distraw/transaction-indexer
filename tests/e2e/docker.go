package e2e

import "testing"

func buildImage(t *testing.T) {
	t.Helper()

	run(t, "docker", "build", "-t", "transaction-indexer", ".")
}

func setupCompose(t *testing.T) {
	t.Helper()

	run(t, "docker", "compose", "down", "-v")
	run(t, "docker", "compose", "up", "-d")
}
