package indexer

import (
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/pkg/errors"
)

var (
	ErrUnusualScript = errors.New("non-usual script detected")
)

func (i *indexer) processInputs(vin []btcjson.Vin, blockHeight int32) error {
	utxos := i.storage.Utxos()

	for _, in := range vin {
		exists, err := utxos.Exists(in.Txid, in.Vout)
		if err != nil {
			return errors.Wrap(err, "failed to check utxo existence in db")
		}
		if !exists {
			continue
		}

		err = utxos.MarkSpent(in.Txid, int(in.Vout), blockHeight)
		if err != nil {
			return errors.Wrap(err, "failed to mark utxos spent in db")
		}
	}

	return nil
}

func (i *indexer) processOutputs(vout []btcjson.Vout, block *btcjson.GetBlockVerboseTxResult, txid string) error {
	for _, out := range vout {
		scriptPubKey := out.ScriptPubKey.Hex

		dbAddress, err := i.storage.Addresses().GetByScriptPubKey(scriptPubKey)
		if errors.Is(err, data.ErrNotFound) {
			continue
		}
		if err != nil {
			return errors.Wrap(err, "failed to fetch address from storage")
		}

		dbBlock, err := i.storage.Blocks().Get(block.Hash)
		if err == data.ErrNotFound {
			return errors.Wrap(err, "processed block does not exist in storage")
		}
		if err != nil {
			return errors.Wrap(err, "failed to get block from storage")
		}

		i.storage.Utxos().Insert(data.Utxo{
			Txid: txid,
			Vout: int(out.N),

			Value: out.Value,

			BlockID:   dbBlock.ID,
			AddressID: dbAddress.ID,
		})
	}

	return nil
}

func (i *indexer) processTransactions(block *btcjson.GetBlockVerboseTxResult) error {
	for _, tx := range block.Tx {
		err := i.processInputs(tx.Vin, int32(block.Height))
		if err != nil {
			return errors.Wrap(err, "failed to process transaction inputs")
		}

		err = i.processOutputs(tx.Vout, block, tx.Txid)
		if err != nil {
			return errors.Wrap(err, "failed to process transaction outputs")
		}
	}

	return nil
}

func (i *indexer) processBlock(blockHash *chainhash.Hash) error {
	header, err := i.rpc.GetBlockHeaderVerbose(blockHash)
	if err != nil {
		return errors.Wrap(err, "failed to get verbose block header from rpc client")
	}

	err = i.validateBlockHeader(header)
	if err != nil {
		return errors.Wrap(err, "invalid block header")
	}

	block, err := i.rpc.GetBlockVerboseTx(blockHash)
	if err != nil {
		return errors.Wrap(err, "failed to get verbose tx block from rpc client")
	}

	err = i.storage.Blocks().Insert(data.Block{
		Hash:   block.Hash,
		Height: int32(block.Height),
	})
	if errors.Is(err, data.ErrAlreadyExists) {
		return errors.Wrap(err, "block already exists in db")
	}
	if err != nil {
		return errors.Wrap(err, "failed to insert new block into storage")
	}

	err = i.processTransactions(block)
	if err != nil {
		return errors.Wrap(err, "failed to process transactions")
	}

	return nil
}
