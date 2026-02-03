package request

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/distraw/transaction-indexer/internal/core/api/ctx"
	"github.com/distraw/transaction-indexer/internal/core/bitcoin"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/josemiguelmelo/btcaddressvalidator"
)

var (
	ErrUnsupportedMediaType = errors.New("Provided media type is unsupported")
)

func parseAddr(body []byte, contentType string) (string, error) {
	var addr string

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", err
	}

	switch mediaType {
	case "text/plain":
		addr = string(body)

	case "application/json":
		var parsedBody struct {
			Addr string `json:"addr"`
		}
		err := json.Unmarshal(body, &parsedBody)
		if err != nil {
			return "", err
		}

		addr = parsedBody.Addr

	default:
		return "", ErrUnsupportedMediaType
	}

	return addr, nil
}

func PostAddresses(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	log := ctx.Logger(c)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.WithError(err).Error("failed unexpectedly to parse request body")
		http.Error(w, "500 internal server error", http.StatusInternalServerError)
		return
	}

	addr, err := parseAddr(body, r.Header.Get("Content-type"))
	if err == ErrUnsupportedMediaType {
		http.Error(w, "415 unsupported media type", http.StatusUnsupportedMediaType)
		return
	}
	if err != nil {
		log.WithError(err).Error("failed unexpectedly to parse addr")
		http.Error(w, "500 internal server error", http.StatusInternalServerError)
		return
	}

	_, err = btcaddressvalidator.CheckBtcAddress(addr)
	if err != nil {
		http.Error(w, "422 unprocessable entity (invalid btc address format)", http.StatusUnprocessableEntity)
		return
	}

	scriptPubKey, err := bitcoin.ToScriptPubKey(addr)
	if err != nil {
		log.WithError(err).Error("failed unexpectedly to convert btc address into script public key")
		http.Error(w, "500 internal server error", http.StatusInternalServerError)
	}

	err = ctx.Storage(c).AddAddress(
		*ctx.UserID(c),
		data.Address{
			Addr:         addr,
			ScriptPubKey: scriptPubKey,
		},
	)
	if err == data.ErrAlreadyExists {
		http.Error(w, "409 conflict (address already exists)", http.StatusConflict)
		return
	}
	if err != nil {
		log.WithError(err).Error("failed unexpectedly to add new address")
		http.Error(w, "500 internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
