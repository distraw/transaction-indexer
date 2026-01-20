package request

import (
	"encoding/json"
	"net/http"

	"github.com/distraw/transaction-indexer/internal/core/api/ctx"
)

func GetAddresses(w http.ResponseWriter, r *http.Request) {
	c := r.Context()

	db := ctx.DB(c)
	userID := ctx.UserID(c)

	addresses, err := db.GetAddresses(*userID)
	if err != nil {
		ctx.Logger(c).
			WithError(err).
			WithField("user_id", userID).
			Error("failed unexpectedly to get all addresses related to the user")
		http.Error(w, "500 internal server error", http.StatusInternalServerError)
		return
	}

	addrs := make([]string, len(addresses))
	for i, a := range addresses {
		addrs[i] = a.Addr
	}

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(addrs)
}
