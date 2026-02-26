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
	router.Post("/start", request.Start)

	router.Route("/addresses", func(r chi.Router) {
		r.Use(AuthMiddleware())

		r.Post("/", request.AddAddressToTrack)
		r.Get("/", request.GetAddresses)
		r.Get("/{address}/balance", request.GetBalance)
		r.Get("/{address}/utxos", request.GetUtxos)
		r.Get("/{address}/txs", request.GetTXs)
	})

	return router
}
