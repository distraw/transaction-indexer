package indexer

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/pkg/errors"
)

func dbBlocksToLocators(tip *data.Block, checkpoint *data.Block) ([]*chainhash.Hash, error) {
	tipHash, err := chainhash.NewHashFromStr(tip.Hash)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get hash from tip block")
	}

	checkpointHash, err := chainhash.NewHashFromStr(checkpoint.Hash)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get hash from checkpoint block")
	}

	return []*chainhash.Hash{tipHash, checkpointHash}, nil
}

func (i *indexer) synchronizeChain() error {
	startBlock, err := i.storage.GetBlockOnDepth(6)
	i.log.Infof("6-depth is %s", startBlock.Hash)
	if err != nil {
		return errors.Wrap(err, "failed to get block from local chain on depth 6")
	}

	tipBlock, err := i.storage.Blocks().GetTip()
	i.log.Infof("Tip %s", tipBlock.Hash)
	if err != nil {
		return errors.Wrap(err, "failed to get tip of the local chain")
	}

	locators, err := dbBlocksToLocators(tipBlock, startBlock)
	if err != nil {
		return errors.Wrap(err, "failed to get block locators from tip and checkpoint blocks")
	}
	i.log.Infof("locators are %s, %s", locators[0], locators[1])

	headers, err := i.node.GetHeaders(locators, nil)
	if err != nil {
		return errors.Wrap(err, "failed to get headers from rpc node")
	}
	i.log.Infof("header[0] is %s, header[0].prevBlock is %s", headers[0].BlockHash(), headers[0].PrevBlock)

	if len(headers) == 0 {
		return nil
	}

	firstHeaderHash := headers[0].BlockHash()
	startBlockHash, err := chainhash.NewHashFromStr(startBlock.Hash)
	if err != nil {
		return errors.Wrap(err, "failed to get hash from starting block")
	}

	if firstHeaderHash.IsEqual(i.netParams.GenesisHash) && !startBlockHash.IsEqual(i.netParams.GenesisHash) {
		panic("Reorganization deeper than 6 blocks occured")
	}

	if !headers[0].PrevBlock.IsEqual(&chainhash.Hash{}) && !headers[0].PrevBlock.IsEqual(nil) {
		err = i.storage.DeleteBlocksAfter(headers[0].PrevBlock.String())
		if err != nil {
			return errors.Wrapf(err, "failed to delete blocks after %s", headers[0].PrevBlock.String())
		}
	}

	var commonAncestorHash chainhash.Hash
	switch {
	case headers[0].PrevBlock.IsEqual(&chainhash.Hash{}) || headers[0].PrevBlock.IsEqual(nil):
		commonAncestorHash = headers[0].BlockHash()
	default:
		commonAncestorHash = headers[0].PrevBlock
	}

	err = i.catchUp(&commonAncestorHash)
	if err != nil {
		return errors.Wrapf(err, "failed to catch-up from local chain tip, previous blockhash=%s", headers[0].PrevBlock.String())
	}

	return nil
}
