package config

import (
	"errors"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/kv"
)

const (
	indexerInfoKey = "indexer"
)

type IndexerInfo interface {
	GetPollFrequency() time.Duration
	GetInitialBlockHeight() int64
	GetNet() chaincfg.Params
}

type indexerInfo struct {
	PollFrequency      time.Duration
	InitialBlockHeight int64
	Net                chaincfg.Params
}

func (i *indexerInfo) GetPollFrequency() time.Duration {
	return i.PollFrequency
}

func (i *indexerInfo) GetInitialBlockHeight() int64 {
	return i.InitialBlockHeight
}

func (i *indexerInfo) GetNet() chaincfg.Params {
	return i.Net
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
			PollFrequencySeconds int    `fig:"poll_frequency_seconds"`
			InitialBlockHeight   int64  `fig:"initial_block_height"`
			Net                  string `fig:"net"`
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

		switch config.Net {
		case "mainnet":
			indexerInfo.Net = chaincfg.MainNetParams
		case "testnet3":
			indexerInfo.Net = chaincfg.TestNet3Params
		case "testnet4":
			indexerInfo.Net = chaincfg.TestNet4Params
		case "regtest":
			indexerInfo.Net = chaincfg.RegressionNetParams
		default:
			panic("invalid network params were providen: must be either mainnet, testnet3, testnet4 or regtest")
		}

		err = validateIndexerInfo(indexerInfo)
		if err != nil {
			panic(err)
		}

		return &indexerInfo
	}).(IndexerInfo)
}
