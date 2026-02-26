package request

import (
	"encoding/json"
	"net/http"

	"github.com/distraw/transaction-indexer/internal/core/api/ctx"
	"github.com/go-chi/chi/v5"
	"github.com/josemiguelmelo/btcaddressvalidator"
)

func GetUtxos(w http.ResponseWriter, r *http.Request) {
	address := chi.URLParam(r, "address")

	_, err := btcaddressvalidator.CheckBtcAddress(address)
	if err != nil {
		http.Error(w, "invalid btc address format", http.StatusUnprocessableEntity)
		return
	}

	tracks, err := ctx.Storage(r.Context()).IsTracking(*ctx.UserID(r.Context()), address)
	switch {
	case err != nil:
		ctx.Logger(r.Context()).WithError(err).Error("failed to access db to check if user tracks the address")
		http.Error(w, "", http.StatusInternalServerError)
		return
	case !tracks:
		http.Error(w, "address is not tracked", http.StatusUnauthorized)
		return
	}

	outs, err := ctx.Storage(r.Context()).GetUtxos(address)
	if err != nil {
		ctx.Logger(r.Context()).WithError(err).Errorf("failed to get unspent outs from db: %s", err.Error())
		http.Error(w, "", http.StatusInternalServerError)
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(outs)
}
