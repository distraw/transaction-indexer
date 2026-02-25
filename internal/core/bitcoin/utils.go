package bitcoin

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/josemiguelmelo/btcaddressvalidator"
	"github.com/pkg/errors"
)

var (
	ErrNoAddress         = errors.New("no addresses were extracted from script pub key")
	ErrMultipleAddresses = errors.New("multiple addresses were extracted from script pub key")
)

func IsCoinbase(in *wire.TxIn) bool {
	return in.PreviousOutPoint.Hash.IsEqual(&chainhash.Hash{}) &&
		in.PreviousOutPoint.Index == 0xffffffff
}

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

// ToAddress decodes providen scriptPubKey (hex format) to bitcoin address
//
// Arg params should be set according to the net on which indexer is running,
// which is chaincfg.MainNetParams for mainnet, chaincfg.TestNet<3/4>Params for testnet,
// and chaincfg.RegressionNetParams for regtest
//
// Returns ErrNoAddress if zero addresses were extracted.
//
// Returns ErrMultipleAddresses if multiple addresses were extracted.
func ToAddress(scriptPubKey []byte, params *chaincfg.Params) (string, error) {
	_, addresses, _, err := txscript.ExtractPkScriptAddrs(scriptPubKey, params)
	if err != nil {
		return "", errors.Wrap(err, "failed to extract addresses from script")
	}

	if len(addresses) == 0 {
		return "", ErrNoAddress
	}

	if len(addresses) > 1 {
		return "", ErrMultipleAddresses
	}

	return addresses[0].EncodeAddress(), nil
}

func SatoshisToBTC(sats int64) float64 {
	return float64(sats) / 100_000_000.0
}
