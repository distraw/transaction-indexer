package bitcoin

import (
	"encoding/hex"
	"errors"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/josemiguelmelo/btcaddressvalidator"
)

var (
	ErrInvalidAddressFormat = errors.New("Provided bitcoin address is invalid")
)

func ToScriptPubKey(addr string) (string, error) {
	_, err := btcaddressvalidator.CheckBtcAddress(addr)
	if err != nil {
		return "", ErrInvalidAddressFormat
	}

	decodedAddr, err := btcutil.DecodeAddress(addr, &chaincfg.MainNetParams)
	if err != nil {
		return "", err
	}

	scriptPubKey, err := txscript.PayToAddrScript(decodedAddr)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(scriptPubKey), nil
}
