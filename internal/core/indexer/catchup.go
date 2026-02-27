package indexer

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/pkg/errors"
)

func (i *indexer) catchUp(initialHash *chainhash.Hash) error {
	i.catchedUp.Store(false)

	headers, err := i.node.GetHeaders([]*chainhash.Hash{initialHash}, nil)
	if err != nil {
		return errors.Wrap(err, "failed to get headers from initial hash to the tip")
	}

	i.log.Debugf("Catch-up started. %d blocks to process", len(headers))
	for _, header := range headers {
		blockHash := header.BlockHash()
		block, err := i.node.GetBlock(&blockHash)
		if err != nil {
			return errors.Wrap(err, "failed to get block from rpc node")
		}

		err = i.processBlock(block)
		if err != nil {
			return errors.Wrapf(err, "failed to process block %s", header.BlockHash().String())
		}
	}

	i.log.Debug("Catch-up finished.")
	i.catchedUp.Store(true)
	return nil
}
