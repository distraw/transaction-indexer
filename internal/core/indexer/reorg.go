package indexer

import (
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/pkg/errors"
)

func (i *indexer) reorganize(fromBlock *btcjson.GetBlockHeaderVerboseResult) error {
	const maxRollBack int = 100

	var currentBlock *btcjson.GetBlockHeaderVerboseResult = fromBlock
	for j := 0; j < maxRollBack; j++ {
		exists, err := i.storage.Blocks().Exists(currentBlock.Hash)
		if err != nil {
			return errors.Wrap(err, "failed to check block existence at set hash")
		}

		if exists {
			err = i.storage.DeleteBlocksAfter(currentBlock.Hash)
			if err != nil {
				return errors.Wrap(err, "failed to delete blocks upon set height")
			}

			err = i.catchUp(int64(currentBlock.Height + 1))
			if err != nil {
				return errors.Wrap(err, "failed to catch up")
			}

			return nil
		}

		prevHash, err := chainhash.NewHashFromStr(currentBlock.PreviousHash)
		if err != nil {
			return errors.Wrap(err, "failed to generate hash from string")
		}

		currentBlock, err = i.rpc.GetBlockHeaderVerbose(prevHash)
		if err != nil {
			return errors.Wrap(err, "failed to get verbose block header from rpc client")
		}
	}

	panic("deep reorg occured!!")
}
