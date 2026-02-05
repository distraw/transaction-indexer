package indexer

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/pkg/errors"
)

// routinePoll polls every once in a while and processes every new incoming block
func (i *indexer) routinePoll() func() error {
	var currentHash *chainhash.Hash

	return func() error {
		newHash, err := i.rpc.GetBestBlockHash()
		if err != nil {
			return errors.Wrap(err, "failed to poll best block hash")
		}
		if newHash.IsEqual(currentHash) {
			// new block was not mined yet
			return nil
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
