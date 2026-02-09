package e2e

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func run(t *testing.T, name string, args ...string) string {
	t.Helper()

	cmd := exec.Command(name, args...)
	cmd.Dir = "../"
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	if len(out) != 0 {
		// output from CombinedOutput always contains /n in the end
		out = out[:len(out)-1]
	}
	return string(out)
}

func waitFor(t *testing.T, timeout time.Duration, fn func() bool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatal("timeout waiting for condition")
		case <-ticker.C:
			if fn() {
				return
			}
		}
	}
}
