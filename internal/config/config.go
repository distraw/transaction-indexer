package config

import (
	"github.com/distraw/transaction-indexer/internal/core/node"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/kit/pgdb"
)

type Config interface {
	comfig.Logger
	comfig.Listenerer
	pgdb.Databaser

	RPCNode() node.Node
	P2PNode() node.Node

	IndexerInfo() IndexerInfo
}

type config struct {
	comfig.Logger
	comfig.Listenerer
	pgdb.Databaser

	rpcNode comfig.Once
	p2pNode comfig.Once

	indexerInfo comfig.Once

	getter kv.Getter
}

func New(getter kv.Getter) Config {
	return &config{
		getter:     getter,
		Logger:     comfig.NewLogger(getter, comfig.LoggerOpts{}),
		Listenerer: comfig.NewListenerer(getter),
		Databaser:  pgdb.NewDatabaser(getter),
	}
}
