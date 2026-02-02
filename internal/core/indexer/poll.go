package indexer

import (
	"context"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/distraw/transaction-indexer/internal/core/api/ctx"
	"github.com/distraw/transaction-indexer/internal/data"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

var (
	ErrUnusualScript = errors.New("non-usual script detected")
)

func processTransactions(c context.Context, block *btcjson.GetBlockVerboseTxResult) error {
	storage := ctx.Storage(c)
	utxoStorage := storage.Utxos()
	log := ctx.Logger(c)

	for _, tx := range block.Tx {
		for _, in := range tx.Vin {
			txid := in.Txid
			vout := in.Vout

			exists, err := storage.Utxos().Exists(txid, int(vout))
			if err != nil {
				return errors.Wrap(err, "failed to check utxo existence in db")
			}
			if !exists {
				continue
			}

			utxo, err := storage.Utxos().Get(txid, int(vout))
			if err != nil {
				return errors.Wrap(err, "failed to fetch utxo from storage")
			}

			log.Info("deleted utxo")
			storage.Utxos().Delete(utxo.ID)
		}

		for _, out := range tx.Vout {
			spk := out.ScriptPubKey.Hex

			log.WithField("key", spk).Info("checking for existence...")
			exists, err := storage.Addresses().Exists(spk)
			if err != nil {
				return errors.Wrap(err, "failed to check if address exists")
			}
			if !exists {
				continue
			}

			dbAddress, err := storage.Addresses().Get(spk)
			if err != nil {
				return errors.Wrap(err, "failed to fetch address from storage")
			}

			// we need block ID in db in order to insert UTXO correctly,
			// thus, block must exist in db so we can reference it upon inserting UTXO
			exists, err = storage.Blocks().Exists(block.Hash)
			if err != nil {
				return errors.Wrap(err, "failed to check block existence in storage")
			}
			if !exists {
				return errors.New("processed block does not exist in storage")
			}

			dbBlock, err := storage.Blocks().Get(block.Hash)
			if err != nil {
				return errors.Wrap(err, "failed to fetch block from storage")
			}

			utxoStorage.Insert(data.Utxo{
				Txid: tx.Txid,
				Vout: int(out.N),

				BlockID:   dbBlock.ID,
				AddressID: dbAddress.ID,
			})
		}
	}

	return nil
}

// routinePoll() checks for updates on node and processes every new block if it is created
func routinePoll(c context.Context, initialHash *chainhash.Hash) func() error {
	storage := ctx.Storage(c)
	rpc := ctx.RPC(c)
	log := ctx.Logger(c)

	currentHash := initialHash

	return func() error {
		newHash, err := rpc.GetBestBlockHash()
		if err != nil {
			return errors.Wrap(err, "failed to poll best block hash")
		}
		if newHash.IsEqual(currentHash) {
			// new block was not mined yet
			return nil
		}

		log.
			WithField("prev_hash", currentHash).
			WithField("new_hash", newHash).
			Info("new best block detected")

		newBlock, err := rpc.GetBlockVerboseTx(newHash)
		if err != nil {
			return errors.Wrap(err, "failed to fetch latest block via rpc")
		}

		storage.Blocks().Insert(data.Block{
			Hash:   newHash.String(),
			Height: int32(newBlock.Height),
		})

		err = processTransactions(c, newBlock)
		if err != nil {
			return errors.Wrap(err, "failed to process transactions")
		}

		currentHash = (*chainhash.Hash)(newHash.CloneBytes())
		return nil
	}
}
