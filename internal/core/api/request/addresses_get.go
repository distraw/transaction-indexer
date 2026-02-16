package request

import (
	"encoding/json"
	"net/http"

	"github.com/distraw/transaction-indexer/internal/core/api/ctx"
)

func GetAddresses(w http.ResponseWriter, r *http.Request) {
	dbAddresses, err := ctx.Storage(r.Context()).GetAddresses(*ctx.UserID(r.Context()))
	if err != nil {
		ctx.Logger(r.Context()).
			WithError(err).
			WithField("user_id", *ctx.UserID(r.Context())).
			Error("failed unexpectedly to get all addresses related to the user")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	rawAddresses := extractRawAddresses(dbAddresses)

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(rawAddresses)
}
