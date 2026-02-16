package request

import (
	"io"
	"net/http"

	"github.com/distraw/transaction-indexer/internal/core/api/ctx"
	"github.com/distraw/transaction-indexer/internal/core/bitcoin"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/josemiguelmelo/btcaddressvalidator"
	"github.com/pkg/errors"
)

var (
	ErrUnsupportedMediaType = errors.New("Provided media type is unsupported")
)

func AddAddressToTrack(w http.ResponseWriter, r *http.Request) {
	limit := io.LimitReader(r.Body, 1024)
	body, err := io.ReadAll(limit)
	if err != nil {
		http.Error(w, "corrupted body", http.StatusBadRequest)
		return
	}

	addr, err := parseAddr(body, r.Header.Get("Content-type"))
	if err == ErrUnsupportedMediaType {
		http.Error(w, "invalid content-type header", http.StatusUnsupportedMediaType)
		return
	}
	if err != nil {
		http.Error(w, "corrupted address", http.StatusBadRequest)
		return
	}

	_, err = btcaddressvalidator.CheckBtcAddress(addr)
	if err != nil {
		http.Error(w, "invalid btc address format", http.StatusUnprocessableEntity)
		return
	}

	scriptPubKey, err := bitcoin.ToScriptPubKey(addr)
	if err != nil {
		ctx.Logger(r.Context()).WithError(err).Error("failed to convert btc address into script public key")
		http.Error(w, "btc address can not be converted to scriptpubkey", http.StatusBadRequest)
		return
	}

	err = ctx.Storage(r.Context()).AddAddress(
		*ctx.UserID(r.Context()),
		data.Address{
			Addr:         addr,
			ScriptPubKey: scriptPubKey,
		},
	)
	if err == data.ErrAlreadyExists {
		http.Error(w, "address already exists", http.StatusConflict)
		return
	}
	if err != nil {
		ctx.Logger(r.Context()).WithError(err).Error("failed unexpectedly to add new address")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
