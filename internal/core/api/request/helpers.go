package request

import (
	"encoding/json"
	"mime"

	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/pkg/errors"
)

func parseAddr(body []byte, contentType string) (string, error) {
	var addr string

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse content type")
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
			return "", errors.Wrap(err, "failed to unmarshal json body")
		}

		addr = parsedBody.Addr

	default:
		return "", ErrUnsupportedMediaType
	}

	return addr, nil
}

// extractRawAddresses extracts only string addresses from the db address struct
func extractRawAddresses(dbAddresses []data.Address) []string {
	rawAddresses := make([]string, len(dbAddresses))
	for i, a := range dbAddresses {
		rawAddresses[i] = a.Addr
	}

	return rawAddresses
}
