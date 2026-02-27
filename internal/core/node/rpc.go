package node

import (
	"context"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/madflojo/tasks"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/logan/v3"
)

type rpcNode struct {
	log       *logan.Entry
	client    *rpcclient.Client
	scheduler *tasks.Scheduler

	pollFrequency time.Duration
	netParams     chaincfg.Params

	currentBlockHash *chainhash.Hash
}

func (r *rpcNode) SubscribeToBlocks(context context.Context, callback func(newBlockHash *chainhash.Hash) error) error {
	errCh := make(chan error, 1)

	defer r.scheduler.Stop()
	_, err := r.scheduler.Add(&tasks.Task{
		Interval: r.pollFrequency,
		TaskFunc: func() error {
			return r.poll(callback)
		},
		ErrFunc: func(err error) {
			select {
			case errCh <- err:
			default:
			}
		},
	})
	if err != nil {
		return errors.Wrap(err, "scheduler failed unexpectedly")
	}

	select {
	case <-context.Done():
		return context.Err()
	case err := <-errCh:
		return errors.Wrap(err, "rpc poller failed")
	}
}

func (r *rpcNode) isNewBlock(newBlockHash *chainhash.Hash) bool {
	if newBlockHash.IsEqual(r.currentBlockHash) {
		return false
	}

	if (r.currentBlockHash.IsEqual(nil) || r.currentBlockHash.IsEqual(&chainhash.Hash{})) &&
		newBlockHash.IsEqual(r.netParams.GenesisHash) {
		return false
	}

	return true
}

func (r *rpcNode) poll(callback func(newBlockHash *chainhash.Hash) error) error {
	newBlockHash, err := r.client.GetBestBlockHash()
	if err != nil {
		return errors.Wrap(err, "failed to get best block hash")
	}

	if !r.isNewBlock(newBlockHash) {
		return nil
	}

	err = callback(newBlockHash)
	if err != nil {
		return errors.Wrap(err, "new block hash callback function failed")
	}

	r.currentBlockHash = (*chainhash.Hash)(newBlockHash.CloneBytes())

	return nil
}

func (r *rpcNode) getGenesisHeader() (*btcjson.GetBlockHeaderVerboseResult, error) {
	genesisHash, err := r.client.GetBlockHash(0)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get hash of genesis block from rpc node")
	}

	genesisHeader, err := r.client.GetBlockHeaderVerbose(genesisHash)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get header of genesis block from rpc node")
	}

	return genesisHeader, nil
}

func (r *rpcNode) doesHeaderExistOnNode(header *btcjson.GetBlockHeaderVerboseResult) (bool, error) {
	blockCount, err := r.client.GetBlockCount()
	if err != nil {
		return false, errors.Wrap(err, "failed to get block count on rpc node")
	}

	if header.Height > int32(blockCount) {
		return false, nil
	}

	competingHeader, err := r.client.GetBlockHash(int64(header.Height))
	if err != nil {
		return false, errors.Wrap(err, "failed to get header at the same height")
	}

	return competingHeader.String() == header.Hash, nil
}

func (r *rpcNode) findCommonAncestor(locators []*chainhash.Hash) (*btcjson.GetBlockHeaderVerboseResult, error) {
	var ancestor *btcjson.GetBlockHeaderVerboseResult
	var err error
	for i := 0; i < len(locators); i++ {
		ancestor, err = r.client.GetBlockHeaderVerbose(locators[i])
		if err != nil {
			return nil, errors.Wrap(err, "failed to get verbose block header from rpc node")
		}

		exists, err := r.doesHeaderExistOnNode(ancestor)
		if err != nil {
			return nil, errors.Wrap(err, "failed to check if block is missing on node")
		}

		if !exists {
			r.log.Infof("header is missing on node %s", locators[i])
			continue
		}

		return ancestor, nil
	}

	return nil, ErrNoCommonAncestor
}

func (r *rpcNode) getStopHeight(stop *chainhash.Hash) (int32, error) {
	if stop.IsEqual(&chainhash.Hash{}) || stop.IsEqual(nil) {
		bestBlockHash, err := r.client.GetBestBlockHash()
		if err != nil {
			return 0, errors.Wrap(err, "failed to get best block hash")
		}

		header, err := r.client.GetBlockHeaderVerbose(bestBlockHash)
		if err != nil {
			return 0, errors.Wrap(err, "failed to get best block header")
		}

		return header.Height, nil
	}

	header, err := r.client.GetBlockHeaderVerbose(stop)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get block header from rpc node")
	}

	return header.Height - 1, nil
}

func (r *rpcNode) getStartHeight(locators []*chainhash.Hash) (int32, error) {
	// if locator is nil hash, then genesis block should be starting point
	if locators[0].IsEqual(&chainhash.Hash{}) {
		return 0, nil
	}

	ancestor, err := r.findCommonAncestor(locators)
	if err != nil {
		return 0, errors.Wrap(err, "failed to find common ancestor on rpc node")
	}

	return ancestor.Height + 1, nil
}

func (r *rpcNode) GetHeaders(locators []*chainhash.Hash, stop *chainhash.Hash) ([]*wire.BlockHeader, error) {
	startHeight, err := r.getStartHeight(locators)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get starting height")
	}
	r.log.Infof("starting height is %d", startHeight)

	stopHeight, err := r.getStopHeight(stop)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get stophash header from rpc node")
	}
	r.log.Infof("stop height is %d", stopHeight)

	var headers []*wire.BlockHeader

	const maxHeadersPerRequest int = 2000
	for i, height := 0, startHeight; i < maxHeadersPerRequest && height <= stopHeight; i, height = i+1, height+1 {
		hash, err := r.client.GetBlockHash(int64(height))
		if err != nil {
			return nil, errors.Wrap(err, "failed to get block hash by its height from rpc node")
		}

		header, err := r.client.GetBlockHeader(hash)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get block header by its hash from rpc node")
		}

		headers = append(headers, header)
	}

	return headers, nil
}

func (r *rpcNode) GetBlock(hash *chainhash.Hash) (*wire.MsgBlock, error) {
	block, err := r.client.GetBlock(hash)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get block from rpc node")
	}

	return block, nil
}

func NewRPC(log *logan.Entry, rpc *rpcclient.Client, pollFrequency time.Duration,
	netParams chaincfg.Params, initialBlockHash *chainhash.Hash) Node {
	return &rpcNode{
		log:              log,
		client:           rpc,
		scheduler:        tasks.New(),
		pollFrequency:    pollFrequency,
		netParams:        netParams,
		currentBlockHash: initialBlockHash,
	}
}
