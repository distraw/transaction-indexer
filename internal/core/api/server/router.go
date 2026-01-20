package server

import (
	"net/http"

	"github.com/distraw/transaction-indexer/internal/core/api/request"
	"github.com/go-chi/chi/v5"
	"gitlab.com/distributed_lab/ape"
)

func (s *server) httpRouter() http.Handler {
	router := chi.NewRouter()

	router.Use(
		ape.LoganMiddleware(s.log),
		ape.RecoverMiddleware(s.log),
		ape.CtxMiddleWare(s.ctxExtenders...),
	)

	router.Get("/healthcheck", request.Healthcheck)
	router.Post("/register", request.Register)
	router.Post("/login", request.Login)

	router.With(AuthMiddleware()).Post("/addresses", request.Addresses)

	return router
}
