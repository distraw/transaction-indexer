package request

import (
	"encoding/json"
	"net/http"
)

// TODO: ping DB
func Healthcheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"alive": true,
	})
}
