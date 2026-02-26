package indexer

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/pkg/errors"
)

// routinePoll polls every once in a while and processes every new incoming block
func (i *indexer) poll(newHash *chainhash.Hash) error {
	i.log.
		WithField("prev_hash", i.currentHash).
		WithField("new_hash", newHash).
		Info("new block detected")

	newBlockHeader, err := i.rpc.GetBlockHeaderVerbose(newHash)
	if err != nil {
		return errors.Wrap(err, "failed to get verbose block header")
	}
	if i.currentHash != nil && newBlockHeader.PreviousHash != i.currentHash.String() {
		err = i.reorganize(newBlockHeader)
		if err != nil {
			return errors.Wrap(err, "failed to reorganize")
		}

		return nil
	}

	err = i.processBlock(*newHash)
	if errors.Is(err, data.ErrAlreadyExists) {
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "failed to process block")
	}

	i.currentHash, err = chainhash.NewHashFromStr(newHash.String())
	if err != nil {
		return errors.Wrapf(err, "failed to generate hash from string %s", newHash.String())
	}

	return nil
}
