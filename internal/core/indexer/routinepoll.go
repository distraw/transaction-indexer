package indexer

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/pkg/errors"
)

// routinePoll polls every once in a while and processes every new incoming block
func (i *indexer) poll() error {
	newHash, err := i.rpc.GetBestBlockHash()
	if err != nil {
		return errors.Wrap(err, "failed to get best block hash")
	}
	if newHash.IsEqual(i.currentHash) {
		// new block was not mined yet
		return nil
	}

	newBlockHeader, err := i.rpc.GetBlockHeaderVerbose(newHash)
	if err != nil {
		return errors.Wrap(err, "failed to get verbose block header")
	}
	if i.currentHash != nil && newBlockHeader.PreviousHash != i.currentHash.String() {
		err = i.reorganize(newBlockHeader)
		if err != nil {
			return errors.Wrap(err, "failed to reorganize")
		}
	}

	i.log.
		WithField("prev_hash", i.currentHash).
		WithField("new_hash", newHash).
		Info("new best block detected")

	err = i.processBlock(newHash)
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
