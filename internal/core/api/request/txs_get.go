package request

import (
	"encoding/json"
	"net/http"

	"github.com/distraw/transaction-indexer/internal/core/api/ctx"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/go-chi/chi/v5"
	"github.com/josemiguelmelo/btcaddressvalidator"
)

func GetTXs(w http.ResponseWriter, r *http.Request) {
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

	txs, err := ctx.Storage(r.Context()).GetTxs(address)
	if err != nil {
		ctx.Logger(r.Context()).WithError(err).Errorf("failed to get transactions from db: %s", err.Error())
		http.Error(w, "", http.StatusInternalServerError)
	}

	var responseBody = make([]struct {
		Tx      data.Transaction `json:"transaction"`
		Inputs  []data.Input     `json:"inputs"`
		Outputs []data.Output    `json:"outputs"`
	}, len(txs))

	for i, tx := range txs {
		responseBody[i].Tx = tx

		inputs, err := ctx.Storage(r.Context()).GetInputsInTransaction(tx.Txid)
		if err != nil {
			ctx.Logger(r.Context()).WithError(err).Errorf("failed to get all inputs in transaction %s", tx.Txid)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		outputs, err := ctx.Storage(r.Context()).GetOutputsInTransaction(tx.Txid)
		if err != nil {
			ctx.Logger(r.Context()).WithError(err).Errorf("failed to get all outputs in transaction %s", tx.Txid)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		responseBody[i].Inputs = inputs
		responseBody[i].Outputs = outputs
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseBody)
}
