package indexer

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/distraw/transaction-indexer/internal/config"
	"github.com/distraw/transaction-indexer/internal/core/node"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/distraw/transaction-indexer/internal/data/pg"
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

	start     chan struct{}
	started   atomic.Bool
	catchedUp atomic.Bool

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
	i.log.Infof("Starting from block \"%s\"", i.initialBlockHash)

	g, ctx := errgroup.WithContext(i.context)

	g.Go(func() error {
		err := i.node.Subscribe(ctx, i.initialBlockHash, i.netParams, i.poll)
		if err != nil {
			return errors.Wrap(err, "failed to subscribe to blocks")
		}
		return nil
	})

	time.Sleep(time.Second)

	err := i.catchUp(i.initialBlockHash)
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

	i.node.ListenEvents()

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

func New(context context.Context, cfg config.Config) Indexer {
	var remoteNode node.Node
	switch cfg.IndexerInfo().GetMode() {
	case node.RPC:
		remoteNode = cfg.RPCNode()
	case node.P2P:
		remoteNode = cfg.P2PNode()
	}

	return &indexer{
		context: context,
		storage: pg.NewStorage(cfg.DB()),
		log:     cfg.Log(),

		node: remoteNode,

		initialBlockHash: cfg.IndexerInfo().GetInitialBlockHash(),
		netParams:        cfg.IndexerInfo().GetNet(),

		start: make(chan struct{}),
	}
}
