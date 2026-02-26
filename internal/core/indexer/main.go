package indexer

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/distraw/transaction-indexer/internal/config"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/madflojo/tasks"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/logan/v3"
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

	context   context.Context
	storage   data.Storage
	rpc       *rpcclient.Client
	scheduler *tasks.Scheduler

	start     chan struct{}
	started   atomic.Bool
	catchedUp atomic.Bool

	pollFrequency      time.Duration
	initialBlockHeight int64
	netParams          chaincfg.Params

	currentHash *chainhash.Hash
}

func (i *indexer) Run() error {
	select {
	case <-i.start:
	case <-i.context.Done():
		return i.context.Err()
	}

	i.started.Store(true)
	i.log.Infof("Starting from block %d", i.initialBlockHeight)

	err := i.catchUp(i.initialBlockHeight)
	if err != nil {
		return errors.Wrap(err, "failed to catch-up to initial block height")
	}

	block, err := i.storage.Blocks().GetHighest()
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

	defer i.scheduler.Stop()
	_, err = i.scheduler.Add(&tasks.Task{
		Interval: i.pollFrequency,
		TaskFunc: i.poll,
		ErrFunc: func(err error) {
			i.log.WithError(err).Error("poller failed during routine poll")
		},
	})
	if err != nil {
		return errors.Wrap(err, "scheduler failed unexpectedly")
	}

	<-i.context.Done()
	return nil
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

		pollFrequency:      info.GetPollFrequency(),
		initialBlockHeight: info.GetInitialBlockHeight(),
		netParams:          info.GetNet(),

		scheduler: tasks.New(),

		start: make(chan struct{}),
	}
}
