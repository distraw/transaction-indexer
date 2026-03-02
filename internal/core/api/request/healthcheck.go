package request

import (
	"encoding/json"
	"net/http"

	"github.com/distraw/transaction-indexer/internal/core/api/ctx"
)

func Healthcheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"alive":     true,
		"started":   ctx.Indexer(r.Context()).IsStarted(),
		"catchedUp": ctx.Indexer(r.Context()).IsCatchedUp(),
	})
}
