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
	"golang.org/x/sync/errgroup"
)

func RunServer(cfg config.Config, jwtSecret []byte) error {
	var (
		logger  = cfg.Log()
		storage = pg.NewStorage(cfg.DB())

		ctx, cancel = signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	indexer := indexer.New(ctx, cfg)

	server := server.NewServer(
		cfg.Listener(),
		storage,
		indexer,
		logger,
		jwtSecret,
	)

	g.Go(func() error {
		return indexer.Run()
	})

	g.Go(func() error {
		if err := server.RunHTTP(ctx); err != nil {
			return errors.Wrap(err, "http server failed unexpectedly")
		}
		return nil
	})

	return g.Wait()
}
