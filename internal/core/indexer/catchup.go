package indexer

import (
	"fmt"

	"github.com/pkg/errors"
)

// catchUp processes every block starting from initial height
// up to the best one before the best available block on the node.
//
// Intentionally does not process the best available block, as it would
// be processed during routinePoll
func (i *indexer) catchUp(initialHeight int64) error {
	targetHeight, err := i.rpc.GetBlockCount()
	if err != nil {
		return errors.New("failed to poll remote node for block count")
	}
	if initialHeight > targetHeight {
		return errors.New(fmt.Sprintf(
			"initial polling height must be less than current block count (%d)",
			targetHeight,
		))
	}

	i.log.Infof("Catch-up started. %d blocks are estimated to process", targetHeight-initialHeight)
	for j := initialHeight; j < targetHeight; j++ {
		blockHash, err := i.rpc.GetBlockHash(j)
		if err != nil {
			return errors.New(fmt.Sprintf("failed to poll remote node for block %d", j))
		}

		err = i.processBlock(blockHash)
		if err != nil {
			return errors.New(fmt.Sprintf("failed to process block %s", blockHash.String()))
		}

		targetHeight, err = i.rpc.GetBlockCount()
		if err != nil {
			return errors.New("failed to poll remote node for block count")
		}
	}

	i.log.Infof("Catch-up finished. %d blocks processed", targetHeight-initialHeight)
	return nil
}
