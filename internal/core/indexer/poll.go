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

	exists, err := i.storage.Blocks().Exists(newHash.String())
	if err != nil {
		return errors.Wrap(err, "failed to check block existence on local chain")
	}
	if exists {
		return nil
	}

	newBlock, err := i.node.GetBlock(newHash)
	if err != nil {
		return errors.Wrap(err, "failed to get verbose block header")
	}
	if !(i.currentHash.IsEqual(nil) || i.currentHash.IsEqual(&chainhash.Hash{})) && !newBlock.Header.PrevBlock.IsEqual(i.currentHash) {
		i.log.Infof("Start syncing because new hash is %s", newHash)
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

	err = i.processBlock(newBlock)
	if errors.Is(err, data.ErrAlreadyExists) {
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "failed to process block")
	}

	return nil
}
