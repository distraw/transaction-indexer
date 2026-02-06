package indexer

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/pkg/errors"
)

// routinePoll polls every once in a while and processes every new incoming block
func (i *indexer) routinePoll(initialHash *chainhash.Hash) func() error {
	var currentHash *chainhash.Hash = initialHash

	return func() error {
		newHash, err := i.rpc.GetBestBlockHash()
		if err != nil {
			return errors.Wrap(err, "failed to poll best block hash")
		}
		if newHash.IsEqual(currentHash) {
			// new block was not mined yet
			return nil
		}

		newBlockHeader, err := i.rpc.GetBlockHeaderVerbose(newHash)
		if err != nil {
			return err
		}
		if currentHash != nil && newBlockHeader.PreviousHash != currentHash.String() {
			err := i.reorganize(newBlockHeader)
			if err != nil {
				return err
			}
		}

		defer func() {
			currentHash, err = chainhash.NewHashFromStr(newHash.String())
			if err != nil {
				panic(fmt.Sprintf("failed unexpectedly to generate hash from string: %s", err.Error()))
			}
		}()

		i.log.
			WithField("prev_hash", currentHash).
			WithField("new_hash", newHash).
			Info("new best block detected")

		err = i.processBlock(newHash)
		if err != nil {
			return err
		}

		return nil
	}
}
