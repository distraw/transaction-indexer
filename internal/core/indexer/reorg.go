package indexer

import (
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func (i *indexer) reorganize(fromBlock *btcjson.GetBlockHeaderVerboseResult) error {
	const maxRollBack int = 200

	var currentBlock *btcjson.GetBlockHeaderVerboseResult = fromBlock
	for j := 0; j < maxRollBack; j++ {
		exists, err := i.storage.Blocks().Exists(currentBlock.Hash)
		if err != nil {
			return err
		}

		if exists {
			err = i.storage.Utxos().MarkUnspentAboveHeight(currentBlock.Height)
			if err != nil {
				return err
			}

			err = i.storage.Blocks().DeleteUpon(currentBlock.Height)
			if err != nil {
				return err
			}

			return i.catchUp(int64(currentBlock.Height + 1))
		}

		prevHash, err := chainhash.NewHashFromStr(currentBlock.PreviousHash)
		if err != nil {
			return err
		}

		currentBlock, err = i.rpc.GetBlockHeaderVerbose(prevHash)
		if err != nil {
			return err
		}
	}

	panic("deep reorg occured!!")
}
