package request

import (
	"encoding/json"
	"net/http"

	"github.com/distraw/transaction-indexer/internal/core/api/ctx"
	"github.com/go-chi/chi/v5"
	"github.com/josemiguelmelo/btcaddressvalidator"
)

func GetBalance(w http.ResponseWriter, r *http.Request) {
	addr := chi.URLParam(r, "address")

	_, err := btcaddressvalidator.CheckBtcAddress(addr)
	if err != nil {
		http.Error(w, "invalid btc address format", http.StatusUnprocessableEntity)
		return
	}

	tracks, err := ctx.Storage(r.Context()).IsTracking(*ctx.UserID(r.Context()), addr)
	switch {
	case err != nil:
		ctx.Logger(r.Context()).WithError(err).Error("failed to access db to check if user tracks the address")
		http.Error(w, "", http.StatusInternalServerError)
		return
	case !tracks:
		http.Error(w, "address is not tracked", http.StatusUnauthorized)
		return
	}

	balance, err := ctx.Storage(r.Context()).GetBalance(addr)
	if err != nil {
		ctx.Logger(r.Context()).WithError(err).Error("failed to access db to check if user tracks the address")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(&balance)
}
