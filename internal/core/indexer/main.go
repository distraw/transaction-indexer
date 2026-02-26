package indexer

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/distraw/transaction-indexer/internal/config"
	"github.com/distraw/transaction-indexer/internal/core/node"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/logan/v3"
	"golang.org/x/sync/errgroup"
)

var (
	ErrAlreadyStarted = errors.New("indexer already started")
)

type Indexer interface {
	Run() error

	IsStarted() bool
	IsCatchedUp() bool

	Start() error
}

type indexer struct {
	log *logan.Entry

	node node.Node

	context context.Context
	storage data.Storage

	// TODO: delete this
	rpc *rpcclient.Client

	start     chan struct{}
	started   atomic.Bool
	catchedUp atomic.Bool

	pollFrequency    time.Duration
	initialBlockHash *chainhash.Hash
	netParams        chaincfg.Params

	currentHash *chainhash.Hash
}

func (i *indexer) Run() error {
	select {
	case <-i.start:
	case <-i.context.Done():
		return i.context.Err()
	}

	i.started.Store(true)
	i.log.Infof("Starting from block %s", i.initialBlockHash)

	initialBlockHash, err := i.rpc.GetBlockHash(0)
	if err != nil {
		return errors.Wrap(err, "failed to get initial block hash")
	}

	i.initialBlockHash = initialBlockHash

	err = i.catchUp(0)
	if err != nil {
		return errors.Wrap(err, "failed to catch-up to initial block height")
	}

	block, err := i.storage.Blocks().GetTip()
	if err != nil && !errors.Is(err, data.ErrNotFound) {
		return errors.Wrap(err, "failed to get highest block from db")
	}
	// if error is data.ErrNotFound, there are no previous blocks in db (db is clean),
	// so we leave currentHash empty for poll() to get first block
	if err == nil {
		i.currentHash, err = chainhash.NewHashFromStr(block.Hash)
		if err != nil {
			return errors.Wrap(err, "failed to generate hash from string")
		}
	}

	g, ctx := errgroup.WithContext(i.context)

	g.Go(func() error {
		err = i.node.SubscribeToBlocks(ctx, i.poll)
		if err != nil {
			return errors.Wrap(err, "failed to subscribe to blocks")
		}
		return nil
	})

	return g.Wait()
}

func (i *indexer) IsStarted() bool {
	return i.started.Load()
}

func (i *indexer) IsCatchedUp() bool {
	return i.catchedUp.Load()
}

func (i *indexer) Start() error {
	if !i.started.CompareAndSwap(false, true) {
		return ErrAlreadyStarted
	}

	close(i.start)
	return nil
}

func New(context context.Context, storage data.Storage, log *logan.Entry,
	rpc *rpcclient.Client, info config.IndexerInfo) Indexer {
	return &indexer{
		context: context,
		storage: storage.New(),
		log:     log,
		rpc:     rpc,

		// TODO
		node: node.NewRPC(log, rpc, info.GetPollFrequency(), info.GetNet(), info.GetInitialBlockHash()),

		pollFrequency:    info.GetPollFrequency(),
		initialBlockHash: info.GetInitialBlockHash(),
		netParams:        info.GetNet(),

		start: make(chan struct{}),
	}
}
