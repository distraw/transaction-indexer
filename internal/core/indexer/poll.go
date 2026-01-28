package indexer

import (
	"context"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/distraw/transaction-indexer/internal/core/api/ctx"
	"github.com/distraw/transaction-indexer/internal/data"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

var (
	ErrUnusualScript = errors.New("non-usual script detected")
)

func extractAddress(vout btcjson.Vout) (btcutil.Address, error) {
	pkScript, err := hex.DecodeString(vout.ScriptPubKey.Hex)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode hex")
	}

	_, addresses, _, err := txscript.ExtractPkScriptAddrs(pkScript, &chaincfg.MainNetParams)
	if err != nil {
		return nil, errors.Wrap(err, "failed to extract addresses from pkScript")
	}
	if len(addresses) != 0 {
		return nil, ErrUnusualScript
	}

	return addresses[0], nil
}

func processTransactions(c context.Context, txs []btcjson.TxRawResult) error {
	storage := ctx.Storage(c)
	log := ctx.Logger(c)

	for _, tx := range txs {
		for _, out := range tx.Vout {
			log.
				WithField("BTC", out.Value).Info("btc")

			address, err := extractAddress(out)
			if err == ErrUnusualScript {
				break
			}
			if err != nil {
				return errors.Wrap(err, "failed to extract address from vout")
			}

			exists, err := storage.Addresses().Exists(address.String())
			if err != nil {
				return errors.Wrap(err, "failed to check if address exists")
			}
			if !exists {
				break
			}

			log.WithField("addr", address).Info("tracked addr detected")
		}
	}

	return nil
}

// routinePoll() checks for updates on node and processes every new block if it is created
func routinePoll(c context.Context, initialBlockHash *chainhash.Hash) func() error {
	storage := ctx.Storage(c)
	rpc := ctx.RPC(c)
	log := ctx.Logger(c)

	prevBlockHash := initialBlockHash

	return func() error {
		blockHash, err := rpc.GetBestBlockHash()
		if err != nil {
			return errors.Wrap(err, "failed to poll best block hash")
		}
		if blockHash.IsEqual(prevBlockHash) {
			return nil
		}

		log.
			WithField("prev_hash", prevBlockHash.String()).
			WithField("new_hash", blockHash.String()).
			Info("new hash detected")

		block, err := rpc.GetBlockVerboseTx(blockHash)
		if err != nil {
			return errors.Wrap(err, "failed to fetch latest block via rpc")
		}

		err = processTransactions(c, block.Tx)
		if err != nil {
			return errors.Wrap(err, "failed to process transactions")
		}

		storage.Blocks().Insert(data.Block{
			Hash:   blockHash.String(),
			Height: int32(block.Height),
		})

		prevBlockHash = (*chainhash.Hash)(blockHash.CloneBytes())
		return nil
	}
}
