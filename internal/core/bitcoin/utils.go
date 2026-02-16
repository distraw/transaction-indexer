package bitcoin

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/josemiguelmelo/btcaddressvalidator"
	"github.com/pkg/errors"
)

func ToScriptPubKey(addr string) (string, error) {
	_, err := btcaddressvalidator.CheckBtcAddress(addr)
	if err != nil {
		return "", errors.Wrap(err, "failed to check btc address")
	}

	decodedAddr, err := btcutil.DecodeAddress(addr, &chaincfg.MainNetParams)
	if err != nil {
		return "", errors.Wrap(err, "failed to decode address")
	}

	scriptPubKey, err := txscript.PayToAddrScript(decodedAddr)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert to pay-to-address script")
	}

	return hex.EncodeToString(scriptPubKey), nil
}
