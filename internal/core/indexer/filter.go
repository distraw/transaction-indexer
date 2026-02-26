package indexer

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil/gcs"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/distraw/transaction-indexer/internal/data"
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

// FilterBlock gets the block filter (defined in BIP 157/158)
// and checks whether the block with a given hash likely contains
// transactions with desired addresses
//
// List of desired addresses is retrieved from storage (table Addresses)
func (i *indexer) filterBlock(blockhash *chainhash.Hash) (bool, error) {
	scriptPubKeys, err := i.storage.Addresses().GetAllScriptPubKeys()
	if err != nil {
		// If storage.Addresses is empty, nothing to check in block
		if errors.Is(err, data.ErrNotFound) {
			return false, nil
		}

		return false, errors.Wrap(err, "failed to get all addresses from local storage")
	}

	decodedScriptPubKeys, err := decodeScriptPubKeys(scriptPubKeys)
	if err != nil {
		return false, errors.Wrap(err, "failed to decode scriptPubKeys")
	}

	filterType := btcjson.FilterTypeBasic
	filterResp, err := i.rpc.GetBlockFilter(*blockhash, &filterType)
	if err != nil {
		return false, errors.Wrap(err, "failed to get block filter")
	}

	filterBytes, err := hex.DecodeString(filterResp.Filter)
	if err != nil {
		return false, errors.Wrap(err, "failed to decode filter string into bytes")
	}

	// Default constants, defined in BIP158
	// More info at https://github.com/bitcoin/bips/blob/master/bip-0158.mediawiki
	const P = 19
	const M = 784931

	filter, err := gcs.FromNBytes(P, M, filterBytes)
	if err != nil {
		return false, errors.Wrap(err, "failed to deserialise a GCS filter")
	}

	var sipHashKey [16]byte
	copy(sipHashKey[:], blockhash[:16])

	matched, err := filter.MatchAny(sipHashKey, decodedScriptPubKeys)
	if err != nil {
		return false, errors.Wrap(err, "failed to check if filter matches any decoded scriptPubKey")
	}

	return matched, nil
}
