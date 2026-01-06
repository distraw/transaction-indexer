package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/distraw/transaction-indexer/internal/core/api/ctx"
	"github.com/distraw/transaction-indexer/internal/core/api/request"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/go-chi/chi/v5"
	"gitlab.com/distributed_lab/ape"
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

	go func() {
		<-ctx.Done()
		deadline, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		if err := server.Shutdown(deadline); err != nil {
			s.log.WithError(err).Error("failed to gracefully shutdown the http server")
		}
		s.log.Info("http serving stopped: context cancelled")
	}()

	s.log.Info("http serving started")
	err := server.Serve(s.http)
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *server) httpRouter() http.Handler {
	router := chi.NewRouter()

	router.Use(
		ape.LoganMiddleware(s.log),
		ape.RecoverMiddleware(s.log),
		ape.CtxMiddleWare(s.ctxExtenders...),
	)

	router.Get("/ping", request.Ping)
	router.Post("/register", request.Register)
	router.Post("/login", request.Login)

	return router
}

func NewServer(
	http net.Listener,
	db data.UsersQ,
	log *logan.Entry,
	jwtSecret []byte,
) Server {
	return &server{
		http: http,
		log:  log,
		ctxExtenders: []func(context.Context) context.Context{
			ctx.LoggerProvider(log),
			ctx.DBProvider(db),
			ctx.JWTSecretProvider(jwtSecret),
		},
	}
}
