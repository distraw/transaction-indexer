package indexer

import (
	"encoding/hex"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/distraw/transaction-indexer/internal/core/bitcoin"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/pkg/errors"
)

// getInputInfo links given input to its output using local utxo set.
//
// If no output were found, getInputInfo returns sender="unknown", value=-1, e=nil,
// since it is expected behaviour when indexer starts not from genesis block
// (output from block before indexer started may have been spent)
func (i *indexer) getInputInfo(input *wire.TxIn) (sender string, value float64, e error) {
	if bitcoin.IsCoinbase(input) {
		return "coinbase", 0, nil
	}

	output, err := i.storage.Outputs().Get(input.PreviousOutPoint.Hash.String(), input.PreviousOutPoint.Index)
	if errors.Is(err, data.ErrNotFound) {
		return "unknown", -1, nil
	}
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get output from local storage")
	}

	return output.Address, output.Value, nil
}

func (i *indexer) processInputs(vin []*wire.TxIn, dbTransactionID int) error {
	for j, in := range vin {
		senderAddress, value, err := i.getInputInfo(in)
		if err != nil {
			return errors.Wrap(err, "failed to get sender address for input")
		}

		i.storage.Inputs().Insert(data.Input{
			FromAddress:   senderAddress,
			Value:         value,
			TransactionID: dbTransactionID,
			Vin:           j,
		})

		if bitcoin.IsCoinbase(in) {
			continue
		}

		exists, err := i.storage.Outputs().Exists(in.PreviousOutPoint.Hash.String(), in.PreviousOutPoint.Index)
		if err != nil {
			return errors.Wrap(err, "failed to check output existence in db")
		}
		if !exists {
			continue
		}

		err = i.storage.Outputs().MarkSpent(in.PreviousOutPoint.Hash.String(), in.PreviousOutPoint.Index, int32(dbTransactionID))
		if err != nil {
			return errors.Wrap(err, "failed to mark outputs spent in db")
		}
	}

	return nil
}

func (i *indexer) processOutputs(vout []*wire.TxOut, txid string, dbTransactionID int) error {
	for j, out := range vout {
		receiverAddress, err := i.getAddressFromScriptPubKey(out.PkScript)
		if err != nil {
			return errors.Wrap(err, "failed to get bitcoin address from its scriptPubKey")
		}

		i.storage.Outputs().Insert(data.Output{
			Txid: txid,
			Vout: int(j),

			Value: bitcoin.SatoshisToBTC(out.Value),

			TransactionID: dbTransactionID,
			Address:       receiverAddress,
		})
	}

	return nil
}

// checkTxForTrackedAddrs checks if providen transaction contains at least one
// tracked address
func (i *indexer) checkTxForTrackedAddrs(tx *wire.MsgTx) (bool, error) {
	for _, out := range tx.TxOut {
		exists, err := i.storage.Addresses().ExistsByScriptPubKey(hex.EncodeToString(out.PkScript))
		if err != nil {
			return false, errors.Wrap(err, "failed to fetch address from storage")
		}
		if !exists {
			continue
		}

		return true, nil
	}

	for _, in := range tx.TxIn {
		if bitcoin.IsCoinbase(in) {
			continue
		}

		exists, err := i.storage.Outputs().Exists(in.PreviousOutPoint.Hash.String(), in.PreviousOutPoint.Index)
		if err != nil {
			return false, errors.Wrap(err, "failed to check utxo existence in db")
		}
		if !exists {
			continue
		}

		return true, nil
	}

	return false, nil
}

func (i *indexer) processTransactions(block *wire.MsgBlock, timestamp int64, dbBlockID int) error {
	for _, tx := range block.Transactions {
		dbTransactionID, err := i.storage.Transactions().Insert(data.Transaction{
			Txid:      tx.TxID(),
			BlockID:   dbBlockID,
			Locktime:  tx.LockTime,
			Timestamp: time.Unix(timestamp, 0).UTC(),
		})
		if err != nil {
			return errors.Wrap(err, "failed to insert new transaction into db")
		}

		// Always process outputs and save them in db since they are
		// needed to maintain local utxo set
		err = i.processOutputs(tx.TxOut, tx.TxID(), *dbTransactionID)
		if err != nil {
			return errors.Wrap(err, "failed to process transaction outputs")
		}

		containsTrackedAddrs, err := i.checkTxForTrackedAddrs(tx)
		if err != nil {
			return errors.Wrap(err, "failed to check if tx contains tracked addresses")
		}
		if !containsTrackedAddrs {
			continue
		}

		err = i.processInputs(tx.TxIn, *dbTransactionID)
		if err != nil {
			return errors.Wrap(err, "failed to process transaction inputs")
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

	var prevBlockID *int32
	switch prevBlock, err := i.storage.Blocks().Get(header.PreviousHash); err {
	case nil:
		prevBlockID = &prevBlock.ID
	case data.ErrNotFound:
		prevBlockID = nil
	default:
		return errors.Wrap(err, "failed to get previous block from storage")
	}

	dbBlockID, err := i.storage.Blocks().Insert(data.Block{
		Hash:            header.Hash,
		PreviousBlockID: prevBlockID,
	})
	if errors.Is(err, data.ErrAlreadyExists) {
		return errors.Wrap(err, "block already exists in db")
	}
	if err != nil {
		return errors.Wrap(err, "failed to insert new block into storage")
	}

	block, err := i.rpc.GetBlock(blockHash)
	if err != nil {
		return errors.Wrap(err, "failed to get verbose tx block from rpc client")
	}

	err = i.processTransactions(block, block.Header.Timestamp.Unix(), *dbBlockID)
	if err != nil {
		return errors.Wrap(err, "failed to process transactions")
	}

	return nil
}
