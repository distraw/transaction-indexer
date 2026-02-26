package config

import (
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/kv"
)

const (
	indexerInfoKey = "indexer"
)

type IndexerInfo interface {
	GetPollFrequency() time.Duration
	GetInitialBlockHash() *chainhash.Hash
	GetNet() chaincfg.Params
}

type indexerInfo struct {
	PollFrequency    time.Duration
	InitialBlockHash *chainhash.Hash
	Net              chaincfg.Params
}

func (i *indexerInfo) GetPollFrequency() time.Duration {
	return i.PollFrequency
}

func (i *indexerInfo) GetInitialBlockHash() *chainhash.Hash {
	return i.InitialBlockHash
}

func (i *indexerInfo) GetNet() chaincfg.Params {
	return i.Net
}

func parsePollFrequency(seconds int) time.Duration {
	const day int = 60 * 60 * 24

	if seconds < 1 || seconds > day {
		panic("poll frequency must be between 1 second and 24 hours (86400 seconds)")
	}

	return time.Duration(seconds) * time.Second
}

func parseNetParams(params string) chaincfg.Params {
	switch params {
	case "mainnet":
		return chaincfg.MainNetParams
	case "testnet3":
		return chaincfg.TestNet3Params
	case "testnet4":
		return chaincfg.TestNet4Params
	case "regtest":
		return chaincfg.RegressionNetParams
	}

	panic("invalid network params were providen: must be either mainnet, testnet3, testnet4 or regtest")
}

func parseInitialHash(hash string, params chaincfg.Params) *chainhash.Hash {
	if hash == "genesis" || len(hash) == 0 {
		return params.GenesisHash
	}

	parsedHash, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		panic(err)
	}

	return parsedHash
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
			PollFrequency:    parsePollFrequency(config.PollFrequencySeconds),
			Net:              parseNetParams(config.Net),
			InitialBlockHash: parseInitialHash(config.InitialBlockHash, parseNetParams(config.Net)),
		}

		return &indexerInfo
	}).(IndexerInfo)
}
