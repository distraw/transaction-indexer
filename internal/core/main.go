package core

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/distraw/transaction-indexer/internal/config"
	"github.com/distraw/transaction-indexer/internal/core/api/server"
	"github.com/distraw/transaction-indexer/internal/core/indexer"
	"github.com/distraw/transaction-indexer/internal/data/pg"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func RunServer(cfg config.Config, jwtSecret []byte) error {
	var (
		logger  = cfg.Log()
		storage = pg.NewStorage(cfg.DB())
		rpc     = cfg.RPCClient()

		indexerInfo = cfg.IndexerInfo()

		ctx, cancel = signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	)
	defer cancel()

	server := server.NewServer(
		cfg.Listener(),
		storage,
		logger,
		jwtSecret,
	)

	indexer := indexer.New(
		ctx,
		storage,
		logger,
		rpc,
		*indexerInfo,
	)

	go indexer.Run()

	err := server.RunHTTP(ctx)
	if err != nil {
		return errors.Wrap(err, "http server failed unexpectedly")
	}

	return nil
}
