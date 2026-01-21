package indexer

import (
	"context"
	"time"

	"github.com/distraw/transaction-indexer/internal/core/indexer/poller"
	"gitlab.com/distributed_lab/logan/v3"
)

type Indexer interface {
	Run()
}

type indexer struct {
	log *logan.Entry

	context context.Context
	poller  poller.Poller
}

func (i *indexer) Run() {
	var prevHash string

	for {
		select {
		case <-i.context.Done():
			return
		default:
		}

		time.Sleep(time.Second * 10)
		hash, err := i.poller.Poll()
		if err != nil {
			i.log.WithError(err).Error("failed to poll best block hash")
			continue
		}

		if hash != prevHash {
			i.log.
				WithField("prev_hash", prevHash).
				WithField("new_hash", hash).
				Info("new hash detected")
			prevHash = hash
		}
	}
}

func New(context context.Context, log *logan.Entry, poller poller.Poller) Indexer {
	return &indexer{
		context: context,
		log:     log,
		poller:  poller,
	}
}
