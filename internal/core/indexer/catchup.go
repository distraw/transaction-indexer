package indexer

import (
	"fmt"

	"github.com/pkg/errors"
)

// catchUp processes every block starting from initial height
// up to the best one included.
func (i *indexer) catchUp(initialHeight int64) error {
	i.catchedUp.Store(false)

	targetHeight, err := i.rpc.GetBlockCount()
	if err != nil {
		return errors.Wrap(err, "failed to get block count from rpc client")
	}
	if initialHeight > targetHeight {
		return errors.New(fmt.Sprintf(
			"initial polling height must be less than current block count (%d)",
			targetHeight,
		))
	}

	i.log.Infof("Catch-up started. %d blocks are estimated to process", targetHeight-initialHeight)
	for j := initialHeight; j <= targetHeight; j++ {
		i.currentHash, err = i.rpc.GetBlockHash(j)
		if err != nil {
			return errors.Wrapf(err, "failed to get block hash on height %d from rpc client", j)
		}

		err = i.processBlock(i.currentHash)
		if err != nil {
			return errors.Wrapf(err, "failed to process block %s", i.currentHash.String())
		}

		targetHeight, err = i.rpc.GetBlockCount()
		if err != nil {
			return errors.Wrap(err, "failed to get block count from rpc client")
		}
	}

	i.log.Infof("Catch-up finished. %d blocks processed", targetHeight-initialHeight)
	i.catchedUp.Store(true)
	return nil
}
