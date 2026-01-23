package indexer

import (
	"context"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/distraw/transaction-indexer/internal/core/api/ctx"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/madflojo/tasks"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
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
}

func routinePoll(c context.Context, initialHash *chainhash.Hash) func() error {
	storage := ctx.Storage(c).Blocks()
	rpc := ctx.RPC(c)
	log := ctx.Logger(c)

	prevHash := initialHash

	return func() error {
		hash, err := rpc.GetBestBlockHash()
		if err != nil {
			return errors.Wrap(err, "failed to poll best block hash")
		}
		if hash.IsEqual(prevHash) {
			return nil
		}

		info, err := rpc.GetBlockHeaderVerbose(hash)
		if err != nil {
			return errors.Wrap(err, "failed to poll best block header")
		}

		log.
			WithField("prev_hash", prevHash.String()).
			WithField("new_hash", hash.String()).
			Info("new hash detected")

		storage.Insert(data.Block{
			Hash:   hash.String(),
			Height: info.Height,
		})

		prevHash = (*chainhash.Hash)(hash.CloneBytes())
		return nil
	}
}

func (i *indexer) Run() error {
	initialHash, err := i.rpc.GetBestBlockHash()
	if err != nil {
		return errors.Wrap(err, "failed to poll initial best block hash")
	}
	i.log.Infof("Starting from block %s", initialHash.String())

	// Provide routinePoll() with necessary dependencies through the context
	routinePollCtx := ctx.RPCProvider(i.rpc)(i.context)
	routinePollCtx = ctx.LoggerProvider(i.log)(routinePollCtx)
	routinePollCtx = ctx.StorageProvider(i.storage)(routinePollCtx)

	id, err := i.scheduler.Add(&tasks.Task{
		Interval: time.Second * 10,
		TaskFunc: routinePoll(routinePollCtx, initialHash),
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

func New(context context.Context, storage data.Storage, log *logan.Entry, rpc *rpcclient.Client) Indexer {
	return &indexer{
		context: context,
		storage: storage.New(),
		log:     log,
		rpc:     rpc,

		scheduler: tasks.New(),
	}
}
