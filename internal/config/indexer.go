package config

import (
	"errors"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/distraw/transaction-indexer/internal/core/node"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/kv"
)

const (
	indexerInfoKey = "indexer"
)

type IndexerInfo interface {
	GetInitialBlockHash() *chainhash.Hash
	GetNet() chaincfg.Params
	GetMode() node.Mode
}

type indexerInfo struct {
	InitialBlockHash *chainhash.Hash
	Net              chaincfg.Params
	Mode             node.Mode
}

func (i *indexerInfo) GetInitialBlockHash() *chainhash.Hash {
	return i.InitialBlockHash
}

func (i *indexerInfo) GetNet() chaincfg.Params {
	return i.Net
}

func (i *indexerInfo) GetMode() node.Mode {
	return i.Mode
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

func parseInitialHash(hash string) *chainhash.Hash {
	if hash == "genesis" || len(hash) == 0 {
		return &chainhash.Hash{}
	}

	parsedHash, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		panic(err)
	}

	return parsedHash
}

func parseMode(mode string) node.Mode {
	switch mode {
	case "rpc":
		return node.RPC
	case "p2p":
		return node.P2P
	}

	panic(errors.New("node mode was not recognized: " + mode))
}

func (c *config) IndexerInfo() IndexerInfo {
	return c.indexerInfo.Do(func() interface{} {
		var config struct {
			InitialBlockHash string `fig:"initial_block_hash,required"`
			Net              string `fig:"net,required"`
			Mode             string `fig:"mode,required"`
		}

		err := figure.Out(&config).
			From(kv.MustGetStringMap(c.getter, indexerInfoKey)).
			Please()
		if err != nil {
			panic(err)
		}

		indexerInfo := indexerInfo{
			InitialBlockHash: parseInitialHash(config.InitialBlockHash),
			Net:              parseNetParams(config.Net),
			Mode:             parseMode(config.Mode),
		}

		return &indexerInfo
	}).(IndexerInfo)
}
