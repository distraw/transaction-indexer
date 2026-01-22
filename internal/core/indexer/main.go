package indexer

import (
	"context"
	"time"

	"github.com/distraw/transaction-indexer/internal/core/indexer/poller"
	"github.com/madflojo/tasks"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type Indexer interface {
	Run() error
}

type indexer struct {
	log *logan.Entry

	context context.Context
	poller  poller.Poller

	scheduler *tasks.Scheduler
}

func (i *indexer) Run() error {
	prevHash, err := i.poller.Poll()
	if err != nil {
		return errors.Wrap(err, "failed to poll best block hash")
	}
	i.log.Infof("Starting from block %s", prevHash)

	id, err := i.scheduler.Add(&tasks.Task{
		Interval: time.Second * 10,
		TaskFunc: func() error {
			hash, err := i.poller.Poll()
			if err != nil {
				return errors.Wrap(err, "failed to poll best block hash")
			}

			if hash == prevHash {
				return nil
			}

			i.log.
				WithField("prev_hash", prevHash).
				WithField("new_hash", hash).
				Info("new hash detected")
			prevHash = hash
			return nil
		},
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

func New(context context.Context, log *logan.Entry, poller poller.Poller) Indexer {
	return &indexer{
		context: context,
		log:     log,
		poller:  poller,

		scheduler: tasks.New(),
	}
}
