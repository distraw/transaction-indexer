package request

import (
	"encoding/json"
	"net/http"
)

// TODO: ping DB
func Healthcheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(contentTypeHeader, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"alive": true,
	})
}
