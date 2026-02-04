package indexer

import (
	"context"
	"fmt"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/distraw/transaction-indexer/internal/core/api/ctx"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/pkg/errors"
)

var (
	ErrUnusualScript = errors.New("non-usual script detected")
)

func processTransactions(c context.Context, block *btcjson.GetBlockVerboseTxResult) error {
	storage := ctx.Storage(c)
	utxos := storage.Utxos()
	log := ctx.Logger(c)

	for _, tx := range block.Tx {
		for _, in := range tx.Vin {
			exists, err := utxos.Exists(in.Txid, int(in.Vout))
			if err != nil {
				return errors.Wrap(err, "failed to check utxo existence in db")
			}
			if !exists {
				continue
			}

			utxo, err := utxos.Get(in.Txid, int(in.Vout))
			if err != nil {
				return errors.Wrap(err, "failed to fetch utxo from storage")
			}

			log.Info("deleted utxo")
			storage.Utxos().Delete(utxo.ID)
		}

		for _, out := range tx.Vout {
			scriptPubKey := out.ScriptPubKey.Hex

			dbAddress, err := storage.Addresses().GetByScriptPubKey(scriptPubKey)
			if errors.Is(err, data.ErrNotFound) {
				continue
			}
			if err != nil {
				return errors.Wrap(err, "failed to fetch address from storage")
			}

			dbBlock, err := storage.Blocks().Get(block.Hash)
			if err == data.ErrNotFound {
				return errors.Wrap(err, "processed block does not exist in storage")
			}
			if err != nil {
				return errors.Wrap(err, "failed to get block from storage")
			}

			utxos.Insert(data.Utxo{
				Txid: tx.Txid,
				Vout: int(out.N),

				BlockID:   dbBlock.ID,
				AddressID: dbAddress.ID,
			})
		}
	}

	return nil
}

func processBlock(c context.Context, blockHash *chainhash.Hash) error {
	rpc := ctx.RPC(c)
	storage := ctx.Storage(c)

	block, err := rpc.GetBlockVerboseTx(blockHash)
	if err != nil {
		return err
	}

	err = storage.Blocks().Insert(data.Block{
		Hash:   block.Hash,
		Height: int32(block.Height),
	})
	if errors.Is(err, data.ErrAlreadyExists) {
		return errors.Wrap(err, "block already exists in db")
	}
	if err != nil {
		return errors.Wrap(err, "failed to insert new block into storage")
	}

	err = processTransactions(c, block)
	if err != nil {
		return errors.Wrap(err, "failed to process transactions")
	}

	return nil
}

// catchUp processes every block starting from initial height
// up to the best one before the best available block on the node.
//
// Intentionally does not process the best available block, as it would
// be processed during routinePoll
func catchUp(c context.Context, initialHeight int64) error {
	rpc := ctx.RPC(c)

	targetHeight, err := rpc.GetBlockCount()
	if err != nil {
		return errors.New("failed to poll remote node for block count")
	}
	if initialHeight > targetHeight {
		return errors.New(fmt.Sprintf(
			"initial polling height must be less than current block count (%d)",
			targetHeight,
		))
	}

	for i := initialHeight; i < targetHeight; i++ {
		blockHash, err := rpc.GetBlockHash(i)
		if err != nil {
			return errors.New(fmt.Sprintf("failed to poll remote node for block %d", i))
		}

		err = processBlock(c, blockHash)
		if err != nil {
			return errors.New(fmt.Sprintf("failed to process block %s", blockHash.String()))
		}

		targetHeight, err = rpc.GetBlockCount()
		if err != nil {
			return errors.New("failed to poll remote node for block count")
		}
	}

	ctx.Logger(c).Infof("Catch-up finished. %d blocks processed", targetHeight-initialHeight)
	return nil
}

// routinePoll polls every once in a while and processes every new incoming block
func routinePoll(c context.Context) func() error {
	rpc := ctx.RPC(c)
	log := ctx.Logger(c)

	var currentHash *chainhash.Hash

	return func() error {
		newHash, err := rpc.GetBestBlockHash()
		if err != nil {
			return errors.Wrap(err, "failed to poll best block hash")
		}
		if newHash.IsEqual(currentHash) {
			// new block was not mined yet
			return nil
		}

		defer func() {
			currentHash, err = chainhash.NewHashFromStr(newHash.String())
			if err != nil {
				panic(fmt.Sprintf("failed unexpectedly to generate hash from string: %s", err.Error()))
			}
		}()

		log.
			WithField("prev_hash", currentHash).
			WithField("new_hash", newHash).
			Info("new best block detected")

		err = processBlock(c, newHash)
		if err != nil {
			return err
		}

		return nil
	}
}
