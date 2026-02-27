package indexer

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/pkg/errors"
)

// routinePoll polls every once in a while and processes every new incoming block
func (i *indexer) poll(newHash *chainhash.Hash) error {
	if newHash.IsEqual(i.currentHash) {
		return nil
	}

	newBlockHeader, err := i.rpc.GetBlockHeaderVerbose(newHash)
	if err != nil {
		return errors.Wrap(err, "failed to get verbose block header")
	}
	if !(i.currentHash.IsEqual(nil) || i.currentHash.IsEqual(&chainhash.Hash{})) && newBlockHeader.PreviousHash != i.currentHash.String() {
		err = i.synchronizeChain()
		if err != nil {
			return errors.Wrap(err, "failed to synchronize remote and local chain")
		}

		return nil
	}

	i.log.
		WithField("prev_hash", i.currentHash).
		WithField("new_hash", newHash).
		Info("new block detected")

	err = i.processBlock(*newHash)
	if errors.Is(err, data.ErrAlreadyExists) {
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "failed to process block")
	}

	return nil
}
