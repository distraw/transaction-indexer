package server

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/distraw/transaction-indexer/internal/core/api/ctx"
	"github.com/distraw/transaction-indexer/internal/core/indexer"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/logan/v3"
)

type Server interface {
	RunHTTP(ctx context.Context) error
}

type server struct {
	http net.Listener

	log          *logan.Entry
	ctxExtenders []func(context.Context) context.Context
}

func (s *server) RunHTTP(ctx context.Context) error {
	server := &http.Server{
		Handler: s.httpRouter(),
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()

		deadline, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := server.Shutdown(deadline); err != nil {
			s.log.WithError(err).Error("failed to gracefully shutdown the http server")
		}
		s.log.Info("http serving stopped: context cancelled")
	}()

	s.log.Info("http serving started")
	err := server.Serve(s.http)
	if !errors.Is(err, http.ErrServerClosed) {
		return errors.Wrap(err, "unexpected server error occured")
	}

	wg.Wait()
	return nil
}

func NewServer(
	http net.Listener,
	storage data.Storage,
	indexer indexer.Indexer,
	log *logan.Entry,
	jwtSecret []byte,
) Server {
	return &server{
		http: http,
		log:  log,
		ctxExtenders: []func(context.Context) context.Context{
			ctx.LoggerProvider(log),
			ctx.StorageProvider(storage),
			ctx.JWTSecretProvider(jwtSecret),
			ctx.IndexerProvider(indexer),
		},
	}
}
