package config

import (
	"errors"
	"time"

	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/kv"
)

const (
	indexerInfoKey = "indexer"
)

type IndexerInfo struct {
	PollFrequency      time.Duration
	InitialBlockHeight int64
}

func (c *config) IndexerInfo() *IndexerInfo {
	return c.indexerInfo.Do(func() interface{} {
		var config struct {
			PollFrequencySeconds int   `fig:"poll_frequency_seconds"`
			InitialBlockHeight   int64 `fig:"initial_block_height"`
		}

		err := figure.Out(&config).
			From(kv.MustGetStringMap(c.getter, indexerInfoKey)).
			Please()
		if err != nil {
			panic(err)
		}

		switch {
		case config.PollFrequencySeconds < 1:
			panic(errors.New("Poll frequency must be [1; 864000]"))
		}

		indexerInfo := IndexerInfo{
			PollFrequency:      time.Duration(config.PollFrequencySeconds) * time.Second,
			InitialBlockHeight: config.InitialBlockHeight,
		}

		if indexerInfo.PollFrequency < time.Second ||
			indexerInfo.PollFrequency > time.Hour*24 {
			panic(errors.New("poll frequency must be between 1 second and 24 hours (86400 seconds)"))
		}

		if indexerInfo.InitialBlockHeight < 0 {
			panic(errors.New("initial block height must be greater than 0"))
		}

		return &indexerInfo
	}).(*IndexerInfo)
}
