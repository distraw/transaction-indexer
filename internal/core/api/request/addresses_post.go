package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"

	"github.com/distraw/transaction-indexer/internal/core/api/ctx"
	"github.com/distraw/transaction-indexer/internal/data"
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

	fmt.Println(mediaType)

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

	storage := ctx.Storage(c)
	userID := ctx.UserID(c)

	err = storage.AddAddress(*userID, data.Address{Addr: addr})
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
