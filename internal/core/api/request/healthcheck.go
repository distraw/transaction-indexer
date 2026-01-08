package request

import (
	"encoding/json"
	"net/http"
)

// TODO: ping DB
func Healthcheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"alive": true,
	})
}
