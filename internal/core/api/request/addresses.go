package request

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/josemiguelmelo/btcaddressvalidator"
)

var (
	ErrIllegalContentType = errors.New("provide content type is not supported")
)

func parseAddr(body []byte, contentType string) (string, error) {
	var addr string

	switch contentType {
	case "plain/text; charset=utf-8":
		addr = string(body)

	case "application/json":
		var parsedBody struct {
			Addr string `json:"addr"`
		}
		json.Unmarshal(body, &parsedBody)

		addr = parsedBody.Addr

	default:
		return "", ErrIllegalContentType
	}

	_, err := btcaddressvalidator.CheckBtcAddress(addr)
	if err != nil {
		return "", err
	}

	return addr, nil
}

func Addresses(w http.ResponseWriter, r *http.Request) {

}
