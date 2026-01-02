package core

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/distraw/transaction-indexer/internal/config"
	"github.com/distraw/transaction-indexer/internal/core/api/server"
	"github.com/distraw/transaction-indexer/internal/data/pg"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func RunServer(cfg config.Config) error {
	var (
		logger      = cfg.Log()
		db          = pg.NewUsersQ(cfg.DB())
		ctx, cancel = signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	)
	defer cancel()

	server := server.NewServer(
		cfg.Listener(),
		db,
		logger,
	)

	err := server.RunHTTP(ctx)
	if err != nil {
		return errors.Wrap(err, "http server failed unexpectedly")
	}

	return nil
}
