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

type IndexerInfo interface {
	GetPollFrequency() time.Duration
	GetInitialBlockHeight() int64
}

type indexerInfo struct {
	PollFrequency      time.Duration
	InitialBlockHeight int64
}

func (i *indexerInfo) GetPollFrequency() time.Duration {
	return i.PollFrequency
}

func (i *indexerInfo) GetInitialBlockHeight() int64 {
	return i.InitialBlockHeight
}

func validateIndexerInfo(i indexerInfo) error {
	if i.PollFrequency < time.Second ||
		i.PollFrequency > time.Hour*24 {
		return errors.New("poll frequency must be between 1 second and 24 hours (86400 seconds)")
	}

	if i.InitialBlockHeight < 0 {
		return errors.New("initial block height must be greater than 0")
	}

	return nil
}

func (c *config) IndexerInfo() IndexerInfo {
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

		indexerInfo := indexerInfo{
			PollFrequency:      time.Duration(config.PollFrequencySeconds) * time.Second,
			InitialBlockHeight: config.InitialBlockHeight,
		}

		err = validateIndexerInfo(indexerInfo)
		if err != nil {
			panic(err)
		}

		return &indexerInfo
	}).(IndexerInfo)
}
