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
	GetInitialBlockHash() string
	GetNet() chaincfg.Params
}

type indexerInfo struct {
	PollFrequency    time.Duration
	InitialBlockHash string
	Net              chaincfg.Params
}

func (i *indexerInfo) GetPollFrequency() time.Duration {
	return i.PollFrequency
}

func (i *indexerInfo) GetInitialBlockHash() string {
	return i.InitialBlockHash
}

func (i *indexerInfo) GetNet() chaincfg.Params {
	return i.Net
}

func validateIndexerInfo(i indexerInfo) error {
	if i.PollFrequency < time.Second ||
		i.PollFrequency > time.Hour*24 {
		return errors.New("poll frequency must be between 1 second and 24 hours (86400 seconds)")
	}

	return nil
}

func (c *config) IndexerInfo() IndexerInfo {
	return c.indexerInfo.Do(func() interface{} {
		var config struct {
			PollFrequencySeconds int    `fig:"poll_frequency_seconds"`
			InitialBlockHash     string `fig:"initial_block_hash"`
			Net                  string `fig:"net"`
		}

		err := figure.Out(&config).
			From(kv.MustGetStringMap(c.getter, indexerInfoKey)).
			Please()
		if err != nil {
			panic(err)
		}

		indexerInfo := indexerInfo{
			PollFrequency:    time.Duration(config.PollFrequencySeconds) * time.Second,
			InitialBlockHash: config.InitialBlockHash,
		}

		if indexerInfo.InitialBlockHash == "genesis" {
			indexerInfo.InitialBlockHash = ""
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
