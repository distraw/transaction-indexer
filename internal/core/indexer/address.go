package indexer

import (
	"github.com/distraw/transaction-indexer/internal/core/bitcoin"
	"github.com/pkg/errors"
)

func (i *indexer) getAddressFromScriptPubKey(scriptPubKey []byte) (string, error) {
	address, err := bitcoin.ToAddress(scriptPubKey, &i.netParams)
	if errors.Is(err, bitcoin.ErrNoAddress) {
		return "not_found", nil
	}
	if errors.Is(err, bitcoin.ErrMultipleAddresses) {
		return "multiple", nil
	}
	if err != nil {
		return "", errors.Wrap(err, "failed to convert scriptPubKey (hex format) to bitcoin address")
	}

	return address, nil
}
