package cli

import "github.com/distraw/transaction-indexer/internal/config"

func RunService(cfg config.Config) error {
	cfg.Log().Info("Hello, World!")
	return nil
}
