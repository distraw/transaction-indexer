package indexer

import (
	"encoding/hex"

	"github.com/distraw/transaction-indexer/internal/core/bitcoin"
	"github.com/pkg/errors"
)

func decodeScriptPubKeys(scriptPubKeys []string) ([][]byte, error) {
	decoded := make([][]byte, len(scriptPubKeys))

	var err error
	for i, spk := range scriptPubKeys {
		decoded[i], err = hex.DecodeString(spk)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode scriptPubKey from string to hex")
		}
	}

	return decoded, nil
}

func (i *indexer) getAddressFromScriptPubKey(scriptPubKeyHex string) (string, error) {
	address, err := bitcoin.ToAddress(scriptPubKeyHex, &i.netParams)
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
