package indexer

import (
	"context"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/distraw/transaction-indexer/internal/config"
	"github.com/distraw/transaction-indexer/internal/core/api/ctx"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/madflojo/tasks"
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
	Start = make(chan bool, 1)
)

func (i *indexer) Run() error {
	<-Start

	i.log.Infof("Starting from block %d", i.initialBlockHeight)

	// Provide routinePoll() with necessary dependencies through the context
	routinePollCtx := ctx.RPCProvider(i.rpc)(i.context)
	routinePollCtx = ctx.LoggerProvider(i.log)(routinePollCtx)
	routinePollCtx = ctx.StorageProvider(i.storage)(routinePollCtx)

	catchUp(routinePollCtx, i.initialBlockHeight)

	id, err := i.scheduler.Add(&tasks.Task{
		Interval: i.pollFrequency,
		TaskFunc: routinePoll(routinePollCtx),
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
