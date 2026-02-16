package request

import (
	"errors"
	"net/http"

	"github.com/distraw/transaction-indexer/internal/core/api/ctx"
	"github.com/distraw/transaction-indexer/internal/core/indexer"
)

func Start(w http.ResponseWriter, r *http.Request) {
	err := ctx.Indexer(r.Context()).Start()
	if errors.Is(err, indexer.ErrAlreadyStarted) {
		http.Error(w, "indexer already started", http.StatusConflict)
		return
	}
	if err != nil {
		ctx.Logger(r.Context()).WithError(err).Error("unexpected error occured during indexer startup")
		http.Error(w, "unable to start indexer due to unknown error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
