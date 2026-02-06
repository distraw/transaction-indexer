package request

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/distraw/transaction-indexer/internal/core/api/ctx"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/go-chi/chi/v5"
	"github.com/josemiguelmelo/btcaddressvalidator"
)

func Balance(w http.ResponseWriter, r *http.Request) {
	addr := chi.URLParam(r, "address")

	_, err := btcaddressvalidator.CheckBtcAddress(addr)
	if err != nil {
		http.Error(w, "422 unprocessable entity (invalid btc address format)", http.StatusUnprocessableEntity)
		return
	}

	c := r.Context()
	storage := ctx.Storage(c)
	tracks, err := storage.IsTracking(*ctx.UserID(c), addr)
	switch {
	case errors.Is(err, data.ErrNotFound):
		// if address does not exist in db, user obviously doesnt track it.
		http.Error(w, "401 unauthorized", http.StatusUnauthorized)
		return
	case err != nil:
		ctx.Logger(c).WithError(err).Error("failed to access db to check if user tracks the address")
		http.Error(w, "500 internal server error", http.StatusInternalServerError)
		return
	case !tracks:
		http.Error(w, "401 unauthorized", http.StatusUnauthorized)
		return
	}

	balance, err := storage.GetBalance(addr)
	if err != nil {
		ctx.Logger(c).WithError(err).Error("failed to access db to check if user tracks the address")
		http.Error(w, "500 internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(balance)
}
