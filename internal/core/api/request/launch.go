package request

import (
	"net/http"

	"github.com/distraw/transaction-indexer/internal/core/indexer"
)

var (
	alreadyLaunched = false
)

// TODO: ping DB
func Launch(w http.ResponseWriter, r *http.Request) {
	if alreadyLaunched {
		http.Error(w, "409 Conflict (already launched)", http.StatusConflict)
		return
	}

	indexer.Start <- true
	w.WriteHeader(http.StatusOK)
}
