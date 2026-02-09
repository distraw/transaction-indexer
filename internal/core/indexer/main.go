package indexer

import (
	"context"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/distraw/transaction-indexer/internal/config"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/madflojo/tasks"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/logan/v3"
)

type Indexer interface {
	Run() error
}

type indexer struct {
	log *logan.Entry

	context   context.Context
	storage   data.Storage
	rpc       *rpcclient.Client
	scheduler *tasks.Scheduler

	pollFrequency      time.Duration
	initialBlockHeight int64
}

var (
	// TODO: delete or fix this
	Start     = make(chan bool, 1)
	Launched  = false
	CatchedUp = false
)

func (i *indexer) Run() error {
	<-Start

	i.log.Infof("Starting from block %d", i.initialBlockHeight)

	err := i.catchUp(i.initialBlockHeight)
	if err != nil {
		panic(err)
	}

	var hash *chainhash.Hash = nil
	block, err := i.storage.Blocks().GetHighest()
	if err == nil {
		hash, err = chainhash.NewHashFromStr(block.Hash)
		if err != nil {
			panic(err)
		}
	}
	if err != nil && !errors.Is(err, data.ErrNotFound) {
		panic(err)
	}

	id, err := i.scheduler.Add(&tasks.Task{
		Interval: i.pollFrequency,
		TaskFunc: i.routinePoll(hash),
		ErrFunc: func(err error) {
			i.log.WithError(err).Error("poller failed during routine poll")
		},
	})
	if err != nil {
		i.log.
			WithError(err).
			WithField("id", id).
			Error("scheduler failed unexpectedly")
	}

	Launched = true
	<-i.context.Done()
	i.scheduler.Stop()
	return nil
}

func New(context context.Context, storage data.Storage, log *logan.Entry,
	rpc *rpcclient.Client, info config.IndexerInfo) Indexer {
	return &indexer{
		context: context,
		storage: storage.New(),
		log:     log,
		rpc:     rpc,

		pollFrequency:      info.PollFrequency,
		initialBlockHeight: info.InitialBlockHeight,

		scheduler: tasks.New(),
	}
}
