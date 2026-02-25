package indexer

import (
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/pkg/errors"
)

func (i *indexer) getInputInfo(in btcjson.Vin) (sender string, value float64, e error) {
	if in.IsCoinBase() {
		return "coinbase", 0, nil
	}

	outTxHash, err := chainhash.NewHashFromStr(in.Txid)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get hash from string in.Txid ")
	}

	// Since this output should be considered utxo in local chain,
	// we can simply get this utxo from db (marking it spent),
	// and linking our input to that utxo
	// 1. Maintain local utxo set

	// DELETE
	outTx, err := i.rpc.GetRawTransactionVerbose(outTxHash)
	if err != nil {
		return "", 0, errors.Wrapf(err, "failed to get transaction %s", in.Txid)
	}
	//

	fromAddress, err := i.getAddressFromScriptPubKey(outTx.Vout[in.Vout].ScriptPubKey.Hex)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to decode scriptPubKey from hex to bitcoin address")
	}

	return fromAddress, outTx.Vout[in.Vout].Value, nil
}

func (i *indexer) processInputs(vin []btcjson.Vin, dbTransactionID int, blockHeight int32) error {
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

		exists, err := i.storage.Outputs().Exists(in.Txid, in.Vout)
		if err != nil {
			return errors.Wrap(err, "failed to check utxo existence in db")
		}
		if !exists {
			continue
		}

		err = i.storage.Outputs().MarkSpent(in.Txid, int(in.Vout), blockHeight)
		if err != nil {
			return errors.Wrap(err, "failed to mark outputs spent in db")
		}
	}

	return nil
}

func (i *indexer) processOutputs(vout []btcjson.Vout, txid string, dbTransactionID int) error {
	for _, out := range vout {
		receiverAddress, err := i.getAddressFromScriptPubKey(out.ScriptPubKey.Hex)
		if err != nil {
			return errors.Wrap(err, "failed to get bitcoin address from its scriptPubKey")
		}

		i.storage.Outputs().Insert(data.Output{
			Txid: txid,
			Vout: int(out.N),

			Value: out.Value,

			TransactionID: dbTransactionID,
			Address:       receiverAddress,
		})
	}

	return nil
}

// checkTxForTrackedAddrs checks if providen transaction contains at least one
// tracked address
func (i *indexer) checkTxForTrackedAddrs(tx btcjson.TxRawResult) (bool, error) {
	for _, out := range tx.Vout {
		scriptPubKey := out.ScriptPubKey.Hex

		exists, err := i.storage.Addresses().ExistsByScriptPubKey(scriptPubKey)
		if err != nil {
			return false, errors.Wrap(err, "failed to fetch address from storage")
		}
		if !exists {
			continue
		}

		return true, nil
	}

	for _, in := range tx.Vin {
		exists, err := i.storage.Outputs().Exists(in.Txid, in.Vout)
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

func (i *indexer) processTransactions(block *btcjson.GetBlockVerboseTxResult, dbBlockID int) error {
	for _, tx := range block.Tx {
		dbTransactionID, err := i.storage.Transactions().Insert(data.Transaction{
			Txid:      tx.Txid,
			BlockID:   dbBlockID,
			Locktime:  tx.LockTime,
			Timestamp: time.Unix(tx.Time, 0).UTC(),
		})
		if err != nil {
			return errors.Wrap(err, "failed to insert new transaction into db")
		}

		containsTrackedAddrs, err := i.checkTxForTrackedAddrs(tx)
		if err != nil {
			return errors.Wrap(err, "failed to check if tx contains tracked addresses")
		}
		if !containsTrackedAddrs {
			continue
		}

		err = i.processInputs(tx.Vin, *dbTransactionID, int32(block.Height))
		if err != nil {
			return errors.Wrap(err, "failed to process transaction inputs")
		}

		err = i.processOutputs(tx.Vout, tx.Txid, *dbTransactionID)
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

	dbBlockID, err := i.storage.Blocks().Insert(data.Block{
		Hash:   header.Hash,
		Height: header.Height,
	})
	if errors.Is(err, data.ErrAlreadyExists) {
		return errors.Wrap(err, "block already exists in db")
	}
	if err != nil {
		return errors.Wrap(err, "failed to insert new block into storage")
	}

	block, err := i.rpc.GetBlockVerboseTx(blockHash)
	if err != nil {
		return errors.Wrap(err, "failed to get verbose tx block from rpc client")
	}

	err = i.processTransactions(block, *dbBlockID)
	if err != nil {
		return errors.Wrap(err, "failed to process transactions")
	}

	return nil
}
